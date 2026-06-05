package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	"backend-server/pkg/logger"
)

// Message WebSocket 消息
type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client WebSocket 客户端
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	UserID uint
	Send   chan []byte
}

// Hub WebSocket 连接管理器
type Hub struct {
	// 已注册的客户端（按用户 ID 分组）
	Clients map[uint]map[*Client]bool

	// 广播通道
	Broadcast chan *Message

	// 注册通道
	Register chan *Client

	// 注销通道
	Unregister chan *Client

	mu sync.RWMutex
}

// NewHub 创建 Hub 实例
func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[uint]map[*Client]bool),
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run 启动 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; !ok {
				h.Clients[client.UserID] = make(map[*Client]bool)
			}
			h.Clients[client.UserID][client] = true
			h.mu.Unlock()
			logger.Info("WebSocket 客户端连接", zap.Uint("user_id", client.UserID))

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.Clients[client.UserID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.Clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			logger.Info("WebSocket 客户端断开", zap.Uint("user_id", client.UserID))

		case msg := <-h.Broadcast:
			// 收集需要清理的客户端，避免在遍历中修改 map
			h.mu.Lock()
			var toRemove []*Client
			data := h.marshal(msg)
			for _, clients := range h.Clients {
				for client := range clients {
					select {
					case client.Send <- data:
					default:
						// 发送缓冲区满，标记为待移除
						toRemove = append(toRemove, client)
					}
				}
			}
			// 清理发送失败的客户端
			for _, client := range toRemove {
				if clients, ok := h.Clients[client.UserID]; ok {
					if _, ok := clients[client]; ok {
						close(client.Send)
						delete(clients, client)
						if len(clients) == 0 {
							delete(h.Clients, client.UserID)
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

// SendToUser 向指定用户发送消息
func (h *Hub) SendToUser(userID uint, msg *Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.Clients[userID]; ok {
		data := h.marshal(msg)
		var toRemove []*Client
		for client := range clients {
			select {
			case client.Send <- data:
			default:
				toRemove = append(toRemove, client)
			}
		}
		// 清理发送失败的客户端
		for _, client := range toRemove {
			if _, ok := clients[client]; ok {
				close(client.Send)
				delete(clients, client)
			}
		}
		if len(clients) == 0 {
			delete(h.Clients, userID)
		}
	}
}

// SendToUsers 向多个用户发送消息
func (h *Hub) SendToUsers(userIDs []uint, msg *Message) {
	for _, uid := range userIDs {
		h.SendToUser(uid, msg)
	}
}

// BroadcastMessage 广播消息
func (h *Hub) BroadcastMessage(msg *Message) {
	h.Broadcast <- msg
}

// OnlineCount 在线用户数
func (h *Hub) OnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

func (h *Hub) marshal(msg *Message) []byte {
	data, err := json.Marshal(msg)
	if err != nil {
		logger.Error("序列化 WebSocket 消息失败", zap.Error(err))
		return nil
	}
	return data
}
