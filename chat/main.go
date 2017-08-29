package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Hub struct {
	clients       map[*Client]bool
	addClient     chan *Client
	removeClient  chan *Client
	broadcast     chan []byte
	broadcastJson chan Smsg
}

var hub = Hub{
	clients:       make(map[*Client]bool),
	addClient:     make(chan *Client),
	removeClient:  make(chan *Client),
	broadcast:     make(chan []byte),
	broadcastJson: make(chan Smsg),
}

func (hub *Hub) start() {
	fmt.Println("Hub start!")

	db, sqlerr := sql.Open("mysql", "root:admin@/project")
	if sqlerr != nil {
		fmt.Println(sqlerr)
	}
	defer db.Close()

	strInsert, sqlerr := db.Prepare("insert into chat_context values(?,?,?)")
	if sqlerr != nil {
		fmt.Println(sqlerr)
	}
	defer strInsert.Close()

	for {
		select {

		case conn := <-hub.addClient:
			hub.clients[conn] = true
			fmt.Println(hub.clients)

		case msg := <-hub.broadcast:
			for key := range hub.clients {
				key.ws.WriteMessage(1, msg)
			}
		case conn := <-hub.removeClient:
			delete(hub.clients, conn)

		case val := <-hub.broadcastJson:
			_, sqlerr = strInsert.Exec(val.Room, val.Msg, time.Now())
			if sqlerr != nil {
				fmt.Println(sqlerr)
			}

			for key := range hub.clients {
				if key.room == val.Room {
					key.ws.WriteMessage(1, []byte(val.Msg))
				}
			}
		}
	}
}

type Client struct {
	ws   *websocket.Conn
	room string
}
type Smsg struct {
	Room string
	Msg  string
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
			var v Smsg
			err := conn.ReadJSON(&v)
			if err != nil {
				fmt.Println(err)
				break
			}
			client.room = v.Room
			hub.addClient <- client
			hub.broadcastJson <- v
		}

		defer func() {
			hub.removeClient <- client
			conn.Close()
		}()

	}(conn)

}
