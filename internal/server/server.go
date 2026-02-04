// Package server
package server

import (
	"context"
	"errors"
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

	for {

		var v any
		err = wsjson.Read(ctx, c, &v)
		if err != nil {
			status := websocket.CloseStatus(err)

			if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
				cs.logf("connection closed from client")
				break
			}

			if errors.Is(err, context.DeadlineExceeded) {
				cs.logf("read timeout")
				return
			}

			if status != -1 {
				cs.logf("connection closed with status %d: %v", status, err)
				return
			}

			cs.logf("json data read error: %v", err)
			break
		}

		cs.logf("Received: %v", v)
	}

	c.Close(websocket.StatusNormalClosure, "")
}
