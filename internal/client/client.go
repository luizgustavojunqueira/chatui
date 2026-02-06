// Package client
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	message "chatui/internal/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type ChatClient struct {
	logf func(f string, v ...any)
}

func CreateChatClient(logf func(f string, v ...any)) *ChatClient {
	return &ChatClient{
		logf: logf,
	}
}

func (cc ChatClient) Connect(addr string) *websocket.Conn {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)

	defer cancel()

	c, _, err := websocket.Dial(ctx, "ws://"+addr+"/chat", nil)
	if err != nil {
		cc.logf("websocket dial error: %v", err)
		return nil
	}

	return c
}

func (cc ChatClient) Disconnect(c *websocket.Conn) {
	c.Close(websocket.StatusNormalClosure, "client disconnecting")
}

func (cc ChatClient) Close(c *websocket.Conn) {
	err := c.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		cc.logf("websocket close error: %v", err)
		return
	}
}

func (cc ChatClient) SendMessage(c *websocket.Conn, msg string, destination string) {
	sendMsg := message.ChatMessage{
		Destination: destination,
		Message:     msg,
	}

	envelope := message.MakeEnvelope(message.TypeChatMessage, sendMsg)

	err := wsjson.Write(context.Background(), c, envelope)
	if err != nil {
		cc.logf("json data write error: %v", err)
		return
	}
}

func (cc ChatClient) SetUsername(c *websocket.Conn, username string) {
	msg := message.LoginRequest{
		Username: username,
	}

	envelope := message.MakeEnvelope(message.TypeLoginRequest, msg)

	err := wsjson.Write(context.Background(), c, envelope)
	if err != nil {
		cc.logf("json data write error: %v", err)
		return
	}
}

func (cc ChatClient) ReceiveMessage(c *websocket.Conn, ctx context.Context) (any, error) {
	var envelope message.Envelope
	err := wsjson.Read(ctx, c, &envelope)
	if err != nil {
		return nil, err
	}

	switch envelope.Type {
	case message.TypeChatMessage:
		var msg message.ChatMessage
		json.Unmarshal(envelope.Data, &msg)
		return msg, nil
	case message.TypeLoginResponse:
		var msg message.LoginResponse
		json.Unmarshal(envelope.Data, &msg)
		return msg, nil
	case message.TypeUserListUpdate:
		var msg message.UserListUpdate
		json.Unmarshal(envelope.Data, &msg)
		return msg, nil
	default:
		return nil, fmt.Errorf("unknown message type: %s", envelope.Type)
	}
}
