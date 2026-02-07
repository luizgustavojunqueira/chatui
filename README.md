# Chatui

Chatui is a terminal chat client/server built in Go using websockets and a Bubble Tea TUI. It includes a server hub for routing messages and a client with a solid, focused UI that supports multiple chats, unread counts, and keyboard-driven navigation.

## Demo Video

<video controls>
  <source src="https://luizgustavojunqueira-personalblog.s3.us-east-1.amazonaws.com/Chatui/screenrecording-2026-02-06_23-08-02.mp4?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Content-Sha256=UNSIGNED-PAYLOAD&X-Amz-Credential=ASIASDJ36U2TCK2GJUIA%2F20260207%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20260207T031253Z&X-Amz-Expires=300&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEIz%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEaCXVzLWVhc3QtMSJHMEUCIGlzRRhaCrzi%2BTa2RCHWSANFWaT3z5AF9ULFe3SrTs%2F%2FAiEAqLAdj2fR9tmtfV1L27PgSmVdrt09lNy2ALfChXlHZlcq2gIIVBABGgwxNDQ1NDQwMTYwMzgiDJNTkLJfWmHwqAUvsiq3AmxGAmxW39QYoYCaskp%2FOz78T3aTCn62xGbskhmhkJ3dJbu%2FmlBIEq5RxOb5TRGJDGTiZ%2Fp79XS21J82S78yaDmzMyMbRQGP5o6v8Kan1ueJrG3ChqQbHgH%2BVSZIiX7pdj7rtQwFmWwnuVZUrZo6C2MwzVeEp4BRK3hfDztuiNwCKysHmvOFPc8mgtkIwBPRS1iN9rUWpw7uQogFPvFe2xJtzCAMvi2j91jHgb1zey5vy6uwelDxR3YXWZKoDSk4OgIIzJUR8iIItv7IpGPc%2FY8kXxGw9VKqR4McEhUpqqoDV%2BV7o321zgNyuQbMLW%2FyJn%2BvKCvvkySHJvrcAH16xUYHarzD6Wp8rm9AyyFa5nlXbLY6AtYOONS56Jg4mLI7faeA%2FMEOjbxxmfmKhjFZJLuC9qfc%2Bok7MIrWmswGOq0Cu%2BRxttvuFVn1%2BVM3KmKX1rwWon18UUo3xdl%2FpdbBqMFjWJgdsReYSKFOvpkEwFAkpsccbDIdN73EUPfP7bLVi3l66u3zHE6UBOVcQz5MlDl5NTcdVr%2FwZ2goMDvWfNJJASSs%2F2aRnHrtpuRmveFIOeQtgdmxqEekwRRdQDAfvYAwWMKVqUwrgsBD76PSWLvS1SmwTlS3ycpwkFa9H%2ByRMSeuqZgtgTFv4aTa%2BmWsgsq9cKSTfONAltGiGuawFSPsSI4HIt7FYNG90OUs0GjskFGPiScQEGODeWlOumG8pXLvSvr44F0WX4fBSKf3g6CITDbVBAWV110dm%2FgczcK3ogEGHtEwwcku5CGR8XVjJJyr3Vq3P3ieOonSVo8Thi1X20Z13t9%2Fr07TH5oFvg%3D%3D&X-Amz-Signature=1f2b173be6216002d464a796206d842eefb90ea155e4d02a5c367819fca8c363&X-Amz-SignedHeaders=host&response-content-disposition=inline" type="video/mp4">
  Seu navegador não suporta o elemento de vídeo.
</video>

## Blog Post

Read more about the build process, UI tweaks, and lessons learned: [Chatui blog post](blog.luizgustavojunqueira.com/blog/chatui/)

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
