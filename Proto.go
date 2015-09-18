package main

////////////////////////////////////////////////
// Data messages
////////////////////////////////////////////////

// Message to be sent from a client to the hub or viceversa
type Message struct {
	Content        string `json:"content"`
	ConversationID int    `json:"conversationID"`
	SenderID       int    `json:"senderID"`
	AuthToken      string `json:"authToken"`
}

// BcastMessage with recipients, used to send to the hub
type BcastMessage struct {
	Message
	RecipientIDs []int `json:"recipientIDs"`
}

////////////////////////////////////////////////
// Control messages
////////////////////////////////////////////////

// ControlOp enum
type ControlOp int

// ControlOp enums
const (
	OpConnect ControlOp = iota
	OpDisconnect
)

// CtrlMessage to the hub
type CtrlMessage struct {
	op   ControlOp
	conn *Connection
}
