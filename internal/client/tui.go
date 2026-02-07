package client

import (
	"context"
	"fmt"
	"log"
	"strings"

	message "chatui/internal/protocol"

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
	username    string
	destination string
	content     string
}
type loginMsg struct {
	success bool
	message string
}
type userListMsg struct {
	users []string
}

type errorMsg struct {
	err error
}

type ViewState int

const (
	ViewLogin ViewState = iota
	ViewChat
)

type FocusState int

const (
	FocusChat FocusState = iota
	FocusUserList
)

type model struct {
	// Login
	usernameInput textinput.Model
	loginHelper   string

	// UserList
	currentUsers     []string
	currentSelection int

	// Chat
	viewport         viewport.Model
	messages         map[string][]string
	qntNotifications map[string]int
	textarea         textarea.Model

	// Focus
	focusedArea FocusState

	// Shared
	chatClient  *ChatClient
	conn        *websocket.Conn
	username    string
	address     string
	currentView ViewState
	senderStyle lipgloss.Style
	err         error
	height      int
	width       int
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

	vp := viewport.New(0, 0)

	vp.SetContent("Welcome to the chat!\n")

	ta.KeyMap.InsertNewline.SetEnabled(false)

	ui := textinput.New()
	ui.Placeholder = "Username"
	ui.Focus()
	ui.CharLimit = 32
	ui.Width = 20

	return model{
		viewport:         vp,
		textarea:         ta,
		messages:         make(map[string][]string),
		err:              nil,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true),
		chatClient:       CreateChatClient(log.Printf),
		address:          addr,
		usernameInput:    ui,
		currentView:      ViewLogin,
		loginHelper:      "",
		currentUsers:     []string{"ALL"},
		currentSelection: 0,
		qntNotifications: make(map[string]int),
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
	case errorMsg:
		m.err = msg.err

	case tea.WindowSizeMsg:
		const decorationSpace = 4
		sidebarWidth := 25
		chatAreaWidth := msg.Width - sidebarWidth
		m.viewport.Width = chatAreaWidth - decorationSpace
		m.textarea.SetWidth(chatAreaWidth - decorationSpace)
		taHeight := m.textarea.Height() + decorationSpace
		m.viewport.Height = msg.Height - taHeight - decorationSpace

		m.viewport.GotoBottom()
		m.height = msg.Height
		m.width = msg.Width
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
	return lipgloss.JoinHorizontal(lipgloss.Top, m.renderSidebar(), m.renderChatArea())
}

func (m model) renderSidebar() string {
	var userList strings.Builder
	userList.WriteString("Active users\n\n")
	for i, user := range m.currentUsers {
		var line strings.Builder
		if i == m.currentSelection {
			fmt.Fprintf(&line, "» %s", user)
			if m.qntNotifications[user] > 0 {
				fmt.Fprintf(&line, " (%d)", m.qntNotifications[user])
			}
			fmt.Fprintf(&userList, "%s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Render(line.String()))
		} else {
			fmt.Fprintf(&line, "  %s", user)
			if m.qntNotifications[user] > 0 {
				fmt.Fprintf(&line, " (%d)", m.qntNotifications[user])
			}
			fmt.Fprintf(&userList, "%s\n", line.String())
		}
	}
	style := lipgloss.NewStyle().
		Width(20).
		Height(m.height - 2).
		Padding(1).
		Border(lipgloss.RoundedBorder())

	if m.focusedArea == FocusUserList {
		style = style.BorderForeground(lipgloss.Color("62"))
	} else {
		style = style.BorderForeground(lipgloss.Color("240"))
	}

	return style.Render(userList.String())
}

func (m model) renderChatArea() string {
	vpStyle := lipgloss.NewStyle().
		Width(m.viewport.Width + 2).
		Height(m.viewport.Height + 2).
		Border(lipgloss.RoundedBorder()).
		Padding(1)

	if m.focusedArea == FocusChat {
		vpStyle = vpStyle.BorderForeground(lipgloss.Color("179"))
	}

	taStyle := lipgloss.NewStyle().
		Width(m.viewport.Width + 2).
		Border(lipgloss.RoundedBorder()).
		Padding(1)

	if m.focusedArea == FocusChat {
		taStyle = taStyle.BorderForeground(lipgloss.Color("179"))
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		vpStyle.Render(m.viewport.View()),
		taStyle.Render(m.textarea.View()),
	)
}

func (m model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	var uiCmd tea.Cmd
	m.usernameInput, uiCmd = m.usernameInput.Update(msg)

	switch msg := msg.(type) {
	case loginMsg:
		if msg.success {
			m.currentView = ViewChat
			return m, listenCmd(m.chatClient, m.conn)
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
				listenCmd(m.chatClient, m.conn),
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

	switch msg := msg.(type) {
	case userListMsg:
		filteredUsers := []string{}
		for _, user := range msg.users {
			if user != m.username {
				filteredUsers = append(filteredUsers, user)
			}
		}

		m.currentUsers = append([]string{"ALL"}, filteredUsers...)
		return m, listenCmd(m.chatClient, m.conn)
	case receivedMsg:

		formattedMsg := fmt.Sprintf("%s: %s", m.senderStyle.Render(msg.username), msg.content)

		var chatTab string
		if msg.destination == "ALL" {
			chatTab = "ALL"
		} else if msg.username == m.username {
			chatTab = msg.destination
		} else {
			chatTab = msg.username
		}

		m.messages[chatTab] = append(m.messages[chatTab], formattedMsg)

		activeUser := m.currentUsers[m.currentSelection]
		if chatTab == activeUser {
			content := strings.Join(m.messages[activeUser], "\n")
			m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(content))
			m.viewport.GotoBottom()
		} else {
			m.qntNotifications[chatTab]++
		}
		m.viewport.GotoBottom()
		return m, listenCmd(m.chatClient, m.conn)

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:

			if m.focusedArea == FocusUserList {
				return m, nil
			} else {
				value := m.textarea.Value()
				m.textarea.Reset()
				m.viewport.GotoBottom()

				if value == "/quit" {
					m.chatClient.Disconnect(m.conn)
					return m, tea.Quit
				}

				destination := "ALL"

				if m.currentSelection > 0 && m.currentSelection < len(m.currentUsers) {
					destination = m.currentUsers[m.currentSelection]
				}

				return m, sendCmd(m.chatClient, m.conn, value, destination)
			}
		case tea.KeyTab:
			if m.focusedArea == FocusChat {
				m.focusedArea = FocusUserList
				m.textarea.Blur()
			} else {
				m.focusedArea = FocusChat
				m.textarea.Focus()
				activeUser := m.currentUsers[m.currentSelection]
				m.qntNotifications[activeUser] = 0
			}
			return m, nil
		case tea.KeyUp:
			if m.focusedArea == FocusUserList {
				if m.currentSelection > 0 {
					m.currentSelection--
					activeUser := m.currentUsers[m.currentSelection]
					m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages[activeUser], "\n")))
					m.viewport.GotoBottom()
				}
			}
		case tea.KeyDown:
			if m.focusedArea == FocusUserList {
				if m.currentSelection < len(m.currentUsers)-1 {
					m.currentSelection++

					activeUser := m.currentUsers[m.currentSelection]
					m.viewport.SetContent(lipgloss.NewStyle().Width(m.viewport.Width).Render(strings.Join(m.messages[activeUser], "\n")))
					m.viewport.GotoBottom()
				}
			}

		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	if m.focusedArea == FocusChat {
		m.textarea, tiCmd = m.textarea.Update(msg)
		m.viewport, vpCmd = m.viewport.Update(msg)
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

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
