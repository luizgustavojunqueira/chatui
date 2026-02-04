// Package server
package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	message "chatui/internal/protocol"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type ConnectedClient struct {
	Conn     *websocket.Conn
	Username string
}

type Hub struct {
	clients    map[*ConnectedClient]bool
	broadcast  chan message.ChatMessage
	register   chan *ConnectedClient
	unregister chan *ConnectedClient
}

func CreateHub() Hub {
	return Hub{
		clients:    make(map[*ConnectedClient]bool),
		broadcast:  make(chan message.ChatMessage),
		register:   make(chan *ConnectedClient),
		unregister: make(chan *ConnectedClient),
	}
}

type ChatServer struct {
	logf func(f string, v ...any)
	hub  Hub
}

func CreateChatServer(logf func(f string, v ...any), hub Hub) *ChatServer {
	return &ChatServer{
		logf: logf,
		hub:  hub,
	}
}

func (cs ChatServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		cs.logf("websocket accept error: %v", err)
		return
	}
	defer c.CloseNow()

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour*24)
	defer cancel()

	userName := generateRandomUsername()
	client := &ConnectedClient{
		Conn:     c,
		Username: userName,
	}

	cs.hub.register <- client

	defer func() {
		cs.hub.unregister <- client
	}()

	for {
		var msg message.ChatMessage
		err = wsjson.Read(ctx, c, &msg)
		if err != nil {
			break
		}

		msg.Username = client.Username
		cs.hub.broadcast <- msg
	}

	c.Close(websocket.StatusNormalClosure, "")
}

func (hub Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = true
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				client.Conn.Close(websocket.StatusNormalClosure, "")
			}
		case msg := <-hub.broadcast:
			for client := range hub.clients {
				wsjson.Write(context.Background(), client.Conn, msg)
			}
		}
	}
}

func generateRandomUsername() string {
	randomNumber := time.Now().UnixNano() % 10000
	return fmt.Sprintf("User%04d", randomNumber)
}
