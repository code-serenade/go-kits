package websocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocketServer 结构体
type WebSocketServer struct {
	clients   map[*websocket.Conn]bool
	broadcast chan []byte
	mutex     sync.Mutex
	upgrader  websocket.Upgrader
}

// NewWebSocketServer 创建 WebSocket 服务器实例
func NewWebSocketServer() *WebSocketServer {
	return &WebSocketServer{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // 允许所有来源
		},
	}
}

// HandleConnections 处理 WebSocket 连接
func (server *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 连接为 WebSocket 连接
	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}

	// 添加到客户端列表
	server.mutex.Lock()
	server.clients[conn] = true
	server.mutex.Unlock()

	log.Println("新客户端连接:", conn.RemoteAddr())

	defer func() {
		server.mutex.Lock()
		delete(server.clients, conn)
		server.mutex.Unlock()
		conn.Close()
		log.Println("客户端断开连接:", conn.RemoteAddr())
	}()

	// 监听消息
	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("读取消息失败:", err)
			break
		}

		// 处理 Ping/Pong
		if msgType == websocket.PingMessage {
			log.Println("收到 Ping，回复 Pong")
			conn.WriteMessage(websocket.PongMessage, nil)
			continue
		}

		log.Printf("收到消息: %s", msg)

		// 发送消息到广播通道
		server.broadcast <- msg
	}
}

// Start 启动 WebSocket 服务器
func (server *WebSocketServer) Start(addr string) error {
	http.HandleFunc("/ws", server.HandleConnections)

	// 启动消息广播
	go server.handleMessages()

	log.Printf("WebSocket 服务器启动，监听 %s", addr)
	return http.ListenAndServe(addr, nil)
}

// handleMessages 处理 WebSocket 消息广播
func (server *WebSocketServer) handleMessages() {
	for {
		msg := <-server.broadcast

		server.mutex.Lock()
		for client := range server.clients {
			err := client.WriteMessage(websocket.TextMessage, msg)
			if err != nil {
				log.Println("发送消息失败:", err)
				client.Close()
				delete(server.clients, client)
			}
		}
		server.mutex.Unlock()
	}
}

// SendMessage 发送消息给所有客户端
func (server *WebSocketServer) SendMessage(msg string) {
	server.broadcast <- []byte(msg)
}

// Close 关闭 WebSocket 服务器
func (server *WebSocketServer) Close() {
	server.mutex.Lock()
	defer server.mutex.Unlock()
	for client := range server.clients {
		client.Close()
		delete(server.clients, client)
	}
	log.Println("WebSocket 服务器已关闭")
}
