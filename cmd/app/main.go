package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
)

// ====== 템플릿 관련 ======
var templates = template.Must(template.ParseGlob("./web/templates/*.html"))

func home(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "home.html", nil)
}

func listPage(w http.ResponseWriter, r *http.Request) {
	templates.ExecuteTemplate(w, "list.html", nil)
}

// ====== Peer 관련 구조 ======
type Peer struct {
	Key  string
	Name string
	Conn *websocket.Conn
}

type ChatMessage struct {
	Author  string `json:"Author"`
	Message string `json:"Message"`
	Type    int    `json:"Type"`
}

type Peers struct {
	V map[string]*Peer
	m sync.Mutex
}

var peers = Peers{
	V: make(map[string]*Peer),
}

// ====== WebSocket 업그레이드 ======
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

var peerKey = 1

// ====== 핸들러들 ======
func socketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}

	peer := &Peer{
		Key:  strconv.Itoa(peerKey),
		Conn: conn,
	}
	peers.m.Lock()
	peers.V[peer.Key] = peer
	peers.m.Unlock()

	peerKey++
	go read(peer)
}

func read(p *Peer) {
	defer func() {
		peers.m.Lock()
		delete(peers.V, p.Key)
		peers.m.Unlock()

		p.Conn.Close()

		if p.Name != "" {
			leave := ChatMessage{
				Author:  "admin",
				Message: fmt.Sprintf("%s님이 나갔습니다.", p.Name),
				Type:    0,
			}
			sendAll(leave)
		}
	}()

	for {
		var chat ChatMessage
		err := p.Conn.ReadJSON(&chat)
		if err != nil {
			log.Println("read error:", err)
			break
		}

		switch chat.Type {
		case 0: // 닉네임 등록
			if chat.Author == "" {
				errMsg := ChatMessage{
					Author:  "admin",
					Message: "닉네임은 필수입니다.",
					Type:    -1,
				}
				_ = p.Conn.WriteJSON(errMsg)
				return
			}

			// 중복 체크 먼저
			for _, other := range peers.V {
				if other.Name == chat.Author {
					errMsg := ChatMessage{
						Author:  "admin",
						Message: "중복되는 닉네임입니다.",
						Type:    -1,
					}
					_ = p.Conn.WriteJSON(errMsg)
					return // 현재 닉네임 등록 시도만 중단
				}
			}

			// 중복 아님 → 이름 등록
			p.Name = chat.Author

			// 닉네임 등록 성공 시 클라이언트에게 알림
			successMsg := ChatMessage{
				Author:  "admin",
				Message: "닉네임이 등록되었습니다.",
				Type:    2,
			}
			_ = p.Conn.WriteJSON(successMsg)

			// 참가 알림 전송
			join := ChatMessage{
				Author:  "admin",
				Message: fmt.Sprintf("%s님이 참가했습니다.", p.Name),
				Type:    0,
			}
			sendAll(join)

		case 1: // 일반 메시지
			if p.Name == "" {
				break // 닉네임 등록 안 된 상태
			}
			chat.Type = 1
			sendAll(chat)

		default:
			log.Println("정의되지 않은 Type:", chat.Type)
		}
	}
}

func sendAll(msg ChatMessage) {
	peers.m.Lock()
	defer peers.m.Unlock()

	for _, p := range peers.V {
		err := p.Conn.WriteJSON(msg)
		if err != nil {
			log.Println("send error:", err)
		}
	}
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	var names []string
	peers.m.Lock()
	for _, p := range peers.V {
		names = append(names, p.Name)
	}
	peers.m.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}

// ====== main() ======
func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/gohy", home)
	mux.HandleFunc("/gohy/list", listPage)

	mux.HandleFunc("/ws", socketHandler)
	mux.HandleFunc("/getUsers", getUsersHandler)

	fs := http.FileServer(http.Dir("./web/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Println("Listening on :4000")
	log.Fatal(http.ListenAndServe(":4000", mux))
}
