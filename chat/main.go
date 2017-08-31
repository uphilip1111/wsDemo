package main

import (
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type Client struct {
	ws *websocket.Conn
}
type ClientSendMsg struct {
	Room   string
	Msg    string
	Status int
}

type ClientToRoom struct {
	room   string
	client *Client
}
type Hub struct {
	clientList      map[string][]*Client
	addClient       chan *Client
	addClientToRoom chan ClientToRoom
	removeClient    chan *Client
	broadcastJson   chan ClientSendMsg
}

var hub = Hub{
	clientList:      make(map[string][]*Client),
	addClient:       make(chan *Client),
	addClientToRoom: make(chan ClientToRoom),
	removeClient:    make(chan *Client),
	broadcastJson:   make(chan ClientSendMsg),
}

func (hub *Hub) start() {
	fmt.Println("Hub start!")
	//SQL CONNECTION
	/*db, sqlerr := sql.Open("mysql", "root:admin@/project")
	if sqlerr != nil {
		fmt.Println(sqlerr)
	}
	defer db.Close()

	strInsert, sqlerr := db.Prepare("insert into chat_context values(?,?,?)")
	if sqlerr != nil {
		fmt.Println(sqlerr)
	}
	defer strInsert.Close()*/

	for {
		select {

		case roomConn := <-hub.addClientToRoom:
			hub.clientList[roomConn.room] = append(hub.clientList[roomConn.room], roomConn.client)
			fmt.Println("client List:", hub.clientList)

		case conn := <-hub.addClient:
			hub.clientList["Lobby"] = append(hub.clientList["Lobby"], conn)

		case conn := <-hub.removeClient:
			for room, clientlist := range hub.clientList {
				for clientlistidx, client := range clientlist {
					if client == conn {
						hub.clientList[room] = append(hub.clientList[room][:clientlistidx], hub.clientList[room][clientlistidx+1:]...)
					}
				}
			}

			for _, client := range hub.clientList["Lobby"] {
				client.ws.WriteMessage(1, []byte("玩家離開房間"))
			}

		case val := <-hub.broadcastJson:
			//SQL EXECUTE
			/*_, sqlerr = strInsert.Exec(val.Room, val.Msg, time.Now())
			if sqlerr != nil {
				fmt.Println(sqlerr)
			}*/

			for _, v := range hub.clientList[val.Room] {
				v.ws.WriteMessage(1, []byte(val.Msg))
			}
		}
	}
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
		var clientToRoom ClientToRoom
		for {
			var msgFromClient ClientSendMsg
			err := conn.ReadJSON(&msgFromClient)
			if err != nil {
				fmt.Println(err)
				break
			}

			if msgFromClient.Status == 1 {
				clientToRoom.room = msgFromClient.Room
				clientToRoom.client = client
				hub.addClientToRoom <- clientToRoom
			}
			hub.broadcastJson <- msgFromClient
		}

		defer func() {
			hub.removeClient <- client
			conn.Close()
		}()

	}(conn)

}
