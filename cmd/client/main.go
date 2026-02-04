package main

import (
	"errors"
	"log"
	"os"

	"chatui/internal/client"
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

	cc := client.CreateChatClient(log.Printf)

	c := cc.Connect(os.Args[1])

	cc.SendMessage(c, "Hello, World!")

	cc.Close(c)
	return nil
}
