# Chatui

Chatui is a terminal chat client/server built in Go using websockets and a Bubble Tea TUI. It includes a server hub for routing messages and a client with a solid, focused UI that supports multiple chats, unread counts, and keyboard-driven navigation.

## Demo Video

<video controls>
  <source src="https://luizgustavojunqueira-personalblog.s3.us-east-1.amazonaws.com/Chatui/screenrecording-2026-02-06_23-08-02.mp4" type="video/mp4">
  Seu navegador não suporta o elemento de vídeo.
</video>

## Blog Post

Read more about the build process, UI tweaks, and lessons learned: [Chatui blog post](https://blog.luizgustavojunqueira.com/blog/chatui/)

## Features

- WebSocket server with hub for broadcasting and direct messages
- Bubble Tea TUI client with solid backgrounds and focus cues
- Unread message counters per chat and auto-clear on focus
- Login screen and chat switching (ALL + private chats)
- Keyboard shortcuts: Tab toggles focus (sidebar/chat), Enter sends, `/quit` exits

## Project Layout

```
cmd/
  server/main.go   # starts the websocket server
  client/main.go   # starts the TUI client
internal/
  server/          # hub, client registration, routing
  client/          # TUI client (model, view, update, commands)
protocol/          # message envelope/types
```

## Prerequisites

- Go 1.21+ (recommended)

## Running

In separate terminals:

1) Start the server
```sh
go run ./cmd/server <address>
```

2) Start the client
```sh
go run ./cmd/client <address>
```

## Development

- Install deps (if any) via `go mod tidy`
- TUI code is split by responsibility: `model.go`, `view.go`, `update.go`, `cmd.go`
