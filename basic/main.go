package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

func main() {
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func home(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "home.html")
}

func echo(rw http.ResponseWriter, req *http.Request) {
	conn, _ := upgrader.Upgrade(rw, req, nil)
	defer conn.Close()
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		err = conn.WriteMessage(messageType, append([]byte("hello"), message[:]...))
		if err != nil {
			break
		}
	}
}
