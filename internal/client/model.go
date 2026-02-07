package client

import (
	"log"

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
