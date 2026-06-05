package handler

import (
	"net/http"

	"backend-server/internal/middleware"
	"backend-server/internal/ws"
	"backend-server/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"

	_ "backend-server/pkg/response" // swagger 类型引用
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 生产环境应限制来源
	},
}

// WSHandler WebSocket 处理器
type WSHandler struct {
	hub *ws.Hub
}

// NewWSHandler 创建 WebSocket 处理器
func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

// Handle WebSocket 连接处理
// @Summary      WebSocket 连接
// @Description  建立 WebSocket 连接，用于实时消息推送。客户端需在 Header 中携带 JWT Token 完成认证后升级协议。
// @Tags         WebSocket
// @Produce      json
// @Security     BearerAuth
// @Success      101  {string}  string  "Switching Protocols (WebSocket 升级成功)"
// @Failure      401  {object}  response.Response "未授权"
// @Router       /ws [get]
func (h *WSHandler) Handle(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}

	client := &ws.Client{
		Hub:    h.hub,
		Conn:   conn,
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	h.hub.Register <- client

	// 启动读写协程
	go h.writePump(client)
	go h.readPump(client)
}

// readPump 读取客户端消息
func (h *WSHandler) readPump(client *ws.Client) {
	defer func() {
		h.hub.Unregister <- client
		client.Conn.Close()
	}()

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

// writePump 向客户端写入消息
func (h *WSHandler) writePump(client *ws.Client) {
	defer client.Conn.Close()

	for msg := range client.Send {
		if err := client.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}
