package main

import (
	"errors"
	"log"
	"os"

	"chatui/internal/client"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) < 2 {
		return errors.New("please provide the server address as an argument")
	}

	serverAddr := os.Args[1]

	p := tea.NewProgram(client.InitialModel(serverAddr), tea.WithAltScreen())

	_, err := p.Run()

	return err
}
