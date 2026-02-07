package client

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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

type rawMessage struct {
	username string
	content  string
}

type model struct {
	// Login
	usernameInput textinput.Model
	loginHelper   string

	// UserList
	currentUsers     []string
	currentSelection int

	// Chat
	viewport         viewport.Model
	messages         map[string][]rawMessage
	qntNotifications map[string]int
	textarea         textarea.Model

	// Focus
	focusedArea FocusState
	blinkOn     bool

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
	errMsg   error
	blinkMsg struct{}
)

const gap = "\n\n"
const sidebarWidth = 26

func InitialModel(addr string) model {
	ta := textarea.New()
	ta.Placeholder = "Type your message... (/quit to exit)"
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 200
	ta.SetWidth(50)
	ta.SetHeight(4)

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	ta.ShowLineNumbers = false

	ta.EndOfBufferCharacter = ' '
	ta.FocusedStyle.EndOfBuffer = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	ta.BlurredStyle.EndOfBuffer = lipgloss.NewStyle().Background(lipgloss.Color("236"))

	ta.FocusedStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	ta.BlurredStyle.Base = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	ta.FocusedStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236"))
	ta.BlurredStyle.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236"))
	ta.FocusedStyle.Text = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	ta.BlurredStyle.Text = lipgloss.NewStyle().Background(lipgloss.Color("236"))
	ta.FocusedStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Background(lipgloss.Color("236"))
	ta.BlurredStyle.Prompt = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(lipgloss.Color("236"))

	vp := viewport.New(0, 0)
	vp.Style = lipgloss.NewStyle().Background(lipgloss.Color("234"))

	ta.KeyMap.InsertNewline.SetEnabled(false)

	ui := textinput.New()
	ui.Placeholder = "Username"
	ui.Focus()
	ui.CharLimit = 32
	ui.Width = 20

	bgColor := lipgloss.Color("234")
	emptyStyle := lipgloss.NewStyle().Background(bgColor)

	ui.PromptStyle = emptyStyle
	ui.TextStyle = emptyStyle
	ui.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Background(bgColor)
	ui.Cursor.Style = emptyStyle
	ui.Cursor.TextStyle = emptyStyle

	ui.Prompt = ""

	return model{
		viewport:         vp,
		textarea:         ta,
		messages:         make(map[string][]rawMessage),
		err:              nil,
		senderStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Background(lipgloss.Color("234")).Bold(true),
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
		blinkCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case connectedMsg:
		m.conn = msg.conn
	case errorMsg:
		m.err = msg.err
	case blinkMsg:
		m.blinkOn = !m.blinkOn
		return m, blinkCmd()

	case tea.WindowSizeMsg:
		chatAreaWidth := msg.Width - sidebarWidth
		taWidth := chatAreaWidth - 2
		m.viewport.Width = taWidth
		m.textarea.SetWidth(taWidth)
		taHeight := m.textarea.Height() + 2
		m.viewport.Height = msg.Height - taHeight

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
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Background(lipgloss.Color("234"))

	helperStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Background(lipgloss.Color("234")).
		Italic(true)

	inputStyled := ""
	if m.usernameInput.Value() == "" {
		inputStyled = lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("240")).
			Width(20).
			Render("Username")
	} else {
		inputStyled = lipgloss.NewStyle().
			Background(lipgloss.Color("234")).
			Foreground(lipgloss.Color("250")).
			Width(20).
			Render(m.usernameInput.Value())
	}

	leftPadding := (m.width - 30) / 2
	topPadding := (m.height - 8) / 2

	if leftPadding < 0 {
		leftPadding = 0
	}
	if topPadding < 0 {
		topPadding = 0
	}

	content := fmt.Sprintf(
		"%s\n\n> %s\n\n%s",
		titleStyle.Render("Enter your username:"),
		inputStyled,
		helperStyle.Render(m.loginHelper),
	)

	centered := lipgloss.NewStyle().
		PaddingTop(topPadding).
		PaddingLeft(leftPadding).
		Background(lipgloss.Color("234")).
		Render(content)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("234")).
		Width(m.width).
		Height(m.height).
		Render(centered)
}

func (m model) viewChat() string {
	chatContent := lipgloss.JoinHorizontal(lipgloss.Top, m.renderSidebar(), m.renderChatArea())

	fullScreenStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("234")).
		Width(m.width).
		Height(m.height)

	return fullScreenStyle.Render(chatContent)
}

func (m model) renderSidebar() string {
	var userList strings.Builder
	contentWidth := sidebarWidth - 2 // account for padding on the sidebar container

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Bold(true).
		Background(lipgloss.Color("236")).
		Padding(0, 1).
		Width(contentWidth).
		Align(lipgloss.Center)

	userList.WriteString(titleStyle.Render("Active users  (Tab: focus)") + "\n\n")

	for i, user := range m.currentUsers {
		var line strings.Builder
		if i == m.currentSelection {
			itemStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("62")).
				Bold(true).
				Padding(0, 1).
				Width(contentWidth)

			if m.focusedArea == FocusUserList && !m.blinkOn {
				itemStyle = itemStyle.Foreground(lipgloss.Color("62")).Background(lipgloss.Color("235"))
			}

			fmt.Fprintf(&line, "» %s", user)
			if m.qntNotifications[user] > 0 {
				fmt.Fprintf(&line, " (%d)", m.qntNotifications[user])
			}
			fmt.Fprintf(&userList, "%s\n", itemStyle.Render(line.String()))
		} else {
			itemStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("250")).
				Background(lipgloss.Color("235")).
				Padding(0, 1).
				Width(contentWidth)

			fmt.Fprintf(&line, "  %s", user)
			if m.qntNotifications[user] > 0 {
				notifStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("208")).
					Bold(true)
				fmt.Fprintf(&line, " %s", notifStyle.Render(fmt.Sprintf("(%d)", m.qntNotifications[user])))
			}
			fmt.Fprintf(&userList, "%s\n", itemStyle.Render(line.String()))
		}
	}

	content := userList.String()

	style := lipgloss.NewStyle().
		Width(sidebarWidth).
		Height(m.height).
		Background(lipgloss.Color("235")).
		Padding(1)

	return style.Render(content)
}

func (m model) renderMessages(user string) string {
	msgs := m.messages[user]
	contentStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("234")).
		Foreground(lipgloss.Color("252"))
	lineStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("234")).
		Width(m.viewport.Width)
	var rendered []string
	for _, raw := range msgs {
		styled := m.senderStyle.Render(raw.username+":") + contentStyle.Render(" "+raw.content)
		rendered = append(rendered, lineStyle.Render(styled))
	}
	return strings.Join(rendered, "\n")
}

func (m model) renderChatArea() string {
	chatWidth := m.width - sidebarWidth

	m.viewport.Style = lipgloss.NewStyle().Background(lipgloss.Color("234"))

	vpStyle := lipgloss.NewStyle().
		Width(chatWidth).
		Height(m.viewport.Height).
		Background(lipgloss.Color("234")).
		Padding(0, 1)

	taHeight := m.height - m.viewport.Height
	if taHeight < 0 {
		taHeight = 0
	}

	taWidth := chatWidth - 2

	bgSeq := "\x1b[48;5;236m"
	resetSeq := "\x1b[0m"

	m.textarea.Focus()
	taContent := m.textarea.View()
	taContent = bgSeq + strings.ReplaceAll(taContent, resetSeq, resetSeq+bgSeq) + resetSeq

	taLines := strings.Split(taContent, "\n")
	lineStyle := lipgloss.NewStyle().Background(lipgloss.Color("236")).Width(taWidth)
	for i, line := range taLines {
		taLines[i] = lineStyle.Render(line)
	}
	for len(taLines) < taHeight {
		taLines = append(taLines, lineStyle.Render(""))
	}

	filledTA := strings.Join(taLines, "\n")

	taStyle := lipgloss.NewStyle().
		Width(chatWidth).
		Height(taHeight).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	return lipgloss.JoinVertical(lipgloss.Left,
		vpStyle.Render(m.viewport.View()),
		taStyle.Render(filledTA),
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
			if m.conn != nil {
				m.chatClient.Disconnect(m.conn)
			}
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
			return m, uiCmd
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

		formattedMsg := rawMessage{username: msg.username, content: msg.content}

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
			m.viewport.SetContent(m.renderMessages(activeUser))
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
					if m.conn != nil {
						m.chatClient.Disconnect(m.conn)
					}
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
				return m, nil
			} else {
				m.focusedArea = FocusChat
				cmd := tea.Batch(m.textarea.Focus(), textarea.Blink)
				activeUser := m.currentUsers[m.currentSelection]
				m.qntNotifications[activeUser] = 0
				return m, cmd
			}
		case tea.KeyUp:
			if m.focusedArea == FocusUserList {
				if m.currentSelection > 0 {
					m.currentSelection--
					activeUser := m.currentUsers[m.currentSelection]
					m.viewport.SetContent(m.renderMessages(activeUser))
					m.viewport.GotoBottom()
				}
			}
		case tea.KeyDown:
			if m.focusedArea == FocusUserList {
				if m.currentSelection < len(m.currentUsers)-1 {
					m.currentSelection++

					activeUser := m.currentUsers[m.currentSelection]
					m.viewport.SetContent(m.renderMessages(activeUser))
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

func blinkCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*150, func(t time.Time) tea.Msg { return blinkMsg{} })
}
