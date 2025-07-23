package utils

import (
	"sync"
)

// Event 事件结构
type Event struct {
	Type     string      `json:"type"`
	ServerID string      `json:"server_id,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Error    string      `json:"error,omitempty"`
}

// EventHandler 事件处理器函数类型
type EventHandler func(event Event)

// EventManager 事件管理器
type EventManager struct {
	handlers map[string][]EventHandler
	mutex    sync.RWMutex
}

var globalEventManager *EventManager
var once sync.Once

// GetEventManager 获取全局事件管理器实例
func GetEventManager() *EventManager {
	once.Do(func() {
		globalEventManager = &EventManager{
			handlers: make(map[string][]EventHandler),
		}
	})
	return globalEventManager
}

// Subscribe 订阅事件
func (em *EventManager) Subscribe(eventType string, handler EventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	if em.handlers[eventType] == nil {
		em.handlers[eventType] = make([]EventHandler, 0)
	}
	em.handlers[eventType] = append(em.handlers[eventType], handler)
}

// Unsubscribe 取消订阅事件
func (em *EventManager) Unsubscribe(eventType string, handler EventHandler) {
	em.mutex.Lock()
	defer em.mutex.Unlock()
	
	handlers := em.handlers[eventType]
	for i, h := range handlers {
		// 注意：这里比较函数指针可能不准确，实际使用中可能需要其他方式
		if &h == &handler {
			em.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
}

// Emit 发送事件
func (em *EventManager) Emit(event Event) {
	em.mutex.RLock()
	handlers := em.handlers[event.Type]
	em.mutex.RUnlock()
	
	for _, handler := range handlers {
		go func(h EventHandler) {
			defer func() {
				if r := recover(); r != nil {
					// 处理panic，避免一个处理器的错误影响其他处理器
				}
			}()
			h(event)
		}(handler)
	}
}

// 便捷函数
func EmitEvent(eventType string, serverID string, data interface{}) {
	GetEventManager().Emit(Event{
		Type:     eventType,
		ServerID: serverID,
		Data:     data,
	})
}

func EmitError(eventType string, serverID string, errorMsg string) {
	GetEventManager().Emit(Event{
		Type:     eventType,
		ServerID: serverID,
		Error:    errorMsg,
	})
}

func SubscribeEvent(eventType string, handler EventHandler) {
	GetEventManager().Subscribe(eventType, handler)
}
