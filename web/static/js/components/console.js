// 服务器控制台组件
class ServerConsoleComponent {
    constructor(serverId, options = {}) {
        this.serverId = serverId;
        this.options = {
            maxLines: 1000,
            autoScroll: true,
            showTimestamp: true,
            enableInput: true,
            ...options
        };
        
        this.logs = [];
        this.isAutoScrolling = true;
        this.wsManager = null;
        this.container = null;
        this.logContainer = null;
        this.inputElement = null;
        this.commandHistory = [];
        this.historyIndex = -1;
        
        this.init();
    }
    
    // 初始化组件
    init() {
        this.createConsoleUI();
        this.setupEventListeners();
        this.connectWebSocket();
        this.loadRecentLogs();
    }
    
    // 创建控制台UI
    createConsoleUI() {
        const modal = document.createElement('div');
        modal.className = 'console-modal';
        modal.innerHTML = `
            <div class="console-modal-content">
                <div class="console-header">
                    <h3>服务器控制台 - ${this.serverId}</h3>
                    <div class="console-controls">
                        <button class="btn small" id="clearConsoleBtn">
                            <i class="mdi mdi-delete-sweep"></i>
                            <span>清空</span>
                        </button>
                        <button class="btn small" id="autoScrollBtn" data-active="${this.isAutoScrolling}">
                            <i class="mdi mdi-arrow-down"></i>
                            <span>自动滚动</span>
                        </button>
                        <button class="btn small" id="closeConsoleBtn">
                            <i class="mdi mdi-close"></i>
                            <span>关闭</span>
                        </button>
                    </div>
                </div>
                
                <div class="console-body">
                    <div class="console-logs" id="consoleLogs">
                        <div class="console-loading">
                            <i class="mdi mdi-loading mdi-spin"></i>
                            <span>加载日志中...</span>
                        </div>
                    </div>
                    
                    ${this.options.enableInput ? `
                    <div class="console-input">
                        <div class="input-group">
                            <span class="input-prefix">></span>
                            <input type="text" id="consoleInput" class="console-command-input" 
                                   placeholder="输入命令..." autocomplete="off">
                            <button class="btn primary" id="sendCommandBtn">
                                <i class="mdi mdi-send"></i>
                            </button>
                        </div>
                    </div>
                    ` : ''}
                </div>
            </div>
        `;
        
        // 添加样式
        if (!document.getElementById('console-styles')) {
            const styles = document.createElement('style');
            styles.id = 'console-styles';
            styles.textContent = `
                .console-modal {
                    position: fixed;
                    top: 0;
                    left: 0;
                    width: 100%;
                    height: 100%;
                    background: rgba(0, 0, 0, 0.8);
                    z-index: 10000;
                    display: flex;
                    align-items: center;
                    justify-content: center;
                }
                .console-modal-content {
                    background: #1e1e1e;
                    color: #ffffff;
                    width: 90%;
                    height: 80%;
                    border-radius: 8px;
                    display: flex;
                    flex-direction: column;
                    font-family: 'Consolas', 'Monaco', monospace;
                }
                .console-header {
                    padding: 16px;
                    border-bottom: 1px solid #333;
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                }
                .console-header h3 {
                    margin: 0;
                    color: #ffffff;
                }
                .console-controls {
                    display: flex;
                    gap: 8px;
                }
                .console-body {
                    flex: 1;
                    display: flex;
                    flex-direction: column;
                    overflow: hidden;
                }
                .console-logs {
                    flex: 1;
                    padding: 16px;
                    overflow-y: auto;
                    background: #000000;
                    font-size: 13px;
                    line-height: 1.4;
                }
                .console-log-line {
                    margin-bottom: 2px;
                    word-wrap: break-word;
                }
                .console-log-timestamp {
                    color: #666;
                    margin-right: 8px;
                }
                .console-log-level-INFO { color: #ffffff; }
                .console-log-level-WARN { color: #ffeb3b; }
                .console-log-level-ERROR { color: #f44336; }
                .console-log-level-DEBUG { color: #9e9e9e; }
                .console-input {
                    padding: 16px;
                    border-top: 1px solid #333;
                }
                .console-command-input {
                    background: #2d2d2d;
                    border: 1px solid #555;
                    color: #ffffff;
                    padding: 8px 12px;
                    border-radius: 4px;
                    font-family: inherit;
                    font-size: 13px;
                    flex: 1;
                }
                .console-command-input:focus {
                    outline: none;
                    border-color: #4CAF50;
                }
                .input-group {
                    display: flex;
                    align-items: center;
                    gap: 8px;
                }
                .input-prefix {
                    color: #4CAF50;
                    font-weight: bold;
                }
                .console-loading {
                    text-align: center;
                    padding: 20px;
                    color: #666;
                }
                .btn[data-active="true"] {
                    background: #4CAF50;
                    color: white;
                }
            `;
            document.head.appendChild(styles);
        }
        
        document.body.appendChild(modal);
        this.container = modal;
        this.logContainer = modal.querySelector('#consoleLogs');
        this.inputElement = modal.querySelector('#consoleInput');
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 关闭按钮
        const closeBtn = this.container.querySelector('#closeConsoleBtn');
        closeBtn.addEventListener('click', () => this.close());
        
        // 清空按钮
        const clearBtn = this.container.querySelector('#clearConsoleBtn');
        clearBtn.addEventListener('click', () => this.clearLogs());
        
        // 自动滚动按钮
        const autoScrollBtn = this.container.querySelector('#autoScrollBtn');
        autoScrollBtn.addEventListener('click', () => this.toggleAutoScroll());
        
        // 发送命令按钮
        const sendBtn = this.container.querySelector('#sendCommandBtn');
        if (sendBtn) {
            sendBtn.addEventListener('click', () => this.sendCommand());
        }
        
        // 输入框事件
        if (this.inputElement) {
            this.inputElement.addEventListener('keydown', (e) => {
                switch (e.key) {
                    case 'Enter':
                        this.sendCommand();
                        break;
                    case 'ArrowUp':
                        e.preventDefault();
                        this.navigateHistory(-1);
                        break;
                    case 'ArrowDown':
                        e.preventDefault();
                        this.navigateHistory(1);
                        break;
                }
            });
        }
        
        // 点击模态框外部关闭
        this.container.addEventListener('click', (e) => {
            if (e.target === this.container) {
                this.close();
            }
        });
        
        // ESC键关闭
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.close();
            }
        });
        
        // 滚动检测
        this.logContainer.addEventListener('scroll', () => {
            const { scrollTop, scrollHeight, clientHeight } = this.logContainer;
            const isAtBottom = scrollTop + clientHeight >= scrollHeight - 10;
            
            if (this.isAutoScrolling && !isAtBottom) {
                this.isAutoScrolling = false;
                this.updateAutoScrollButton();
            } else if (!this.isAutoScrolling && isAtBottom) {
                this.isAutoScrolling = true;
                this.updateAutoScrollButton();
            }
        });
    }
    
    // 连接WebSocket
    connectWebSocket() {
        this.wsManager = window.getWebSocketManager();
        if (this.wsManager) {
            // 订阅服务器日志
            this.wsManager.subscribeToLogs(this.serverId);
            
            // 监听日志消息
            this.wsManager.onMessage('log_message', (message) => {
                if (message.server_id === this.serverId) {
                    this.addLogLine(message.data, message.level || 'INFO', new Date(message.timestamp));
                }
            });
        }
    }
    
    // 加载最近的日志
    async loadRecentLogs() {
        try {
            const response = await fetch(`/api/servers/${this.serverId}/logs?lines=100`);
            const result = await response.json();
            
            if (response.ok && result.success) {
                const logs = result.data || [];
                this.logContainer.innerHTML = '';
                
                logs.forEach(log => {
                    this.addLogLine(log.message, log.level || 'INFO', new Date(log.timestamp));
                });
                
                if (logs.length === 0) {
                    this.logContainer.innerHTML = '<div class="console-loading">暂无日志</div>';
                }
            } else {
                this.logContainer.innerHTML = '<div class="console-loading">加载日志失败</div>';
            }
        } catch (error) {
            console.error('Failed to load logs:', error);
            this.logContainer.innerHTML = '<div class="console-loading">加载日志失败</div>';
        }
    }

    // 添加日志行
    addLogLine(message, level = 'INFO', timestamp = new Date()) {
        // 移除加载提示
        const loading = this.logContainer.querySelector('.console-loading');
        if (loading) {
            loading.remove();
        }

        const logLine = document.createElement('div');
        logLine.className = `console-log-line console-log-level-${level}`;

        let content = '';
        if (this.options.showTimestamp) {
            const timeStr = timestamp.toLocaleTimeString();
            content += `<span class="console-log-timestamp">[${timeStr}]</span>`;
        }

        // 处理ANSI颜色代码（简单实现）
        const processedMessage = this.processAnsiColors(message);
        content += processedMessage;

        logLine.innerHTML = content;
        this.logContainer.appendChild(logLine);

        // 限制日志行数
        this.logs.push({ message, level, timestamp });
        if (this.logs.length > this.options.maxLines) {
            this.logs.shift();
            const firstLine = this.logContainer.firstElementChild;
            if (firstLine && !firstLine.classList.contains('console-loading')) {
                firstLine.remove();
            }
        }

        // 自动滚动
        if (this.isAutoScrolling) {
            this.scrollToBottom();
        }
    }

    // 处理ANSI颜色代码
    processAnsiColors(text) {
        // 简单的ANSI颜色处理
        const ansiMap = {
            '\\u001b\\[31m': '<span style="color: #f44336;">', // 红色
            '\\u001b\\[32m': '<span style="color: #4caf50;">', // 绿色
            '\\u001b\\[33m': '<span style="color: #ffeb3b;">', // 黄色
            '\\u001b\\[34m': '<span style="color: #2196f3;">', // 蓝色
            '\\u001b\\[35m': '<span style="color: #9c27b0;">', // 紫色
            '\\u001b\\[36m': '<span style="color: #00bcd4;">', // 青色
            '\\u001b\\[37m': '<span style="color: #ffffff;">', // 白色
            '\\u001b\\[0m': '</span>', // 重置
        };

        let processed = text;
        for (const [ansi, html] of Object.entries(ansiMap)) {
            processed = processed.replace(new RegExp(ansi, 'g'), html);
        }

        return processed;
    }

    // 发送命令
    async sendCommand() {
        if (!this.inputElement) return;

        const command = this.inputElement.value.trim();
        if (!command) return;

        // 添加到历史记录
        this.commandHistory.push(command);
        this.historyIndex = this.commandHistory.length;

        // 显示发送的命令
        this.addLogLine(`> ${command}`, 'INFO');

        // 清空输入框
        this.inputElement.value = '';

        try {
            const response = await fetch(`/api/servers/${this.serverId}/command`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({ command })
            });

            const result = await response.json();

            if (!response.ok || !result.success) {
                const errorMessage = result.message || result.error || '命令发送失败';
                this.addLogLine(`错误: ${errorMessage}`, 'ERROR');
            }
        } catch (error) {
            console.error('Send command failed:', error);
            this.addLogLine('错误: 网络错误，命令发送失败', 'ERROR');
        }
    }

    // 导航命令历史
    navigateHistory(direction) {
        if (!this.inputElement || this.commandHistory.length === 0) return;

        this.historyIndex += direction;

        if (this.historyIndex < 0) {
            this.historyIndex = 0;
        } else if (this.historyIndex >= this.commandHistory.length) {
            this.historyIndex = this.commandHistory.length;
            this.inputElement.value = '';
            return;
        }

        this.inputElement.value = this.commandHistory[this.historyIndex] || '';
    }

    // 清空日志
    clearLogs() {
        this.logs = [];
        this.logContainer.innerHTML = '<div class="console-loading">日志已清空</div>';
    }

    // 切换自动滚动
    toggleAutoScroll() {
        this.isAutoScrolling = !this.isAutoScrolling;
        this.updateAutoScrollButton();

        if (this.isAutoScrolling) {
            this.scrollToBottom();
        }
    }

    // 更新自动滚动按钮状态
    updateAutoScrollButton() {
        const autoScrollBtn = this.container.querySelector('#autoScrollBtn');
        if (autoScrollBtn) {
            autoScrollBtn.setAttribute('data-active', this.isAutoScrolling);
        }
    }

    // 滚动到底部
    scrollToBottom() {
        this.logContainer.scrollTop = this.logContainer.scrollHeight;
    }

    // 关闭控制台
    close() {
        // 取消订阅
        if (this.wsManager) {
            this.wsManager.unsubscribeFromLogs(this.serverId);
        }

        // 移除DOM元素
        if (this.container) {
            this.container.remove();
        }
    }

    // 销毁组件
    destroy() {
        this.close();
    }
}

// 导出到全局作用域
if (typeof window !== 'undefined') {
    window.ServerConsoleComponent = ServerConsoleComponent;
}
