package client

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

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
