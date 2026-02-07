package client

import (
	"context"
	"fmt"
	"time"

	message "chatui/internal/protocol"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/coder/websocket"
)

func connectCmd(cc *ChatClient, addr string) tea.Cmd {
	return func() tea.Msg {
		conn := cc.Connect(addr)
		if conn == nil {
			return errorMsg{err: fmt.Errorf("failed to connect to server")}
		}

		return connectedMsg{conn: conn, username: "User"}
	}
}

func setUsernameCmd(cc *ChatClient, conn *websocket.Conn, username string) tea.Cmd {
	return func() tea.Msg {
		cc.SetUsername(conn, username)
		return nil
	}
}

func listenCmd(cc *ChatClient, conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		msg, err := cc.ReceiveMessage(conn, context.Background())
		if err != nil {
			return errorMsg{err: err}
		}

		switch msg := msg.(type) {
		case message.ChatMessage:
			return receivedMsg{username: msg.Username, content: msg.Message, destination: msg.Destination}
		case message.LoginResponse:
			return loginMsg{success: msg.Success, message: msg.Message}
		case message.UserListUpdate:
			return userListMsg{users: msg.Users}

		default:
			return nil
		}
	}
}

func sendCmd(cc *ChatClient, conn *websocket.Conn, content string, destination string) tea.Cmd {
	return func() tea.Msg {
		cc.SendMessage(conn, content, destination)
		return nil
	}
}

func blinkCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg { return blinkMsg{} })
}
