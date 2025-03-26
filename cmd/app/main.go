package main

import (
	"html/template"
	"log"
	"net/http"
)

// ParseGlob : 경로안에 있는거 다 잡을때 좋음음
var templates = template.Must(template.ParseGlob("./cmd/web/templates/*.html"))

// Home 핸들러
func home(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "home.html", nil)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
	}
}

// list 핸들러
func listPage(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "list.html", nil)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
	}
}

// api 핸들러
func clickApi(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "photo_api.html", nil)
	if err != nil {
		http.Error(w, "Failed to load template", http.StatusInternalServerError)
	}
}

func main() {
	mux := http.NewServeMux() //여러개 라우터 만들때 좋음 NewServeMux
	mux.HandleFunc("/gohy", home)
	mux.HandleFunc("/gohy/list", listPage)
	mux.HandleFunc("/gohy/api", clickApi)
	log.Println("Starting server on :4000")
	err := http.ListenAndServe(":4000", mux)
	log.Fatal(err)
}
