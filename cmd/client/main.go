package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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

	defer cc.Close(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go cc.ReceiveMessage(c, ctx)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type messages and press Enter. Type /quit to exit.")

	for scanner.Scan() {
		msg := scanner.Text()

		if msg == "/quit" {
			break
		}

		if msg != "" {
			cc.SendMessage(c, msg)
		}
	}
	return nil
}
