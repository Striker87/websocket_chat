package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/Striker87/websocket_chat/chat"
)

const httpPort = "8080"

type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("templates", t.filename)))
	})
	err := t.templ.Execute(w, r)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	r := chat.NewRoom()

	http.Handle("/", &templateHandler{filename: "chat.html"})
	http.Handle("/room", r)

	// get the chat going
	go r.Run()

	fmt.Println("Server started at port:", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}
