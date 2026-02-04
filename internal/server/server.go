// Package server
package server

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type ChatServer struct {
	logf func(f string, v ...any)
}

func CreateChatServer(logf func(f string, v ...any)) *ChatServer {
	return &ChatServer{
		logf: logf,
	}
}

func (cs ChatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		cs.logf("websocket accept error: %v", err)
		return
	}
	defer c.CloseNow()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var v any
	err = wsjson.Read(ctx, c, &v)
	if err != nil {
		cs.logf("json data read error: %v", err)
		return
	}

	cs.logf("Received: %v", v)

	c.Close(websocket.StatusNormalClosure, "")
}
