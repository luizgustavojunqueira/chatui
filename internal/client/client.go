// Package client
package client

import (
	"context"
	"time"

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

func (cc ChatClient) SendMessage(c *websocket.Conn, message any) {
	err := wsjson.Write(context.Background(), c, message)
	if err != nil {
		cc.logf("json data write error: %v", err)
		return
	}
}
