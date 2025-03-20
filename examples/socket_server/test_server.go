package main

import (
	"log"

	"github.com/code-serenade/go-kits/websocket"
)

func main() {
	server := websocket.NewWebSocketServer()
	err := server.Start(":13785")
	if err != nil {
		log.Fatal("WebSocket 服务器启动失败:", err)
	}
}
