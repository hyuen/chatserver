package main

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	connections map[int]*Connection

	// Inbound messages from the connections.
	data chan BcastMessage
	ctrl chan *CtrlMessage
}

// Myhub is the global variable for a hub
var MyHub = hub{
	// state
	connections: make(map[int]*Connection),

	// incoming channels
	data: make(chan BcastMessage, 256),
	ctrl: make(chan *CtrlMessage),
}

func (h *hub) run() {
	for {
		select {
		case m := <-h.ctrl:
			switch m.op {
			case OpConnect:
				h.connections[m.conn.id] = m.conn
			case OpDisconnect:
				log.Info("Deleting connection %d", m.conn.id)
				if _, ok := h.connections[m.conn.id]; ok {
					log.Info("Found and deleting connection %d", m.conn.id)
					delete(h.connections, m.conn.id)
					close(m.conn.receive)
				}
			}
		case m := <-h.data:
			for _, RecipientID := range m.RecipientIDs {
				// Find a connection and place the message in the queue
				/*if RecipientID == m.SenderID {
					continue
				}*/
				if conn, ok := h.connections[RecipientID]; ok {
					dstmsg := Message{Content: m.Content,
						ConversationID: m.ConversationID,
						SenderID:       m.SenderID,
						AuthToken:      m.AuthToken}
					conn.receive <- dstmsg
				} else {
					// save for later
					log.Debug("dropping packet for %d", RecipientID)
				}
			}
		}
	}
}
