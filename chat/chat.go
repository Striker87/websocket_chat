package chat

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

// client representsa singlechatting user
type client struct {
	socket  *websocket.Conn
	receive chan []byte //channel to receive messages from other clients
	room    *Room       // chat this client is chatting in
}

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  socketBufferSize,
	WriteBufferSize: socketBufferSize,
}

type Room struct {
	Clients map[*client]struct{} //bool // holds all current clients in this chat
	Join    chan *client         // channel for clients wishing to join the chat
	Leave   chan *client         // channel for clients wishing to leave the chat
	Forward chan []byte          // channel that holds incoming messages that should be forwarded to the other clients
}

// creating new chat chat
func NewRoom() *Room {
	return &Room{
		Forward: make(chan []byte),
		Join:    make(chan *client),
		Leave:   make(chan *client),
		Clients: make(map[*client]struct{}),
	}
}

func (r *Room) Run() {
	for {
		select {
		case client := <-r.Join:
			r.Clients[client] = struct{}{}
		case client := <-r.Leave:
			delete(r.Clients, client)
			close(client.receive)
		case msg := <-r.Forward:
			for client := range r.Clients {
				client.receive <- msg
			}
		}
	}
}

func (r *Room) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP err:", err)
	}
	client := &client{
		socket:  socket,
		receive: make(chan []byte, messageBufferSize),
		room:    r,
	}
	r.Join <- client
	defer func() {
		r.Leave <- client
	}()
	go client.write()
	client.read()
}

func (c *client) read() {
	defer c.socket.Close()

	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			log.Println("read() err:", err)
			return
		}
		c.room.Forward <- msg
	}
}

func (c *client) write() {
	defer c.socket.Close()

	for msg := range c.receive {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			log.Println("write() err:", err)
			return
		}
	}
}
