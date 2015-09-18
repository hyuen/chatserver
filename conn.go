package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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
	id      int
	receive chan Message // incoming used to receive messages from the hub
}

// Sender reads messages from the websocket and sends them to the hub
func (c *Connection) Sender() {
	defer func() {
		log.Info("deleting connection")
		MyHub.ctrl <- &CtrlMessage{op: OpDisconnect, conn: c}
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	//c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Error("error reading message: %v", err)
			break
		}
		numreqs++
		log.Debug("%d", numreqs)
		/*if numreqs > 100000 {
			pprof.StopCPUProfile()
			os.Exit(0)
		}*/
		var msg BcastMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Error("error parsing: %s", message)
			panic(err)
		}

		/*if !SessionValid(msg.SenderID, msg.AuthToken) {
			log.Info("invalid session, breaking")
			break
		}*/

		log.Debug("%s", msg)
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
			log.Debug("received %s", msgStr)
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

var numreqs int

// serverWs handles websocket requests from the peer.
//var serveWs = SessionRequired(
func serveWs(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	vars := mux.Vars(r)
	sUserID, ok := vars["user_id"]
	if !ok {
		http.Error(w, "Invalid url", 405)
		return
	}

	userID, err := strconv.Atoi(sUserID)
	if err != nil {
		http.Error(w, "Invalid url", 405)
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("%v", err)
		return
	}
	log.Info("creating connection for %d", userID)
	// Create Connection
	c := &Connection{ws: ws, id: userID, receive: make(chan Message, 256)}
	ctrlmsg := &CtrlMessage{op: OpConnect, conn: c}
	MyHub.ctrl <- ctrlmsg
	// Spawn goroutine for the receiver
	go c.Receiver()

	// sender
	c.Sender()
}
