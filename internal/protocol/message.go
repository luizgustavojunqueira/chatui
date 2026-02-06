// Package message defines the structures for different types of messages in the chat application.
package message

import "encoding/json"

const (
	TypeChatMessage    = "chat_message"
	TypeLoginResponse  = "login_response"
	TypeUserListUpdate = "user_list_update"
	TypeLoginRequest   = "login_request"
)

type Envelope struct {
	Type string          `json:"type"`
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
