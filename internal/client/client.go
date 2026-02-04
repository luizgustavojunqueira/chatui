// Package client
package client

import (
	"context"
	"errors"
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

func (cc ChatClient) Close(c *websocket.Conn) {
	err := c.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		cc.logf("websocket close error: %v", err)
		return
	}
}

func (cc ChatClient) SendMessage(c *websocket.Conn, msg string) {
	sendMsg := message.ChatMessage{
		Message: msg,
	}

	err := wsjson.Write(context.Background(), c, sendMsg)
	if err != nil {
		cc.logf("json data write error: %v", err)
		return
	}
}

func (cc ChatClient) ReceiveMessage(c *websocket.Conn, ctx context.Context) message.ChatMessage {
	var msg message.ChatMessage
	err := wsjson.Read(context.Background(), c, &msg)
	if err != nil {
		status := websocket.CloseStatus(err)

		if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
			cc.logf("websocket closed")
			return message.ChatMessage{
				Message: "Websocket connection closed.",
			}
		}

		if errors.Is(err, context.Canceled) {
			return message.ChatMessage{
				Message: "Context canceled.",
			}
		}

		cc.logf("error receiving message: %v", err)
		return message.ChatMessage{
			Message: "Error receiving message.",
		}
	}
	return msg
}
