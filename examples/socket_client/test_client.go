package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/code-serenade/go-kits/websocket"
)

func main() {
	// 创建 WebSocket 客户端实例，传入 WebSocket 服务地址
	client, err := websocket.NewWebSocketClient("ws://localhost:13785/ws")
	if err != nil {
		log.Fatalf("Failed to create WebSocket client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing WebSocket: %v", err)
		}
	}()

	// 注册路由处理函数
	client.RegisterRoute("sms_code", SmsCodeHandler)
	client.RegisterRoute("voice_code", VoiceCodeHandler)

	// 启动一个 goroutine 来交替发送消息
	go sendMessages(client)

	// 模拟程序运行，持续运行，直到用户终止
	select {}
}

func sendMessages(client *websocket.WebSocketClient) {
	messages := []websocket.MessageData{
		{
			Action: "sms_code",
			Data:   json.RawMessage(`{"code":"123456"}`),
		},
		{
			Action: "voice_code",
			Data:   json.RawMessage(`{"code":"654321"}`),
		},
	}

	for {
		for _, msg := range messages {
			// 将消息转化为 JSON 格式
			msgJSON, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshalling message: %v", err)
				return
			}

			// 发送消息
			err = client.SendMessage(string(msgJSON))
			if err != nil {
				log.Printf("Error sending message: %v", err)
				break
			}

			// 每隔 2 秒发送一次消息
			time.Sleep(2 * time.Second)
		}
	}
}

// action: sms_code
type SmsData struct {
	Code string `json:"code"`
}

func SmsCodeHandler(data []byte) {
	var smsData SmsData
	if err := json.Unmarshal(data, &smsData); err != nil {
		log.Printf("Error unmarshalling SmsData: %v", err)
		return
	}
	log.Printf("Received SMS Code: %s", smsData.Code)
}

// action: voice_code
type VoiceData struct {
	Code string `json:"code"`
}

func VoiceCodeHandler(data []byte) {
	var voiceData VoiceData
	if err := json.Unmarshal(data, &voiceData); err != nil {
		log.Printf("Error unmarshalling VoiceData: %v", err)
		return
	}
	log.Printf("Received Voice Code: %s", voiceData.Code)
}
