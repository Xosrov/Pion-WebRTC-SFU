package user

import (
	"encoding/json"
)

type MessageType string

// valid message types
const (
	MESSAGE_ICERESTART   MessageType = "icerestart"
	MESSAGE_PCFAILED     MessageType = "pcfailed"
	MESSAGE_STARTRTC     MessageType = "startrtc"
	MESSAGE_SDP          MessageType = "sdp"
	MESSAGE_ICECANDIDATE MessageType = "icecandidate"
)

// generic serializable message type for all communications
type Message struct {
	Type       MessageType     `json:"type"`
	RawPayload json.RawMessage `json:"payload"`
}

// 2-way message buffer structure
type messageBuffer struct {
	serverToClientMsgBuffer chan Message
	clientToServerMsgBuffer chan Message
}

// create new message buffer
func NewMessageBuffer() messageBuffer {
	return messageBuffer{
		make(chan Message),
		make(chan Message),
	}
}

// push to the server buffer of 2-way communication
func (mb *messageBuffer) PushToServerBuffer(message Message) {
	mb.clientToServerMsgBuffer <- message
}

// push to the client buffer of 2-way communication
func (mb *messageBuffer) PushToClientBuffer(message Message) {
	mb.serverToClientMsgBuffer <- message
}

// read from server buffer of 2-way communication
func (mb *messageBuffer) ReadFromServerBuffer() <-chan Message {
	return mb.clientToServerMsgBuffer
}

// read from client buffer of 2-way communication
func (mb *messageBuffer) ReadFromClientBuffer() <-chan Message {
	return mb.serverToClientMsgBuffer
}
