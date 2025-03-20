package websocket

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// 重连间隔时间
	reconnectInterval = 5 * time.Second
	// Ping 间隔时间
	pingInterval = 10 * time.Second
)

// WebSocketClient WebSocket 客户端结构体
type WebSocketClient struct {
	conn       *websocket.Conn
	stopChan   chan struct{}
	serverAddr string
	sendMutex  sync.Mutex    // 确保发送时的线程安全
	sendSignal chan struct{} // 用来控制消息发送
	pingTicker *time.Ticker
	router     *Router // 添加 Router 字段
}

// NewWebSocketClient 创建一个新的 WebSocket 客户端实例并连接到 WebSocket 服务
func NewWebSocketClient(serverAddr string) (*WebSocketClient, error) {
	client := &WebSocketClient{
		stopChan:   make(chan struct{}),
		serverAddr: serverAddr,
		sendSignal: make(chan struct{}, 1), // 控制发送
		router:     NewRouter(),            // 初始化 Router
	}

	if err := client.Connect(); err != nil {
		return nil, err
	}
	return client, nil
}

// RegisterRoute 注册一个消息类型及其对应的处理函数
func (client *WebSocketClient) RegisterRoute(action string, handler func(message []byte)) {
	client.router.RegisterRoute(action, handler)
}

// Connect 连接到 WebSocket 服务
func (client *WebSocketClient) Connect() error {
	// 尝试连接 WebSocket 服务
	conn, _, err := websocket.DefaultDialer.Dial(client.serverAddr, nil)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	client.conn = conn
	log.Println("Connected to WebSocket server")

	// 启动处理 ping 消息
	client.pingTicker = time.NewTicker(pingInterval)
	go client.sendPingMessages()

	// 启动读取消息
	go client.readMessages()

	// 启动消息发送处理
	go client.handleMessageSending()

	return nil
}

// ReadMessages 持续读取 WebSocket 消息
func (client *WebSocketClient) readMessages() {
	for {
		_, msg, err := client.conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			client.reconnect()
			return
		}
		log.Printf("Received message: %s", msg)
		client.router.HandleMessage(msg) // 使用 Router 处理接收到的消息
	}
}

// SendMessage 发送消息到 WebSocket 服务
func (client *WebSocketClient) SendMessage(msg string) error {
	client.sendMutex.Lock()
	defer client.sendMutex.Unlock()

	// 发送消息
	err := client.conn.WriteMessage(websocket.TextMessage, []byte(msg))
	if err != nil {
		return fmt.Errorf("send error: %v", err)
	}
	return nil
}

// sendPingMessages 定时发送 Ping 消息
func (client *WebSocketClient) sendPingMessages() {
	for {
		select {
		case <-client.stopChan:
			return
		case <-client.pingTicker.C:
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Ping error: %v", err)
			}
		}
	}
}

// Reconnect 处理连接断开后的重连逻辑
func (client *WebSocketClient) reconnect() {
	for {
		log.Println("Attempting to reconnect...")
		// 等待重连间隔时间
		time.Sleep(reconnectInterval)

		// 尝试重新连接
		err := client.Connect()
		if err == nil {
			log.Println("Reconnected to WebSocket server")
			// 恢复消息发送的 goroutine
			client.sendSignal <- struct{}{}
			break
		}
		log.Printf("Reconnect failed: %v", err)
	}
}

// Close 关闭 WebSocket 连接
func (client *WebSocketClient) Close() error {
	close(client.stopChan)
	if client.pingTicker != nil {
		client.pingTicker.Stop()
	}
	return client.conn.Close()
}

// handleMessageSending 处理消息的发送
func (client *WebSocketClient) handleMessageSending() {
	for {
		select {
		case <-client.stopChan:
			return
		case <-client.sendSignal:
			// 启动消息发送
			for {
				err := client.SendMessage("Hello, WebSocket!")
				if err != nil {
					log.Printf("Error sending message: %v", err)
					break
				}
				// 每隔 2 秒发送一次消息
				time.Sleep(2 * time.Second)
			}
		}
	}
}
