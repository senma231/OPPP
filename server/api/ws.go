package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/senma231/p3/server/monitor"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，生产环境应该限制
	},
}

// WSHandler WebSocket 处理器
type WSHandler struct {
	monitor    *monitor.Monitor
	clients    map[*websocket.Conn]string
	clientsMu  sync.Mutex
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(m *monitor.Monitor) *WSHandler {
	return &WSHandler{
		monitor: m,
		clients: make(map[*websocket.Conn]string),
	}
}

// HandleWS 处理 WebSocket 连接
func (h *WSHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	// 升级 HTTP 连接为 WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("升级 WebSocket 失败: %v\n", err)
		return
	}
	
	// 生成客户端 ID
	clientID := fmt.Sprintf("%s-%d", r.RemoteAddr, time.Now().UnixNano())
	
	// 添加到客户端列表
	h.clientsMu.Lock()
	h.clients[conn] = clientID
	h.clientsMu.Unlock()
	
	// 订阅事件
	eventCh := h.monitor.Subscribe(clientID)
	
	// 清理函数
	defer func() {
		h.monitor.Unsubscribe(clientID)
		conn.Close()
		h.clientsMu.Lock()
		delete(h.clients, conn)
		h.clientsMu.Unlock()
	}()
	
	// 处理接收的消息
	go func() {
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					fmt.Printf("读取 WebSocket 消息失败: %v\n", err)
				}
				break
			}
		}
	}()
	
	// 发送事件
	for event := range eventCh {
		data, err := json.Marshal(event)
		if err != nil {
			fmt.Printf("序列化事件失败: %v\n", err)
			continue
		}
		
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Printf("发送 WebSocket 消息失败: %v\n", err)
			break
		}
	}
}

// BroadcastMessage 广播消息
func (h *WSHandler) BroadcastMessage(messageType string, data interface{}) {
	message := map[string]interface{}{
		"type": messageType,
		"data": data,
		"time": time.Now(),
	}
	
	messageJSON, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("序列化消息失败: %v\n", err)
		return
	}
	
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()
	
	for conn := range h.clients {
		if err := conn.WriteMessage(websocket.TextMessage, messageJSON); err != nil {
			fmt.Printf("发送 WebSocket 消息失败: %v\n", err)
			conn.Close()
			delete(h.clients, conn)
		}
	}
}
