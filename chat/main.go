package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Hub struct {
	clients   map[*Client]bool
	addClient chan *Client
	broadcast chan []byte
}

var hub = Hub{
	clients:   make(map[*Client]bool),
	addClient: make(chan *Client),
	broadcast: make(chan []byte),
}

func (hub *Hub) start() {
	fmt.Println("123")
	for {
		select {

		case conn := <-hub.addClient:
			hub.clients[conn] = true
			fmt.Println(hub.clients)

		case msg := <-hub.broadcast:
			for key := range hub.clients {
				key.ws.WriteMessage(1, msg)
			}
		}
	}
}

type Client struct {
	ws *websocket.Conn
	//room string
}

func main() {
	http.HandleFunc("/", home)
	go hub.start()
	http.HandleFunc("/ws", chat)
	http.ListenAndServe(":3000", nil)
}

func home(rw http.ResponseWriter, req *http.Request) {
	http.ServeFile(rw, req, "index.html")
}

func chat(rw http.ResponseWriter, req *http.Request) {
	conn, _ := upgrader.Upgrade(rw, req, nil)
	client := &Client{ws: conn}

	hub.addClient <- client

	go func(conn *websocket.Conn) {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				fmt.Println(err)
				break
			}
			hub.broadcast <- msg
		}
	}(conn)

}
