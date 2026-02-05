package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
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
type loginMsg struct {
	success bool
	message string
}

type connectionErrorMsg struct {
	err error
}

type ViewState int

const (
	ViewLogin ViewState = iota
	ViewChat
)

type model struct {
	// Login
	usernameInput textinput.Model
	loginHelper   string

	// Chat
	viewport viewport.Model
	messages []string
	textarea textarea.Model

	// Shared
	chatClient  *ChatClient
	conn        *websocket.Conn
	username    string
	address     string
	currentView ViewState
	senderStyle lipgloss.Style
	err         error
}

type (
	errMsg error
)

const gap = "\n\n"

func InitialModel(addr string) model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 200
	ta.SetWidth(50)
	ta.SetHeight(4)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(50, 20)

	vp.SetContent("Welcome to the chat!\n")

	ta.KeyMap.InsertNewline.SetEnabled(false)

	ui := textinput.New()
	ui.Placeholder = "Username"
	ui.Focus()
	ui.CharLimit = 32
	ui.Width = 20

	return model{
		viewport:      vp,
		textarea:      ta,
		messages:      []string{},
		err:           nil,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		chatClient:    CreateChatClient(log.Printf),
		address:       addr,
		usernameInput: ui,
		currentView:   ViewLogin,
		loginHelper:   "",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		connectCmd(m.chatClient, m.address),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectedMsg:
		m.conn = msg.conn
	case connectionErrorMsg:
		m.err = msg.err
	}
	switch m.currentView {
	case ViewLogin:
		return m.updateLogin(msg)
	case ViewChat:
		return m.updateChat(msg)
	}
	return m, nil
}

func (m model) View() string {
	switch m.currentView {
	case ViewLogin:
		return m.viewLogin()
	case ViewChat:
		return m.viewChat()
	}
	return ""
}

func (m model) viewLogin() string {
	return fmt.Sprintf(
		"Enter your username:\n\n%s\n\n%s",
		m.usernameInput.View(),
		m.loginHelper,
	)
}

func (m model) viewChat() string {
	return fmt.Sprintf("%s%s%s", m.viewport.View(), gap, m.textarea.View())
}

func (m model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	var uiCmd tea.Cmd
	m.usernameInput, uiCmd = m.usernameInput.Update(msg)

	switch msg := msg.(type) {
	case loginMsg:
		if msg.success {
			m.currentView = ViewChat
			return m,
				listenCmd(m.chatClient, m.conn)
		} else {
			m.loginHelper = "Login failed: " + msg.message
			return m, nil
		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			username := m.usernameInput.Value()
			m.username = username

			return m, tea.Batch(
				setUsernameCmd(m.chatClient, m.conn, username),
				listenUsernameCmd(m.chatClient, m.conn),
			)
		default:
			m.loginHelper = ""
			return m, nil
		}
	}

	return m, uiCmd
}

func (m model) updateChat(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {

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

func connectCmd(cc *ChatClient, addr string) tea.Cmd {
	return func() tea.Msg {
		conn := cc.Connect(addr)
		if conn == nil {
			return connectionErrorMsg{err: fmt.Errorf("failed to connect to server")}
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

func listenUsernameCmd(cc *ChatClient, conn *websocket.Conn) tea.Cmd {
	return func() tea.Msg {
		msg := cc.ReceiveLoginResponse(conn, context.Background())
		return loginMsg{success: msg.Success, message: msg.Message}
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
