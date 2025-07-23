package api

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"easilypanel5/server"
	"easilypanel5/utils"
)

// 常量定义
const (
	// 写入超时时间
	writeWait = 10 * time.Second

	// 读取超时时间
	pongWait = 60 * time.Second

	// ping间隔时间（必须小于pongWait）
	pingPeriod = (pongWait * 9) / 10

	// 最大消息大小
	maxMessageSize = 512
)

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type     string      `json:"type"`
	ServerID string      `json:"server_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// WSClient WebSocket客户端
type WSClient struct {
	// WebSocket连接
	conn *websocket.Conn

	// 发送消息的缓冲通道
	send chan WSMessage

	// 客户端订阅的服务器ID
	subscriptions map[string]bool

	// 保护subscriptions的互斥锁
	mutex sync.RWMutex

	// 客户端ID（用于调试）
	id string

	// 连接时间
	connectedAt time.Time
}

// WSHub WebSocket连接管理器
type WSHub struct {
	// 注册的客户端
	clients map[*WSClient]bool

	// 注册客户端的通道
	register chan *WSClient

	// 注销客户端的通道
	unregister chan *WSClient

	// 广播消息的通道
	broadcast chan WSMessage

	// 保护clients的互斥锁
	mutex sync.RWMutex
}

// 全局Hub实例
var hub = &WSHub{
	clients:    make(map[*WSClient]bool),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
	broadcast:  make(chan WSMessage, 256), // 增加缓冲区大小
}

// WebSocket升级器配置
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 在生产环境中应该检查origin
		return true
	},
}

// 启动WebSocket Hub
func init() {
	go hub.run()

	// 订阅事件管理器的事件
	utils.SubscribeEvent("log_message", func(event utils.Event) {
		BroadcastMessage(WSMessage{
			Type:      event.Type,
			ServerID:  event.ServerID,
			Data:      event.Data,
			Error:     event.Error,
			Timestamp: time.Now(),
		})
	})

	utils.SubscribeEvent("server_status", func(event utils.Event) {
		BroadcastMessage(WSMessage{
			Type:      event.Type,
			ServerID:  event.ServerID,
			Data:      event.Data,
			Error:     event.Error,
			Timestamp: time.Now(),
		})
	})

	utils.SubscribeEvent("download_progress", func(event utils.Event) {
		BroadcastMessage(WSMessage{
			Type:      event.Type,
			ServerID:  event.ServerID,
			Data:      event.Data,
			Error:     event.Error,
			Timestamp: time.Now(),
		})
	})
}

// run 运行WebSocket Hub
func (h *WSHub) run() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("WebSocket hub panic recovered: %v", r)
			// 重新启动hub
			go h.run()
		}
	}()

	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			clientCount := len(h.clients)
			h.mutex.Unlock()

			log.Printf("WebSocket client [%s] connected at %v, total: %d",
				client.id, client.connectedAt.Format("15:04:05"), clientCount)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				clientCount := len(h.clients)
				h.mutex.Unlock()

				duration := time.Since(client.connectedAt)
				log.Printf("WebSocket client [%s] disconnected after %v, total: %d",
					client.id, duration, clientCount)
			} else {
				h.mutex.Unlock()
			}

		case message := <-h.broadcast:
			h.mutex.RLock()
			clients := make([]*WSClient, 0, len(h.clients))
			for client := range h.clients {
				clients = append(clients, client)
			}
			h.mutex.RUnlock()

			// 在锁外进行消息发送，避免长时间持锁
			for _, client := range clients {
				// 检查客户端是否订阅了该服务器
				if message.ServerID != "" {
					client.mutex.RLock()
					subscribed := client.subscriptions[message.ServerID]
					client.mutex.RUnlock()

					if !subscribed {
						continue
					}
				}

				select {
				case client.send <- message:
					// 消息发送成功
				default:
					// 发送缓冲区满，断开客户端
					log.Printf("Client [%s] send buffer full, disconnecting", client.id)
					h.unregister <- client
				}
			}
		}
	}
}

// BroadcastMessage 广播消息
func BroadcastMessage(message WSMessage) {
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	select {
	case hub.broadcast <- message:
		// 消息发送成功
	default:
		log.Println("WebSocket hub broadcast channel is full, dropping message")
	}
}

// ServeWS 处理WebSocket升级请求
func ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// 生成客户端ID
	clientID := generateClientID()

	client := &WSClient{
		conn:          conn,
		send:          make(chan WSMessage, 256),
		subscriptions: make(map[string]bool),
		id:            clientID,
		connectedAt:   time.Now(),
	}

	// 注册客户端
	hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return time.Now().Format("20060102-150405") + "-" +
		   string(rune('A' + time.Now().Nanosecond()%26))
}

// readPump 读取客户端消息
func (c *WSClient) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	// 设置读取限制和超时
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var message WSMessage
		if err := c.conn.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket client [%s] unexpected close: %v", c.id, err)
			}
			break
		}

		// 处理接收到的消息
		c.handleMessage(message)
	}
}

// writePump 向客户端发送消息
func (c *WSClient) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub关闭了通道
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// 发送JSON消息
			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket client [%s] write error: %v", c.id, err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket client [%s] ping error: %v", c.id, err)
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *WSClient) handleMessage(message WSMessage) {
	switch message.Type {
	case "ping":
		// 响应心跳
		select {
		case c.send <- WSMessage{
			Type:      "pong",
			Data:      "heartbeat",
			Timestamp: time.Now(),
		}:
		default:
			log.Printf("Client [%s] send buffer full on pong", c.id)
		}

	case "subscribe_logs":
		if serverID, ok := message.Data.(string); ok {
			c.mutex.Lock()
			c.subscriptions[serverID] = true
			c.mutex.Unlock()

			log.Printf("Client [%s] subscribed to server [%s] logs", c.id, serverID)

			// 发送确认消息
			select {
			case c.send <- WSMessage{
				Type:      "subscription_confirmed",
				ServerID:  serverID,
				Data:      "subscribed to logs",
				Timestamp: time.Now(),
			}:
			default:
				log.Printf("Client [%s] send buffer full on subscription confirm", c.id)
			}
		}

	case "unsubscribe_logs":
		if serverID, ok := message.Data.(string); ok {
			c.mutex.Lock()
			delete(c.subscriptions, serverID)
			c.mutex.Unlock()

			log.Printf("Client [%s] unsubscribed from server [%s] logs", c.id, serverID)

			// 发送确认消息
			select {
			case c.send <- WSMessage{
				Type:      "subscription_confirmed",
				ServerID:  serverID,
				Data:      "unsubscribed from logs",
				Timestamp: time.Now(),
			}:
			default:
				log.Printf("Client [%s] send buffer full on unsubscription confirm", c.id)
			}
		}

	case "send_command":
		var cmdData struct {
			ServerID string `json:"server_id"`
			Command  string `json:"command"`
		}

		if data, err := json.Marshal(message.Data); err == nil {
			if err := json.Unmarshal(data, &cmdData); err == nil {
				log.Printf("Client [%s] sending command to server [%s]: %s", c.id, cmdData.ServerID, cmdData.Command)
				if err := server.SendCommand(cmdData.ServerID, cmdData.Command); err != nil {
					select {
					case c.send <- WSMessage{
						Type:      "error",
						ServerID:  cmdData.ServerID,
						Error:     err.Error(),
						Timestamp: time.Now(),
					}:
					default:
						log.Printf("Client [%s] send buffer full on command error", c.id)
					}
				}
			}
		}

	case "get_server_status":
		if serverID, ok := message.Data.(string); ok {
			if srv, exists := server.GetServerStatus(serverID); exists {
				select {
				case c.send <- WSMessage{
					Type:      "server_status",
					ServerID:  serverID,
					Data:      srv,
					Timestamp: time.Now(),
				}:
				default:
					log.Printf("Client [%s] send buffer full on server status", c.id)
				}
			} else {
				select {
				case c.send <- WSMessage{
					Type:      "error",
					ServerID:  serverID,
					Error:     "Server not found",
					Timestamp: time.Now(),
				}:
				default:
					log.Printf("Client [%s] send buffer full on server not found error", c.id)
				}
			}
		}

	default:
		log.Printf("Client [%s] sent unknown message type: %s", c.id, message.Type)
		select {
		case c.send <- WSMessage{
			Type:      "error",
			Error:     "Unknown message type: " + message.Type,
			Timestamp: time.Now(),
		}:
		default:
			log.Printf("Client [%s] send buffer full on unknown message error", c.id)
		}
	}
}
