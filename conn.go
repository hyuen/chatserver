package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Connection is an middleman between the websocket connection and the hub.
type Connection struct {
	ws      *websocket.Conn
	receive chan Message // incoming used to receive messages from the hub
}

// Sender reads messages from the websocket and sends them to the hub
func (c *Connection) Sender() {
	defer func() {
		MyHub.ctrl <- &CtrlMessage{op: OpDisconnect, conn: c}
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		var msg BcastMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("%s", message)
			panic(err)
		}

		log.Print(msg)
		MyHub.data <- msg
	}
}

// write writes a message with the given message type and payload.
func (c *Connection) write(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}

// Receiver pumps messages from the hub to the websocket connection.
func (c *Connection) Receiver() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.receive:
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			msgStr, _ := json.Marshal(message)
			if err := c.write(websocket.TextMessage, msgStr); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.write(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// Configuration description
type Configuration struct {
	SessionID   string
	RecipientID int
}

// serverWs handles websocket requests from the peer.
var serveWs = SessionRequired(
	func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", 405)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		/*
			var config Configuration
			err = ws.ReadJSON(&config)
			log.Print("config=", config)
		*/

		// Get User ID
		UserID := 12333
		// Create Connection
		c := &Connection{ws: ws, receive: make(chan Message, 256)}
		ctrlmsg := &CtrlMessage{op: OpConnect, id: UserID, conn: c}
		MyHub.ctrl <- ctrlmsg
		// Spawn goroutine for the receiver
		go c.Receiver()

		// sender
		c.Sender()
	},
)
