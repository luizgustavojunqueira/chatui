// Package message defines the structures for different types of messages in the chat application.
package message

import "encoding/json"

type MessageType string

const (
	TypeChatMessage    MessageType = "chat_message"
	TypeLoginResponse  MessageType = "login_response"
	TypeUserListUpdate MessageType = "user_list_update"
	TypeLoginRequest   MessageType = "login_request"
)

type Envelope struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data"`
}

type ChatMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type UserListUpdate struct {
	Users []string `json:"users"`
}

func MakeEnvelope(msgType MessageType, msg any) Envelope {
	return Envelope{
		Type: msgType,
		Data: func() []byte {
			data, _ := json.Marshal(msg)
			return data
		}(),
	}
}
