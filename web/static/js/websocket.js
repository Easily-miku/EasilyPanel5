// WebSocket连接状态常量
const WS_STATES = {
    DISCONNECTED: 'disconnected',
    CONNECTING: 'connecting',
    CONNECTED: 'connected',
    RECONNECTING: 'reconnecting',
    DESTROYED: 'destroyed'
};

// WebSocket管理器
class WebSocketManager {
    constructor() {
        // 连接相关
        this.ws = null;
        this.state = WS_STATES.DISCONNECTED;
        this.url = null;

        // 重连相关
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
        this.baseReconnectDelay = 1000;
        this.maxReconnectDelay = 30000;
        this.reconnectTimer = null;

        // 心跳相关
        this.heartbeatInterval = null;
        this.heartbeatTimeout = null;
        this.heartbeatIntervalMs = 30000; // 30秒
        this.heartbeatTimeoutMs = 10000;  // 10秒超时
        this.lastHeartbeatTime = null;
        this.heartbeatFailures = 0;

        // 事件处理
        this.messageHandlers = new Map();
        this.subscriptions = new Set();

        // 连接超时
        this.connectionTimeout = null;
        this.connectionTimeoutMs = 10000; // 10秒连接超时

        // 连接质量监控
        this.connectionQuality = 'unknown'; // excellent, good, poor, bad
        this.latency = 0;
        this.messagesSent = 0;
        this.messagesReceived = 0;
        this.connectionStartTime = null;

        // 状态变化回调
        this.stateChangeCallbacks = new Set();

        // 调试信息
        this.id = this.generateId();
        this.createdAt = new Date();

        console.log(`WebSocket管理器已创建 [${this.id}]`);

        // 初始化状态指示器
        this.initStatusIndicator();
    }
    
    // 生成唯一ID
    generateId() {
        return 'ws-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    }

    // 初始化状态指示器
    initStatusIndicator() {
        // 创建状态指示器元素
        if (!document.getElementById('ws-status-indicator')) {
            const indicator = document.createElement('div');
            indicator.id = 'ws-status-indicator';
            indicator.className = 'ws-status-indicator';
            indicator.innerHTML = `
                <div class="ws-status-dot"></div>
                <span class="ws-status-text">连接中...</span>
                <div class="ws-status-details">
                    <div class="ws-latency">延迟: --ms</div>
                    <div class="ws-quality">质量: --</div>
                </div>
            `;

            // 添加样式
            if (!document.getElementById('ws-status-styles')) {
                const styles = document.createElement('style');
                styles.id = 'ws-status-styles';
                styles.textContent = `
                    .ws-status-indicator {
                        position: fixed;
                        top: 10px;
                        right: 10px;
                        background: rgba(0, 0, 0, 0.8);
                        color: white;
                        padding: 8px 12px;
                        border-radius: 6px;
                        font-size: 12px;
                        z-index: 10000;
                        display: flex;
                        align-items: center;
                        gap: 6px;
                        transition: all 0.3s ease;
                    }
                    .ws-status-dot {
                        width: 8px;
                        height: 8px;
                        border-radius: 50%;
                        background: #666;
                        transition: background-color 0.3s ease;
                    }
                    .ws-status-indicator.connected .ws-status-dot { background: #4CAF50; }
                    .ws-status-indicator.connecting .ws-status-dot { background: #FF9800; animation: pulse 1s infinite; }
                    .ws-status-indicator.disconnected .ws-status-dot { background: #F44336; }
                    .ws-status-indicator.reconnecting .ws-status-dot { background: #FF9800; animation: pulse 1s infinite; }
                    .ws-status-details {
                        font-size: 10px;
                        opacity: 0.8;
                        margin-left: 8px;
                    }
                    @keyframes pulse {
                        0%, 100% { opacity: 1; }
                        50% { opacity: 0.5; }
                    }
                `;
                document.head.appendChild(styles);
            }

            document.body.appendChild(indicator);
        }

        this.updateStatusIndicator();
    }

    // 更新状态指示器
    updateStatusIndicator() {
        const indicator = document.getElementById('ws-status-indicator');
        if (!indicator) return;

        indicator.className = `ws-status-indicator ${this.state}`;

        const statusText = indicator.querySelector('.ws-status-text');
        const latencyEl = indicator.querySelector('.ws-latency');
        const qualityEl = indicator.querySelector('.ws-quality');

        if (statusText) {
            switch (this.state) {
                case WS_STATES.CONNECTED:
                    statusText.textContent = '已连接';
                    break;
                case WS_STATES.CONNECTING:
                    statusText.textContent = '连接中...';
                    break;
                case WS_STATES.DISCONNECTED:
                    statusText.textContent = '已断开';
                    break;
                case WS_STATES.RECONNECTING:
                    statusText.textContent = `重连中... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`;
                    break;
                default:
                    statusText.textContent = '未知状态';
            }
        }

        if (latencyEl) {
            latencyEl.textContent = `延迟: ${this.latency > 0 ? this.latency + 'ms' : '--'}`;
        }

        if (qualityEl) {
            qualityEl.textContent = `质量: ${this.getQualityText()}`;
        }
    }

    // 获取连接质量文本
    getQualityText() {
        switch (this.connectionQuality) {
            case 'excellent': return '优秀';
            case 'good': return '良好';
            case 'poor': return '较差';
            case 'bad': return '很差';
            default: return '--';
        }
    }

    // 设置状态并更新指示器
    setState(newState) {
        if (this.state !== newState) {
            const oldState = this.state;
            this.state = newState;

            console.log(`WebSocket管理器 [${this.id}] 状态变化: ${oldState} -> ${newState}`);

            // 更新状态指示器
            this.updateStatusIndicator();

            // 通知状态变化回调
            this.stateChangeCallbacks.forEach(callback => {
                try {
                    callback(newState, oldState);
                } catch (error) {
                    console.error('状态变化回调执行失败:', error);
                }
            });
        }
    }

    // 添加状态变化回调
    onStateChange(callback) {
        this.stateChangeCallbacks.add(callback);
        return () => this.stateChangeCallbacks.delete(callback);
    }

    // 计算连接质量
    calculateConnectionQuality() {
        if (this.state !== WS_STATES.CONNECTED) {
            this.connectionQuality = 'unknown';
            return;
        }

        // 基于延迟和心跳失败次数计算质量
        if (this.latency < 100 && this.heartbeatFailures === 0) {
            this.connectionQuality = 'excellent';
        } else if (this.latency < 300 && this.heartbeatFailures < 2) {
            this.connectionQuality = 'good';
        } else if (this.latency < 1000 && this.heartbeatFailures < 5) {
            this.connectionQuality = 'poor';
        } else {
            this.connectionQuality = 'bad';
        }
    }

    // 连接WebSocket
    connect(url = null) {
        if (this.state === WS_STATES.DESTROYED) {
            console.warn(`WebSocket管理器 [${this.id}] 已销毁，无法连接`);
            return false;
        }

        if (this.state === WS_STATES.CONNECTING) {
            console.log(`WebSocket管理器 [${this.id}] 正在连接中，跳过重复连接`);
            return false;
        }

        if (this.state === WS_STATES.CONNECTED) {
            console.log(`WebSocket管理器 [${this.id}] 已连接，跳过重复连接`);
            return true;
        }

        // 设置URL
        if (url) {
            this.url = url;
        } else if (!this.url) {
            const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
            this.url = `${protocol}//${window.location.host}/ws`;
        }

        this.setState(WS_STATES.CONNECTING);
        console.log(`WebSocket管理器 [${this.id}] 开始连接: ${this.url}`);

        // 清除之前的定时器
        this.clearTimers();

        // 设置连接超时
        this.connectionTimeout = setTimeout(() => {
            if (this.state === WS_STATES.CONNECTING) {
                console.error(`WebSocket管理器 [${this.id}] 连接超时`);
                this.handleConnectionError('连接超时');
            }
        }, this.connectionTimeoutMs);

        try {
            this.ws = new WebSocket(this.url);
            this.setupEventHandlers();
            return true;
        } catch (error) {
            console.error(`WebSocket管理器 [${this.id}] 创建连接失败:`, error);
            this.handleConnectionError(`创建连接失败: ${error.message}`);
            return false;
        }
    }

    // 设置WebSocket事件处理器
    setupEventHandlers() {
        if (!this.ws) return;

        this.ws.onopen = (event) => {
            console.log(`WebSocket管理器 [${this.id}] 连接已建立`);
            this.setState(WS_STATES.CONNECTED);
            this.reconnectAttempts = 0;
            this.connectionStartTime = Date.now();
            this.heartbeatFailures = 0;

            // 清除连接超时
            this.clearConnectionTimeout();

            // 启动心跳
            this.startHeartbeat();

            // 重新订阅之前的服务器
            this.resubscribe();

            // 通知连接成功
            this.notifyHandlers('connected', {
                reconnected: this.reconnectAttempts > 0
            });
        };
            
        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleMessage(message);
            } catch (error) {
                console.error(`WebSocket管理器 [${this.id}] 解析消息失败:`, error);
            }
        };
            
        this.ws.onclose = (event) => {
            console.log(`WebSocket管理器 [${this.id}] 连接已关闭: ${event.code} ${event.reason}`);

            const wasConnected = this.state === WS_STATES.CONNECTED;
            this.setState(WS_STATES.DISCONNECTED);
            this.ws = null;

            // 停止心跳和清除定时器
            this.stopHeartbeat();
            this.clearTimers();

            // 通知断开连接
            this.notifyHandlers('disconnected', {
                code: event.code,
                reason: event.reason,
                wasConnected: wasConnected
            });

            // 判断是否需要重连
            if (this.shouldReconnect(event.code)) {
                this.scheduleReconnect(`连接关闭: ${event.code} ${event.reason}`);
            }
        };

        this.ws.onerror = (error) => {
            console.error(`WebSocket管理器 [${this.id}] 连接错误:`, error);
            this.notifyHandlers('error', error);
        };
    }
            
    // 判断是否应该重连
    shouldReconnect(closeCode) {
        if (this.state === WS_STATES.DESTROYED) {
            return false;
        }

        // 正常关闭不重连
        if (closeCode === 1000 || closeCode === 1001) {
            return false;
        }

        // 达到最大重试次数不重连
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.warn(`WebSocket管理器 [${this.id}] 达到最大重试次数，停止重连`);
            this.notifyHandlers('reconnect_failed', {
                attempts: this.reconnectAttempts
            });
            return false;
        }

        return true;
    }

    // 安排重连
    scheduleReconnect(reason) {
        if (this.state === WS_STATES.DESTROYED) {
            return;
        }

        this.setState(WS_STATES.RECONNECTING);
        this.reconnectAttempts++;

        // 计算重连延迟（指数退避）
        const delay = Math.min(
            this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts - 1),
            this.maxReconnectDelay
        );

        console.log(`WebSocket管理器 [${this.id}] 将在 ${delay}ms 后重连 (${this.reconnectAttempts}/${this.maxReconnectAttempts}): ${reason}`);

        this.reconnectTimer = setTimeout(() => {
            if (this.state === WS_STATES.RECONNECTING) {
                this.connect();
            }
        }, delay);
    }

    // 处理连接错误
    handleConnectionError(reason) {
        console.error(`WebSocket管理器 [${this.id}] 连接错误: ${reason}`);

        if (this.state === WS_STATES.CONNECTING) {
            this.setState(WS_STATES.DISCONNECTED);
        }

        this.clearTimers();

        if (this.ws) {
            this.ws.close();
            this.ws = null;
        }

        this.scheduleReconnect(reason);
    }

    // 启动心跳
    startHeartbeat() {
        this.stopHeartbeat(); // 确保没有重复的心跳

        console.log(`WebSocket管理器 [${this.id}] 启动心跳检测`);

        this.heartbeatInterval = setInterval(() => {
            if (this.state === WS_STATES.CONNECTED && this.ws) {
                this.sendHeartbeat();
            }
        }, this.heartbeatIntervalMs);
    }

    // 发送心跳
    sendHeartbeat() {
        this.lastHeartbeatTime = Date.now();

        if (!this.send({ type: 'ping', timestamp: this.lastHeartbeatTime })) {
            console.warn(`WebSocket管理器 [${this.id}] 心跳发送失败`);
            this.heartbeatFailures++;
            this.calculateConnectionQuality();
            this.updateStatusIndicator();
            return;
        }

        // 设置心跳超时
        this.heartbeatTimeout = setTimeout(() => {
            console.error(`WebSocket管理器 [${this.id}] 心跳超时，关闭连接`);
            this.heartbeatFailures++;
            this.calculateConnectionQuality();
            this.updateStatusIndicator();

            if (this.ws) {
                this.ws.close(1000, 'Heartbeat timeout');
            }
        }, this.heartbeatTimeoutMs);
    }

    // 停止心跳
    stopHeartbeat() {
        if (this.heartbeatInterval) {
            clearInterval(this.heartbeatInterval);
            this.heartbeatInterval = null;
        }

        if (this.heartbeatTimeout) {
            clearTimeout(this.heartbeatTimeout);
            this.heartbeatTimeout = null;
        }
    }

    // 清除连接超时
    clearConnectionTimeout() {
        if (this.connectionTimeout) {
            clearTimeout(this.connectionTimeout);
            this.connectionTimeout = null;
        }
    }

    // 清除所有定时器
    clearTimers() {
        this.clearConnectionTimeout();
        this.stopHeartbeat();

        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }
    }
    
    // 处理接收到的消息
    handleMessage(message) {
        switch (message.type) {
            case 'pong':
                // 收到心跳响应，清除超时
                if (this.heartbeatTimeout) {
                    clearTimeout(this.heartbeatTimeout);
                    this.heartbeatTimeout = null;
                }

                // 计算延迟
                if (this.lastHeartbeatTime && message.timestamp) {
                    this.latency = Date.now() - message.timestamp;
                } else if (this.lastHeartbeatTime) {
                    this.latency = Date.now() - this.lastHeartbeatTime;
                }

                // 重置心跳失败计数
                this.heartbeatFailures = 0;

                // 更新连接质量
                this.calculateConnectionQuality();
                this.updateStatusIndicator();
                break;

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
                console.log(`WebSocket管理器 [${this.id}] 订阅确认: ${message.server_id}`);
                break;

            case 'error':
                console.error(`WebSocket管理器 [${this.id}] 服务器错误:`, message.error);
                this.notifyHandlers('server_error', message);
                break;

            default:
                console.log(`WebSocket管理器 [${this.id}] 未知消息类型: ${message.type}`);
        }
    }
    
    // 发送消息
    send(message) {
        if (this.state !== WS_STATES.CONNECTED || !this.ws) {
            console.warn(`WebSocket管理器 [${this.id}] 未连接，无法发送消息:`, message.type);
            return false;
        }

        try {
            this.ws.send(JSON.stringify(message));
            return true;
        } catch (error) {
            console.error(`WebSocket管理器 [${this.id}] 发送消息失败:`, error);
            return false;
        }
    }

    // 订阅服务器日志
    subscribeToLogs(serverId) {
        this.subscriptions.add(serverId);
        return this.send({
            type: 'subscribe_logs',
            data: serverId
        });
    }

    // 取消订阅服务器日志
    unsubscribeFromLogs(serverId) {
        this.subscriptions.delete(serverId);
        return this.send({
            type: 'unsubscribe_logs',
            data: serverId
        });
    }

    // 发送服务器命令
    sendCommand(serverId, command) {
        return this.send({
            type: 'send_command',
            data: {
                server_id: serverId,
                command: command
            }
        });
    }

    // 获取服务器状态
    getServerStatus(serverId) {
        return this.send({
            type: 'get_server_status',
            data: serverId
        });
    }

    // 重新订阅所有服务器
    resubscribe() {
        this.subscriptions.forEach(serverId => {
            this.subscribeToLogs(serverId);
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
                    console.error(`WebSocket管理器 [${this.id}] 事件处理器错误 (${event}):`, error);
                }
            });
        }
    }

    // 获取连接状态
    isConnected() {
        return this.state === WS_STATES.CONNECTED;
    }

    // 获取当前状态
    getState() {
        return this.state;
    }

    // 手动重连
    reconnect() {
        console.log(`WebSocket管理器 [${this.id}] 手动重连`);
        this.reconnectAttempts = 0;

        if (this.ws) {
            this.ws.close(1000, 'Manual reconnect');
        } else {
            this.connect();
        }
    }

    // 订阅服务器日志
    subscribeToLogs(serverId) {
        if (this.send({
            type: 'subscribe_logs',
            data: serverId
        })) {
            console.log(`WebSocket管理器 [${this.id}] 订阅服务器日志: ${serverId}`);
            this.subscriptions.add(`logs_${serverId}`);
        }
    }

    // 取消订阅服务器日志
    unsubscribeFromLogs(serverId) {
        if (this.send({
            type: 'unsubscribe_logs',
            data: serverId
        })) {
            console.log(`WebSocket管理器 [${this.id}] 取消订阅服务器日志: ${serverId}`);
            this.subscriptions.delete(`logs_${serverId}`);
        }
    }

    // 发送服务器命令
    sendServerCommand(serverId, command) {
        return this.send({
            type: 'server_command',
            data: {
                server_id: serverId,
                command: command
            }
        });
    }

    // 销毁WebSocket管理器
    destroy() {
        console.log(`WebSocket管理器 [${this.id}] 正在销毁`);

        this.setState(WS_STATES.DESTROYED);

        // 清除所有定时器
        this.clearTimers();

        // 关闭连接
        if (this.ws) {
            this.ws.close(1000, 'Client destroying');
            this.ws = null;
        }

        // 清除订阅和处理器
        this.subscriptions.clear();
        this.messageHandlers.clear();

        console.log(`WebSocket管理器 [${this.id}] 已销毁`);
    }

    // 关闭连接（保持兼容性）
    close() {
        this.destroy();
    }

    // 获取调试信息
    getDebugInfo() {
        return {
            id: this.id,
            state: this.state,
            reconnectAttempts: this.reconnectAttempts,
            subscriptions: Array.from(this.subscriptions),
            createdAt: this.createdAt,
            url: this.url
        };
    }
}

// 全局WebSocket实例
let wsManager = null;

// 初始化WebSocket连接
function initWebSocket() {
    // 如果已存在实例，先销毁
    if (wsManager) {
        console.log('销毁现有WebSocket实例');
        wsManager.destroy();
        wsManager = null;
    }

    console.log('创建新的WebSocket管理器');
    wsManager = new WebSocketManager();

    return wsManager;
}

// 启动WebSocket连接
function startWebSocketConnection() {
    if (!wsManager) {
        wsManager = initWebSocket();
    }

    if (wsManager.getState() === WS_STATES.DISCONNECTED) {
        console.log('启动WebSocket连接');
        wsManager.connect();
    }
}

// 获取WebSocket管理器实例
function getWebSocketManager() {
    return wsManager;
}

// 页面可见性变化处理
document.addEventListener('visibilitychange', () => {
    if (wsManager) {
        if (document.hidden) {
            console.log('页面隐藏，暂停WebSocket重连');
            // 不关闭连接，只是暂停重连
            wsManager.maxReconnectAttempts = 0;
        } else {
            console.log('页面显示，恢复WebSocket重连');
            wsManager.maxReconnectAttempts = 5;

            // 如果连接断开，尝试重连
            if (wsManager.getState() === WS_STATES.DISCONNECTED) {
                wsManager.connect();
            }
        }
    }
});

// 页面卸载时清理资源
window.addEventListener('beforeunload', () => {
    if (wsManager) {
        console.log('页面卸载，清理WebSocket资源');
        wsManager.destroy();
        wsManager = null;
    }
});

// 导出到全局作用域
window.WebSocketManager = WebSocketManager;
window.WS_STATES = WS_STATES;
window.initWebSocket = initWebSocket;
window.startWebSocketConnection = startWebSocketConnection;
window.getWebSocketManager = getWebSocketManager;
