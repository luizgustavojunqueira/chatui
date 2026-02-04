package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/coder/websocket"
)

type connectedMsg struct {
	conn     *websocket.Conn
	username string
}

type receivedMsg struct {
	username string
	content  string
}

type connectionErrorMsg struct {
	err error
}

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	chatClient  *ChatClient
	conn        *websocket.Conn
	username    string
	address     string
}

type (
	errMsg error
)

const gap = "\n\n"

func InitialModel(addr string) model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()

	ta.Prompt = ">>> "
	ta.CharLimit = 200
	ta.SetWidth(50)
	ta.SetHeight(1)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(50, 20)

	vp.SetContent("Welcome to the chat!\n")

	ta.KeyMap.InsertNewline.SetEnabled(false)

	return model{
		viewport:    vp,
		textarea:    ta,
		messages:    []string{},
		err:         nil,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		chatClient:  CreateChatClient(log.Printf),
		address:     addr,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		connectCmd(m.chatClient, m.address),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {

	case connectedMsg:
		m.conn = msg.conn
		m.messages = append(m.messages, "Connected")
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.textarea.Reset()
		m.viewport.GotoBottom()
		return m, listenCmd(m.chatClient, m.conn)
	case receivedMsg:
		m.messages = append(m.messages, m.senderStyle.Render(msg.username+": ")+msg.content)
		m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		m.viewport.GotoBottom()
		return m, listenCmd(m.chatClient, m.conn)
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		m.viewport.Height = msg.Height - m.textarea.Height() - lipgloss.Height(gap)

		if len(m.messages) > 0 {
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages, "\n")))
		}

		m.viewport.GotoBottom()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case tea.KeyEnter:
			value := m.textarea.Value()
			m.textarea.Reset()
			m.viewport.GotoBottom()

			return m, sendCmd(m.chatClient, m.conn, value)
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
}

func connectCmd(cc *ChatClient, addr string) tea.Cmd {
	return func() tea.Msg {
		conn := cc.Connect(addr)
		if conn == nil {
			return connectionErrorMsg{err: fmt.Errorf("failed to connect to server")}
		}

		return connectedMsg{conn: conn, username: "User"}
	}
}

func listenCmd(cc *ChatClient, conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		msg := cc.ReceiveMessage(conn, context.Background())
		return receivedMsg{username: msg.Username, content: msg.Message}
	}
}

func sendCmd(cc *ChatClient, conn *websocket.Conn, content string) tea.Cmd {
	return func() tea.Msg {
		cc.SendMessage(conn, content)
		return nil
	}
}
