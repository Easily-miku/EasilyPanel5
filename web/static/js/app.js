// EasilyPanel5 v// EasilyPanel5 v1.1.0 主应用程序
class EasilyPanel5App {
    constructor() {
        this.wsManager = null;
        this.uiManager = null;
        this.servers = new Map();
        this.templates = new Map();
        this.groups = new Map();
        this.frpTunnels = new Map();
        this.systemStatus = {
            java: 'unknown',
            daemon: 'unknown',
            frp: 'unknown'
        };

        this.init();
    }

    async init() {
        console.log('EasilyPanel5 v1.1.0 应用程序启动中...');

        // 等待UI管理器初始化
        await this.waitForUIManager();

        // 初始化数据
        await this.loadInitialData();

        // 初始化WebSocket管理器
        console.log('初始化WebSocket管理器...');
        this.wsManager = initWebSocket();
        this.setupWebSocketHandlers();

        // 延迟启动WebSocket连接
        setTimeout(() => {
            this.startWebSocketIfNeeded();
        }, 2000);

        // 定期更新数据
        this.startDataRefresh();

        console.log('应用程序初始化完成');
    }

    // 等待UI管理器初始化
    async waitForUIManager() {
        return new Promise((resolve) => {
            const checkUIManager = () => {
                if (window.getUIManager && window.getUIManager()) {
                    this.uiManager = window.getUIManager();
                    console.log('UI管理器已连接');
                    resolve();
                } else {
                    setTimeout(checkUIManager, 100);
                }
            };
            checkUIManager();
        });
    }

    // 加载初始数据
    async loadInitialData() {
        console.log('加载初始数据...');

        try {
            // 并行加载所有数据
            const [
                javaInfo,
                serversData,
                templatesData,
                groupsData,
                daemonStatus,
                frpStatus
            ] = await Promise.allSettled([
                this.fetchJavaInfo(),
                this.fetchServers(),
                this.fetchTemplates(),
                this.fetchGroups(),
                this.fetchDaemonStatus(),
                this.fetchFRPStatus()
            ]);

            // 处理Java信息
            if (javaInfo.status === 'fulfilled') {
                this.updateJavaStatus(javaInfo.value);
            }

            // 处理服务器数据
            if (serversData.status === 'fulfilled') {
                this.updateServersData(serversData.value);
            }

            // 处理模板数据
            if (templatesData.status === 'fulfilled') {
                this.updateTemplatesData(templatesData.value);
            }

            // 处理分组数据
            if (groupsData.status === 'fulfilled') {
                this.updateGroupsData(groupsData.value);
            }

            // 处理守护状态
            if (daemonStatus.status === 'fulfilled') {
                this.updateDaemonStatus(daemonStatus.value);
            }

            // 处理FRP状态
            if (frpStatus.status === 'fulfilled') {
                this.updateFRPStatus(frpStatus.value);
            }

            // 更新仪表盘
            this.updateDashboard();

        } catch (error) {
            console.error('加载初始数据失败:', error);
            this.uiManager?.showNotification('加载失败', '初始数据加载失败，请刷新页面重试', 'error');
        }
    }

    // API调用方法
    async fetchJavaInfo() {
        const response = await fetch('/api/java/detect');
        if (!response.ok) throw new Error('Failed to fetch Java info');
        return await response.json();
    }

    async fetchServers() {
        const response = await fetch('/api/servers');
        if (!response.ok) throw new Error('Failed to fetch servers');
        return await response.json();
    }

    async fetchTemplates() {
        const response = await fetch('/api/templates');
        if (!response.ok) throw new Error('Failed to fetch templates');
        return await response.json();
    }

    async fetchGroups() {
        const response = await fetch('/api/groups');
        if (!response.ok) throw new Error('Failed to fetch groups');
        return await response.json();
    }

    async fetchDaemonStatus() {
        const response = await fetch('/api/daemon/status');
        if (!response.ok) throw new Error('Failed to fetch daemon status');
        return await response.json();
    }

    async fetchFRPStatus() {
        const response = await fetch('/api/frp/status');
        if (!response.ok) throw new Error('Failed to fetch FRP status');
        return await response.json();
    }

    // 数据更新方法
    updateJavaStatus(javaInfo) {
        this.systemStatus.java = javaInfo.available ? 'available' : 'unavailable';

        const javaStatusEl = document.getElementById('javaStatus');
        if (javaStatusEl) {
            if (javaInfo.available) {
                javaStatusEl.textContent = `Java ${javaInfo.version}`;
                javaStatusEl.className = 'status-value success';
            } else {
                javaStatusEl.textContent = '未安装';
                javaStatusEl.className = 'status-value error';
            }
        }
    }

    updateServersData(servers) {
        this.servers.clear();
        let runningCount = 0;
        let stoppedCount = 0;

        servers.forEach(server => {
            this.servers.set(server.id, server);
            if (server.status === 'running') {
                runningCount++;
            } else {
                stoppedCount++;
            }
        });

        // 更新仪表盘统计
        const totalServersEl = document.getElementById('totalServers');
        const runningServersEl = document.getElementById('runningServers');
        const stoppedServersEl = document.getElementById('stoppedServers');

        if (totalServersEl) totalServersEl.textContent = servers.length;
        if (runningServersEl) runningServersEl.textContent = runningCount;
        if (stoppedServersEl) stoppedServersEl.textContent = stoppedCount;
    }

    updateTemplatesData(templates) {
        this.templates.clear();
        templates.forEach(template => {
            this.templates.set(template.id, template);
        });
        console.log(`加载了 ${templates.length} 个模板`);
    }

    updateGroupsData(groups) {
        this.groups.clear();
        groups.forEach(group => {
            this.groups.set(group.id, group);
        });
        console.log(`加载了 ${groups.length} 个分组`);
    }

    updateDaemonStatus(status) {
        this.systemStatus.daemon = status.enabled ? 'enabled' : 'disabled';

        const daemonStatusEl = document.getElementById('daemonStatus');
        if (daemonStatusEl) {
            if (status.enabled) {
                daemonStatusEl.textContent = '已启用';
                daemonStatusEl.className = 'status-value success';
            } else {
                daemonStatusEl.textContent = '已禁用';
                daemonStatusEl.className = 'status-value warning';
            }
        }
    }

    updateFRPStatus(status) {
        this.systemStatus.frp = status.enabled ? 'enabled' : 'disabled';

        const frpStatusEl = document.getElementById('frpStatus');
        if (frpStatusEl) {
            if (status.enabled) {
                frpStatusEl.textContent = status.is_running ? '运行中' : '已配置';
                frpStatusEl.className = status.is_running ? 'status-value success' : 'status-value warning';
            } else {
                frpStatusEl.textContent = '未启用';
                frpStatusEl.className = 'status-value secondary';
            }
        }
    }

    updateDashboard() {
        // 更新最近活动
        this.updateRecentActivity();
    }

    updateRecentActivity() {
        const activityList = document.getElementById('activityList');
        if (!activityList) return;

        const activities = [
            {
                icon: 'mdi-information',
                title: '系统启动',
                time: '刚刚',
                type: 'info'
            },
            {
                icon: 'mdi-server',
                title: `加载了 ${this.servers.size} 个服务器`,
                time: '刚刚',
                type: 'success'
            },
            {
                icon: 'mdi-file-document-multiple',
                title: `加载了 ${this.templates.size} 个模板`,
                time: '刚刚',
                type: 'info'
            }
        ];

        activityList.innerHTML = activities.map(activity => `
            <div class="activity-item">
                <div class="activity-icon ${activity.type}">
                    <i class="mdi ${activity.icon}"></i>
                </div>
                <div class="activity-content">
                    <div class="activity-title">${activity.title}</div>
                    <div class="activity-time">${activity.time}</div>
                </div>
            </div>
        `).join('');
    }

    // WebSocket处理
    setupWebSocketHandlers() {
        if (!this.wsManager) return;

        this.wsManager.on('connected', (data) => {
            console.log('WebSocket连接已建立');
            this.updateConnectionStatus('connected');

            // 只在重连成功时显示通知
            if (data && data.reconnected) {
                this.uiManager?.showNotification('连接恢复', 'WebSocket连接已恢复', 'success');
            }
        });

        this.wsManager.on('disconnected', (data) => {
            console.log('WebSocket连接已断开:', data);
            this.updateConnectionStatus('disconnected');

            // 只在意外断开且之前已连接时显示警告
            if (data && data.wasConnected && data.code !== 1000 && data.code !== 1001) {
                this.uiManager?.showNotification('连接断开', '正在尝试重新连接...', 'warning');
            }
        });

        this.wsManager.on('reconnect_failed', (data) => {
            this.uiManager?.showNotification('连接失败', `连接失败 (${data.attempts}次重试)，请检查网络或刷新页面`, 'error');
        });

        // 监听服务器事件
        this.wsManager.on('server_status', (message) => {
            this.handleServerStatusUpdate(message);
        });

        // 监听日志消息
        this.wsManager.on('log_message', (message) => {
            this.handleLogMessage(message);
        });
    }

    updateConnectionStatus(status) {
        const connectionStatus = document.getElementById('connectionStatus');
        if (!connectionStatus) return;

        const statusIcon = connectionStatus.querySelector('i');
        const statusText = connectionStatus.querySelector('span');

        connectionStatus.className = `connection-status ${status}`;

        switch (status) {
            case 'connected':
                statusText.textContent = '已连接';
                break;
            case 'disconnected':
                statusText.textContent = '已断开';
                break;
            case 'connecting':
                statusText.textContent = '连接中...';
                break;
        }
    }

    handleServerStatusUpdate(message) {
        if (message.server_id && message.data) {
            const server = this.servers.get(message.server_id);
            if (server) {
                Object.assign(server, message.data);
                this.updateServersData(Array.from(this.servers.values()));
            }
        }
    }

    handleLogMessage(message) {
        // TODO: 处理日志消息
        console.log('收到日志消息:', message);
    }

    // 启动WebSocket连接（如果需要）
    startWebSocketIfNeeded() {
        if (!this.wsManager) {
            console.log('初始化WebSocket管理器...');
            this.wsManager = initWebSocket();
            this.setupWebSocketHandlers();
        }

        if (this.wsManager.getState() === WS_STATES.DISCONNECTED) {
            console.log('启动WebSocket连接');
            this.wsManager.connect();
        }
    }

    // 定期数据刷新
    startDataRefresh() {
        // 每30秒刷新一次数据
        setInterval(() => {
            this.refreshData();
        }, 30000);
    }

    async refreshData() {
        try {
            // 只刷新关键数据
            const [serversData, daemonStatus, frpStatus] = await Promise.allSettled([
                this.fetchServers(),
                this.fetchDaemonStatus(),
                this.fetchFRPStatus()
            ]);

            if (serversData.status === 'fulfilled') {
                this.updateServersData(serversData.value);
            }

            if (daemonStatus.status === 'fulfilled') {
                this.updateDaemonStatus(daemonStatus.value);
            }

            if (frpStatus.status === 'fulfilled') {
                this.updateFRPStatus(frpStatus.value);
            }

        } catch (error) {
            console.error('数据刷新失败:', error);
        }
    }
}

// 全局应用实例
let app;

// 页面加载完成后初始化应用
document.addEventListener('DOMContentLoaded', () => {
    app = new EasilyPanel5App();
});

// 导出到全局作用域
window.EasilyPanel5App = EasilyPanel5App;
window.getApp = () => app;.1.0 主应用程序
class EasilyPanel5App {
    constructor() {
        this.wsManager = null;
        this.uiManager = null;
        this.servers = new Map();
        this.templates = new Map();
        this.groups = new Map();
        this.frpTunnels = new Map();
        this.systemStatus = {
            java: 'unknown',
            daemon: 'unknown',
            frp: 'unknown'
        };
        
        this.init();
    }

    async init() {
        console.log('EasilyPanel5 v1.1.0 应用程序启动中...');
        
        // 等待UI管理器初始化
        await this.waitForUIManager();
        
        // 初始化数据
        await this.loadInitialData();
        
        // 初始化WebSocket管理器
        console.log('初始化WebSocket管理器...');
        this.wsManager = initWebSocket();
        this.setupWebSocketHandlers();
        
        // 延迟启动WebSocket连接
        setTimeout(() => {
            this.startWebSocketIfNeeded();
        }, 2000);
        
        // 定期更新数据
        this.startDataRefresh();
        
        console.log('应用程序初始化完成');
    }
    
    // 等待UI管理器初始化
    async waitForUIManager() {
        return new Promise((resolve) => {
            const checkUIManager = () => {
                if (window.getUIManager && window.getUIManager()) {
                    this.uiManager = window.getUIManager();
                    console.log('UI管理器已连接');
                    resolve();
                } else {
                    setTimeout(checkUIManager, 100);
                }
            };
            checkUIManager();
        });
    }
    
    // 加载初始数据
    async loadInitialData() {
        console.log('加载初始数据...');
        
        try {
            // 并行加载所有数据
            const [
                javaInfo,
                serversData,
                templatesData,
                groupsData,
                daemonStatus,
                frpStatus
            ] = await Promise.allSettled([
                this.fetchJavaInfo(),
                this.fetchServers(),
                this.fetchTemplates(),
                this.fetchGroups(),
                this.fetchDaemonStatus(),
                this.fetchFRPStatus()
            ]);
            
            // 处理Java信息
            if (javaInfo.status === 'fulfilled') {
                this.updateJavaStatus(javaInfo.value);
            }
            
            // 处理服务器数据
            if (serversData.status === 'fulfilled') {
                this.updateServersData(serversData.value);
            }
            
            // 处理模板数据
            if (templatesData.status === 'fulfilled') {
                this.updateTemplatesData(templatesData.value);
            }
            
            // 处理分组数据
            if (groupsData.status === 'fulfilled') {
                this.updateGroupsData(groupsData.value);
            }
            
            // 处理守护状态
            if (daemonStatus.status === 'fulfilled') {
                this.updateDaemonStatus(daemonStatus.value);
            }
            
            // 处理FRP状态
            if (frpStatus.status === 'fulfilled') {
                this.updateFRPStatus(frpStatus.value);
            }
            
            // 更新仪表盘
            this.updateDashboard();
            
        } catch (error) {
            console.error('加载初始数据失败:', error);
            this.uiManager?.showNotification('加载失败', '初始数据加载失败，请刷新页面重试', 'error');
        }
    }
    
    // API调用方法
    async fetchJavaInfo() {
        const response = await fetch('/api/java/detect');
        if (!response.ok) throw new Error('Failed to fetch Java info');
        return await response.json();
    }
    
    async fetchServers() {
        const response = await fetch('/api/servers');
        if (!response.ok) throw new Error('Failed to fetch servers');
        return await response.json();
    }
    
    async fetchTemplates() {
        const response = await fetch('/api/templates');
        if (!response.ok) throw new Error('Failed to fetch templates');
        return await response.json();
    }
    
    async fetchGroups() {
        const response = await fetch('/api/groups');
        if (!response.ok) throw new Error('Failed to fetch groups');
        return await response.json();
    }
    
    async fetchDaemonStatus() {
        const response = await fetch('/api/daemon/status');
        if (!response.ok) throw new Error('Failed to fetch daemon status');
        return await response.json();
    }
    
    async fetchFRPStatus() {
        const response = await fetch('/api/frp/status');
        if (!response.ok) throw new Error('Failed to fetch FRP status');
        return await response.json();
    }
    
    // 数据更新方法
    updateJavaStatus(javaInfo) {
        this.systemStatus.java = javaInfo.available ? 'available' : 'unavailable';
        
        const javaStatusEl = document.getElementById('javaStatus');
        if (javaStatusEl) {
            if (javaInfo.available) {
                javaStatusEl.textContent = `Java ${javaInfo.version}`;
                javaStatusEl.className = 'status-value success';
            } else {
                javaStatusEl.textContent = '未安装';
                javaStatusEl.className = 'status-value error';
            }
        }
    }
    
    updateServersData(servers) {
        this.servers.clear();
        let runningCount = 0;
        let stoppedCount = 0;
        
        servers.forEach(server => {
            this.servers.set(server.id, server);
            if (server.status === 'running') {
                runningCount++;
            } else {
                stoppedCount++;
            }
        });
        
        // 更新仪表盘统计
        const totalServersEl = document.getElementById('totalServers');
        const runningServersEl = document.getElementById('runningServers');
        const stoppedServersEl = document.getElementById('stoppedServers');
        
        if (totalServersEl) totalServersEl.textContent = servers.length;
        if (runningServersEl) runningServersEl.textContent = runningCount;
        if (stoppedServersEl) stoppedServersEl.textContent = stoppedCount;
    }
    
    updateTemplatesData(templates) {
        this.templates.clear();
        templates.forEach(template => {
            this.templates.set(template.id, template);
        });
        console.log(`加载了 ${templates.length} 个模板`);
    }
    
    updateGroupsData(groups) {
        this.groups.clear();
        groups.forEach(group => {
            this.groups.set(group.id, group);
        });
        console.log(`加载了 ${groups.length} 个分组`);
    }
    
    updateDaemonStatus(status) {
        this.systemStatus.daemon = status.enabled ? 'enabled' : 'disabled';
        
        const daemonStatusEl = document.getElementById('daemonStatus');
        if (daemonStatusEl) {
            if (status.enabled) {
                daemonStatusEl.textContent = '已启用';
                daemonStatusEl.className = 'status-value success';
            } else {
                daemonStatusEl.textContent = '已禁用';
                daemonStatusEl.className = 'status-value warning';
            }
        }
    }
    
    updateFRPStatus(status) {
        this.systemStatus.frp = status.enabled ? 'enabled' : 'disabled';
        
        const frpStatusEl = document.getElementById('frpStatus');
        if (frpStatusEl) {
            if (status.enabled) {
                frpStatusEl.textContent = status.is_running ? '运行中' : '已配置';
                frpStatusEl.className = status.is_running ? 'status-value success' : 'status-value warning';
            } else {
                frpStatusEl.textContent = '未启用';
                frpStatusEl.className = 'status-value secondary';
            }
        }
    }
    
    updateDashboard() {
        // 更新最近活动
        this.updateRecentActivity();
    }
    
    updateRecentActivity() {
        const activityList = document.getElementById('activityList');
        if (!activityList) return;
        
        const activities = [
            {
                icon: 'mdi-information',
                title: '系统启动',
                time: '刚刚',
                type: 'info'
            },
            {
                icon: 'mdi-server',
                title: `加载了 ${this.servers.size} 个服务器`,
                time: '刚刚',
                type: 'success'
            },
            {
                icon: 'mdi-file-document-multiple',
                title: `加载了 ${this.templates.size} 个模板`,
                time: '刚刚',
                type: 'info'
            }
        ];
        
        activityList.innerHTML = activities.map(activity => `
            <div class="activity-item">
                <div class="activity-icon ${activity.type}">
                    <i class="mdi ${activity.icon}"></i>
                </div>
                <div class="activity-content">
                    <div class="activity-title">${activity.title}</div>
                    <div class="activity-time">${activity.time}</div>
                </div>
            </div>
        `).join('');
    }
    
    // WebSocket处理
    setupWebSocketHandlers() {
        if (!this.wsManager) return;
        
        this.wsManager.on('connected', (data) => {
            console.log('WebSocket连接已建立');
            this.updateConnectionStatus('connected');
            
            // 只在重连成功时显示通知
            if (data && data.reconnected) {
                this.uiManager?.showNotification('连接恢复', 'WebSocket连接已恢复', 'success');
            }
        });
        
        this.wsManager.on('disconnected', (data) => {
            console.log('WebSocket连接已断开:', data);
            this.updateConnectionStatus('disconnected');
            
            // 只在意外断开且之前已连接时显示警告
            if (data && data.wasConnected && data.code !== 1000 && data.code !== 1001) {
                this.uiManager?.showNotification('连接断开', '正在尝试重新连接...', 'warning');
            }
        });
        
        this.wsManager.on('reconnect_failed', (data) => {
            this.uiManager?.showNotification('连接失败', `连接失败 (${data.attempts}次重试)，请检查网络或刷新页面`, 'error');
        });
        
        // 监听服务器事件
        this.wsManager.on('server_status', (message) => {
            this.handleServerStatusUpdate(message);
        });
        
        // 监听日志消息
        this.wsManager.on('log_message', (message) => {
            this.handleLogMessage(message);
        });
    }
    
    updateConnectionStatus(status) {
        const connectionStatus = document.getElementById('connectionStatus');
        if (!connectionStatus) return;
        
        const statusIcon = connectionStatus.querySelector('i');
        const statusText = connectionStatus.querySelector('span');
        
        connectionStatus.className = `connection-status ${status}`;
        
        switch (status) {
            case 'connected':
                statusText.textContent = '已连接';
                break;
            case 'disconnected':
                statusText.textContent = '已断开';
                break;
            case 'connecting':
                statusText.textContent = '连接中...';
                break;
        }
    }
    
    handleServerStatusUpdate(message) {
        if (message.server_id && message.data) {
            const server = this.servers.get(message.server_id);
            if (server) {
                Object.assign(server, message.data);
                this.updateServersData(Array.from(this.servers.values()));
            }
        }
    }
    
    handleLogMessage(message) {
        // TODO: 处理日志消息
        console.log('收到日志消息:', message);
    }
    
    // 启动WebSocket连接（如果需要）
    startWebSocketIfNeeded() {
        if (!this.wsManager) {
            console.log('初始化WebSocket管理器...');
            this.wsManager = initWebSocket();
            this.setupWebSocketHandlers();
        }
        
        if (this.wsManager.getState() === WS_STATES.DISCONNECTED) {
            console.log('启动WebSocket连接');
            this.wsManager.connect();
        }
    }
    
    // 定期数据刷新
    startDataRefresh() {
        // 每30秒刷新一次数据
        setInterval(() => {
            this.refreshData();
        }, 30000);
    }
    
    async refreshData() {
        try {
            // 只刷新关键数据
            const [serversData, daemonStatus, frpStatus] = await Promise.allSettled([
                this.fetchServers(),
                this.fetchDaemonStatus(),
                this.fetchFRPStatus()
            ]);
            
            if (serversData.status === 'fulfilled') {
                this.updateServersData(serversData.value);
            }
            
            if (daemonStatus.status === 'fulfilled') {
                this.updateDaemonStatus(daemonStatus.value);
            }
            
            if (frpStatus.status === 'fulfilled') {
                this.updateFRPStatus(frpStatus.value);
            }
            
        } catch (error) {
            console.error('数据刷新失败:', error);
        }
    }
}

// 全局应用实例
let app;

// 页面加载完成后初始化应用
document.addEventListener('DOMContentLoaded', () => {
    app = new EasilyPanel5App();
});

// 导出到全局作用域
window.EasilyPanel5App = EasilyPanel5App;
window.getApp = () => app;
