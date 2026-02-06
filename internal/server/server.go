// Package server
package server

import (
	"context"
	"encoding/json"
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

type usernameCheck struct {
	username string
	response chan bool
}

type Hub struct {
	clients       map[*ConnectedClient]bool
	broadcast     chan message.ChatMessage
	register      chan *ConnectedClient
	unregister    chan *ConnectedClient
	checkUsername chan usernameCheck
}

func CreateHub() Hub {
	return Hub{
		clients:       make(map[*ConnectedClient]bool),
		broadcast:     make(chan message.ChatMessage),
		register:      make(chan *ConnectedClient),
		unregister:    make(chan *ConnectedClient),
		checkUsername: make(chan usernameCheck),
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

	client := &ConnectedClient{
		Conn: c,
	}

	if !cs.handleUsernameRegistration(ctx, client) {
		return
	}

	cs.hub.register <- client

	defer func() {
		cs.hub.unregister <- client
	}()

	for {
		var env message.Envelope
		err := wsjson.Read(ctx, c, &env)
		if err != nil {
			break
		}
		if env.Type != message.TypeChatMessage {
			continue
		}
		var msg message.ChatMessage

		json.Unmarshal(env.Data, &msg)

		msg.Username = client.Username
		cs.hub.broadcast <- msg
	}

	c.Close(websocket.StatusNormalClosure, "")
}

func (cs ChatServer) handleUsernameRegistration(ctx context.Context, client *ConnectedClient) bool {
	for {

		var envelope message.Envelope

		err := wsjson.Read(ctx, client.Conn, &envelope)
		if err != nil {
			cs.logf("error reading envelope: %v", err)
			return false
		}

		if envelope.Type != message.TypeLoginRequest {
			resp := message.MakeEnvelope(message.TypeLoginResponse, message.LoginResponse{
				Success: false,
				Message: "Expected login request",
			})
			wsjson.Write(ctx, client.Conn, resp)
			continue
		}

		loginReq := message.LoginRequest{}

		json.Unmarshal(envelope.Data, &loginReq)

		if cs.hub.isUsernameTaken(loginReq.Username) {
			resp := message.MakeEnvelope(message.TypeLoginResponse, message.LoginResponse{
				Success: false,
				Message: "Username is already taken",
			})
			wsjson.Write(ctx, client.Conn, resp)

			continue

		}

		client.Username = loginReq.Username
		resp := message.MakeEnvelope(message.TypeLoginResponse, message.LoginResponse{
			Success: true,
			Message: "Login successful",
		})
		wsjson.Write(ctx, client.Conn, resp)
		return true
	}
}

func (hub Hub) Run() {
	for {
		select {
		case client := <-hub.register:
			hub.clients[client] = true
			userList := message.UserListUpdate{
				Users: make([]string, 0, len(hub.clients)),
			}

			for client := range hub.clients {
				userList.Users = append(userList.Users, client.Username)
			}

			envelope := message.MakeEnvelope(message.TypeUserListUpdate, userList)
			for client := range hub.clients {
				wsjson.Write(context.Background(), client.Conn, envelope)
			}
		case client := <-hub.unregister:
			if _, ok := hub.clients[client]; ok {
				delete(hub.clients, client)
				client.Conn.Close(websocket.StatusNormalClosure, "")
			}
		case msg := <-hub.broadcast:
			envelope := message.MakeEnvelope(message.TypeChatMessage, msg)

			fmt.Println("Broadcasting message to", msg.Destination)

			if msg.Destination == "ALL" {
				for client := range hub.clients {
					wsjson.Write(context.Background(), client.Conn, envelope)
				}
			} else {
				for client := range hub.clients {
					if client.Username == msg.Destination || client.Username == msg.Username {
						wsjson.Write(context.Background(), client.Conn, envelope)
					}
				}
			}
		case check := <-hub.checkUsername:
			taken := false
			for client := range hub.clients {
				if client.Username == check.username {
					taken = true
				}
			}
			check.response <- taken
		}
	}
}

func (hub Hub) isUsernameTaken(username string) bool {
	responseChan := make(chan bool)
	hub.checkUsername <- usernameCheck{
		username: username,
		response: responseChan,
	}
	return <-responseChan
}
