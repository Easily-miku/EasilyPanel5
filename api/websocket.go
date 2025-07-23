package api

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"easilypanel5/server"
	"easilypanel5/utils"
)

// WSMessage WebSocket消息结构
type WSMessage struct {
	Type     string      `json:"type"`
	ServerID string      `json:"server_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// WSClient WebSocket客户端
type WSClient struct {
	conn         *websocket.Conn
	send         chan WSMessage
	subscriptions map[string]bool // 订阅的服务器ID
	mutex        sync.RWMutex
}

// WSHub WebSocket连接管理器
type WSHub struct {
	clients    map[*WSClient]bool
	register   chan *WSClient
	unregister chan *WSClient
	broadcast  chan WSMessage
	mutex      sync.RWMutex
}

var hub = &WSHub{
	clients:    make(map[*WSClient]bool),
	register:   make(chan *WSClient),
	unregister: make(chan *WSClient),
	broadcast:  make(chan WSMessage),
}

// 启动WebSocket Hub
func init() {
	go hub.run()

	// 订阅事件管理器的事件
	utils.SubscribeEvent("log_message", func(event utils.Event) {
		BroadcastMessage(WSMessage{
			Type:     event.Type,
			ServerID: event.ServerID,
			Data:     event.Data,
			Error:    event.Error,
		})
	})

	utils.SubscribeEvent("server_status", func(event utils.Event) {
		BroadcastMessage(WSMessage{
			Type:     event.Type,
			ServerID: event.ServerID,
			Data:     event.Data,
			Error:    event.Error,
		})
	})

	utils.SubscribeEvent("download_progress", func(event utils.Event) {
		BroadcastMessage(WSMessage{
			Type:     event.Type,
			ServerID: event.ServerID,
			Data:     event.Data,
			Error:    event.Error,
		})
	})
}

// run 运行WebSocket Hub
func (h *WSHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected, total: %d", len(h.clients))

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected, total: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
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
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastMessage 广播消息
func BroadcastMessage(message WSMessage) {
	select {
	case hub.broadcast <- message:
	default:
		log.Println("WebSocket hub broadcast channel is full")
	}
}

// handleWSConnection 处理WebSocket连接
func handleWSConnection(conn *websocket.Conn) {
	client := &WSClient{
		conn:          conn,
		send:          make(chan WSMessage, 256),
		subscriptions: make(map[string]bool),
	}

	hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 读取客户端消息
func (c *WSClient) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message WSMessage
		if err := c.conn.ReadJSON(&message); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump 向客户端发送消息
func (c *WSClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理客户端消息
func (c *WSClient) handleMessage(message WSMessage) {
	switch message.Type {
	case "subscribe_logs":
		if serverID, ok := message.Data.(string); ok {
			c.mutex.Lock()
			c.subscriptions[serverID] = true
			c.mutex.Unlock()
			
			// 发送确认消息
			c.send <- WSMessage{
				Type:     "subscription_confirmed",
				ServerID: serverID,
				Data:     "subscribed to logs",
			}
		}

	case "unsubscribe_logs":
		if serverID, ok := message.Data.(string); ok {
			c.mutex.Lock()
			delete(c.subscriptions, serverID)
			c.mutex.Unlock()
			
			// 发送确认消息
			c.send <- WSMessage{
				Type:     "subscription_confirmed",
				ServerID: serverID,
				Data:     "unsubscribed from logs",
			}
		}

	case "send_command":
		var cmdData struct {
			ServerID string `json:"server_id"`
			Command  string `json:"command"`
		}
		
		if data, err := json.Marshal(message.Data); err == nil {
			if err := json.Unmarshal(data, &cmdData); err == nil {
				if err := server.SendCommand(cmdData.ServerID, cmdData.Command); err != nil {
					c.send <- WSMessage{
						Type:     "error",
						ServerID: cmdData.ServerID,
						Error:    err.Error(),
					}
				}
			}
		}

	case "get_server_status":
		if serverID, ok := message.Data.(string); ok {
			if srv, exists := server.GetServerStatus(serverID); exists {
				c.send <- WSMessage{
					Type:     "server_status",
					ServerID: serverID,
					Data:     srv,
				}
			} else {
				c.send <- WSMessage{
					Type:     "error",
					ServerID: serverID,
					Error:    "Server not found",
				}
			}
		}

	default:
		c.send <- WSMessage{
			Type:  "error",
			Error: "Unknown message type",
		}
	}
}
