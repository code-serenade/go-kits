package websocket

import (
	"encoding/json"
	"log"
)

// Router 负责管理消息路由
type Router struct {
	routes map[string]func(message []byte) // 路由映射: 将 action 映射到处理函数
}

// NewRouter 创建一个新的 Router 实例
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]func(message []byte)),
	}
}

// RegisterRoute 注册一个消息类型及其对应的处理函数
func (r *Router) RegisterRoute(action string, handler func(message []byte)) {
	r.routes[action] = handler
}

// HandleMessage 根据消息的 action 字段调用相应的处理函数
func (r *Router) HandleMessage(message []byte) {
	// 假设消息格式为 JSON，我们需要解析它
	// 在这里我们假设我们已经有了 action 字段和消息体
	action, data := getActionFromMessage(message) // 解析 action 字段（假设实现此方法）
	if handler, exists := r.routes[action]; exists {
		// TODO 这里可以根据需要对 data 进行解析 handler
		handler(data) // 调用对应的处理函数
	} else {
		log.Printf("没有为 action %s 注册处理函数", action)
	}
}

type MessageData struct {
	Action string          `json:"action"` // "action" 字段映射到 Go 结构体的 Action 字段
	Data   json.RawMessage `json:"data"`   // "data" 字段映射到 Go 结构体的 Data 字段，保持原始 JSON 数据
}

// getActionFromMessage 解析消息并返回 action 字段
func getActionFromMessage(message []byte) (string, json.RawMessage) {
	var msgData MessageData

	// 使用 json.Unmarshal 解析消息
	if err := json.Unmarshal(message, &msgData); err != nil {
		log.Printf("Error unmarshalling JSON: %v", err)
		return "", nil // 如果解析失败，返回空字符串和 nil
	}

	// 返回解析后的 action 和 data
	return msgData.Action, msgData.Data
}

