package client

import (
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

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

func (m model) updateLogin(msg tea.Msg) (tea.Model, tea.Cmd) {
	var uiCmd tea.Cmd
	m.usernameInput, uiCmd = m.usernameInput.Update(msg)

	switch msg := msg.(type) {
	case loginMsg:
		if msg.success {
			m.currentView = ViewChat
			return m, listenCmd(m.chatClient, m.conn)
		}
		m.loginHelper = "Login failed: " + msg.message
		return m, nil
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
			}

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
		case tea.KeyTab:
			if m.focusedArea == FocusChat {
				m.focusedArea = FocusUserList
				m.textarea.Blur()
				return m, nil
			}
			m.focusedArea = FocusChat
			cmd := tea.Batch(m.textarea.Focus(), textarea.Blink)
			activeUser := m.currentUsers[m.currentSelection]
			m.qntNotifications[activeUser] = 0
			return m, cmd
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
