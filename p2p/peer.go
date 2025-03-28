package p2p

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/hay000000/golang.git/utils"
)

// peers struct 정의 V에 map으로 Peer struct 저장
type peers struct {
	V map[string]*Peer
	m sync.Mutex // Mutex를 넣어야 unlock/lock 가능
}

// Peer struct 차례대로 : 구분값, 이름, websocket 연결 객체를 저장
type Peer struct {
	Key  string
	Name string
	Conn *websocket.Conn
}

// ChatMessage struct : 메시지의 struct
type ChatMessage struct {
	Author  string `json:"Author"`
	Message string `json:"Message"`
	Type    string `json:"Type"`
}

// Peers peers 로 변수 생성
var Peers = peers{
	V: make(map[string]*Peer),
}

// Peer Pointer receiver Read 메서드 정의
func (peer *Peer) Read() {
	// Peer 종료시 close 함수 실행
	defer peer.close()

	for {
		var chat ChatMessage
		var byteChat []byte
		var isDuplication bool

		// ReadMessage 는 새로운 메시지가 올 때 까지 기다림, 메세지가 오면 payload 에 저장
		messageType, payload, err := peer.Conn.ReadMessage()

		// conn 에서 에러가 날 경우 빠른 구분을 위해 아래 처럼 코딩, utils.HandleErr(err)로 변경 가능
		if err != nil {
			log.Printf("conn.ReadMessage: %v", err)
			return
		}

		// 중복되는 name 을 가진지 체크
		isDuplication = PeerNameDuplicationCheck(peer)

		// payload를 메시지 형태로 변경
		utils.HandleErr(json.Unmarshal(payload, &chat))
		// Type 이 1은 일반 메시지로, 0이면 관리자 메세지로 정의
		chat.Type = strconv.Itoa(1)

		// Peer 에 이름 추가
		if !isDuplication {
			peer.Name = chat.Author
		}

		// 메세지 전체 전송
		byteChat = utils.StructToBytes(chat)
		SendMessageToPeers(messageType, byteChat)
	}
}

// close : Peer 정리 및 퇴장 메시지 전송 함수
func (p *Peer) close() {
	// data race 보호를 위한 코드 추가
	Peers.m.Lock()
	defer func() {
		Peers.m.Unlock()
	}()
	p.Conn.Close()

	// 해당 Peer 를 Peers 에서 삭제
	delete(Peers.V, p.Key)

	// 퇴장 메시지 생성 및 전송
	var leaveChat ChatMessage

	leaveChat.Author = "admin"
	leaveChat.Message = fmt.Sprintf("%s님이 나갔습니다.", p.Name)
	leaveChat.Type = strconv.Itoa(0)

	byteChat := utils.StructToBytes(leaveChat)
	SendMessageToPeers(websocket.TextMessage, byteChat)
}

// PeerNameDuplicationCheck : 파라미터로 온 값이 Peers 에 존재 여부 반환
func PeerNameDuplicationCheck(peer *Peer) bool {
	for _, p := range Peers.V {
		if p.Name != "" && p.Name == peer.Name {
			return true
		}
	}
	return false
}

// SendMessageToPeers : Peers 의 전체 peer 에 메세지 전달
func SendMessageToPeers(messageType int, byteChat []byte) {
	for _, p := range Peers.V {
		if err := p.Conn.WriteMessage(messageType, byteChat); err != nil {
			log.Printf("conn.WriteMessage: %v", err)
			continue
		}
	}
}
