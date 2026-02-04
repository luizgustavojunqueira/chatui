// Package message defines the structures for different types of messages in the chat application.
package message

type ChatMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}
