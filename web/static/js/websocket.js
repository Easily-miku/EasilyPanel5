// WebSocket管理器
class WebSocketManager {
    constructor() {
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.reconnectDelay = 1000;
        this.isConnecting = false;
        this.messageHandlers = new Map();
        this.subscriptions = new Set();
        
        this.connect();
    }
    
    connect() {
        if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.CONNECTING)) {
            return;
        }
        
        this.isConnecting = true;
        
        try {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            const wsUrl = `${protocol}//${window.location.host}/ws`;
            
            this.ws = new WebSocket(wsUrl);
            
            this.ws.onopen = () => {
                console.log('WebSocket连接已建立');
                this.isConnecting = false;
                this.reconnectAttempts = 0;
                
                // 重新订阅之前的服务器
                this.subscriptions.forEach(serverId => {
                    this.subscribeToLogs(serverId);
                });
                
                this.notifyHandlers('connected', null);
            };
            
            this.ws.onmessage = (event) => {
                try {
                    const message = JSON.parse(event.data);
                    this.handleMessage(message);
                } catch (error) {
                    console.error('解析WebSocket消息失败:', error);
                }
            };
            
            this.ws.onclose = (event) => {
                console.log('WebSocket连接已关闭:', event.code, event.reason);
                this.isConnecting = false;
                this.ws = null;
                
                this.notifyHandlers('disconnected', { code: event.code, reason: event.reason });
                
                // 自动重连
                if (this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.reconnectAttempts++;
                    console.log(`尝试重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts})...`);
                    
                    setTimeout(() => {
                        this.connect();
                    }, this.reconnectDelay * this.reconnectAttempts);
                } else {
                    console.error('WebSocket重连失败，已达到最大重试次数');
                    this.notifyHandlers('reconnect_failed', null);
                }
            };
            
            this.ws.onerror = (error) => {
                console.error('WebSocket错误:', error);
                this.isConnecting = false;
                this.notifyHandlers('error', error);
            };
            
        } catch (error) {
            console.error('创建WebSocket连接失败:', error);
            this.isConnecting = false;
        }
    }
    
    handleMessage(message) {
        console.log('收到WebSocket消息:', message);
        
        switch (message.type) {
            case 'log_message':
                this.notifyHandlers('log_message', message);
                break;
                
            case 'server_status':
                this.notifyHandlers('server_status', message);
                break;
                
            case 'download_progress':
                this.notifyHandlers('download_progress', message);
                break;
                
            case 'subscription_confirmed':
                console.log('订阅确认:', message.server_id);
                break;
                
            case 'error':
                console.error('服务器错误:', message.error);
                this.notifyHandlers('server_error', message);
                break;
                
            default:
                console.log('未知消息类型:', message.type);
        }
    }
    
    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
            return true;
        } else {
            console.warn('WebSocket未连接，无法发送消息');
            return false;
        }
    }
    
    subscribeToLogs(serverId) {
        this.subscriptions.add(serverId);
        return this.send({
            type: 'subscribe_logs',
            data: serverId
        });
    }
    
    unsubscribeFromLogs(serverId) {
        this.subscriptions.delete(serverId);
        return this.send({
            type: 'unsubscribe_logs',
            data: serverId
        });
    }
    
    sendCommand(serverId, command) {
        return this.send({
            type: 'send_command',
            data: {
                server_id: serverId,
                command: command
            }
        });
    }
    
    getServerStatus(serverId) {
        return this.send({
            type: 'get_server_status',
            data: serverId
        });
    }
    
    // 事件处理器管理
    on(event, handler) {
        if (!this.messageHandlers.has(event)) {
            this.messageHandlers.set(event, []);
        }
        this.messageHandlers.get(event).push(handler);
    }
    
    off(event, handler) {
        if (this.messageHandlers.has(event)) {
            const handlers = this.messageHandlers.get(event);
            const index = handlers.indexOf(handler);
            if (index > -1) {
                handlers.splice(index, 1);
            }
        }
    }
    
    notifyHandlers(event, data) {
        if (this.messageHandlers.has(event)) {
            this.messageHandlers.get(event).forEach(handler => {
                try {
                    handler(data);
                } catch (error) {
                    console.error(`事件处理器错误 (${event}):`, error);
                }
            });
        }
    }
    
    // 获取连接状态
    isConnected() {
        return this.ws && this.ws.readyState === WebSocket.OPEN;
    }
    
    // 手动重连
    reconnect() {
        if (this.ws) {
            this.ws.close();
        }
        this.reconnectAttempts = 0;
        this.connect();
    }
    
    // 关闭连接
    close() {
        this.maxReconnectAttempts = 0; // 禁用自动重连
        if (this.ws) {
            this.ws.close();
        }
    }
}

// 全局WebSocket实例
let wsManager = null;

// 初始化WebSocket连接
function initWebSocket() {
    if (!wsManager) {
        wsManager = new WebSocketManager();
    }
    return wsManager;
}

// 获取WebSocket管理器实例
function getWebSocketManager() {
    return wsManager || initWebSocket();
}

// 导出到全局作用域
window.WebSocketManager = WebSocketManager;
window.initWebSocket = initWebSocket;
window.getWebSocketManager = getWebSocketManager;
