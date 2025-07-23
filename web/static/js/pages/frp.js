// FRP内网穿透页面管理器
class FRPPageManager {
    constructor() {
        this.frpConfig = null;
        this.tunnels = new Map();
        this.selectedTunnels = new Set();
        this.searchQuery = '';
        this.currentTab = 'tunnels'; // tunnels, config, logs
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadData();
    }
    
    // 渲染页面
    renderPage() {
        const frpPage = document.getElementById('frp-page');
        if (!frpPage) return;
        
        frpPage.innerHTML = `
            <div class="page-header">
                <h2>内网穿透</h2>
                <p>管理FRP隧道和配置</p>
            </div>
            
            <div class="frp-container">
                ${this.renderTabs()}
                ${this.renderTabContent()}
            </div>
        `;
    }
    
    // 渲染标签页
    renderTabs() {
        return `
            <div class="frp-tabs">
                <button class="tab-btn ${this.currentTab === 'tunnels' ? 'active' : ''}" data-tab="tunnels">
                    <i class="mdi mdi-tunnel"></i>
                    <span>隧道管理</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'config' ? 'active' : ''}" data-tab="config">
                    <i class="mdi mdi-cog"></i>
                    <span>FRP配置</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'logs' ? 'active' : ''}" data-tab="logs">
                    <i class="mdi mdi-text-box"></i>
                    <span>运行日志</span>
                </button>
            </div>
        `;
    }
    
    // 渲染标签页内容
    renderTabContent() {
        return `
            <div class="tab-content">
                <div class="tab-pane ${this.currentTab === 'tunnels' ? 'active' : ''}" id="tunnels-tab">
                    ${this.renderTunnelsTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'config' ? 'active' : ''}" id="config-tab">
                    ${this.renderConfigTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'logs' ? 'active' : ''}" id="logs-tab">
                    ${this.renderLogsTab()}
                </div>
            </div>
        `;
    }
    
    // 渲染隧道管理标签页
    renderTunnelsTab() {
        return `
            <div class="tunnels-toolbar">
                <div class="toolbar-left">
                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="tunnelSearch" placeholder="搜索隧道..." value="${this.searchQuery}">
                    </div>
                    
                    <div class="status-filter">
                        <select id="statusFilter" class="filter-select">
                            <option value="all">全部状态</option>
                            <option value="online">在线</option>
                            <option value="offline">离线</option>
                            <option value="error">错误</option>
                        </select>
                    </div>
                </div>
                
                <div class="toolbar-right">
                    <div class="batch-actions" id="batchActions" style="display: none;">
                        <span class="selected-count" id="selectedCount">已选择 0 个隧道</span>
                        <button class="btn" id="batchStartBtn">
                            <i class="mdi mdi-play"></i>
                            <span>批量启动</span>
                        </button>
                        <button class="btn" id="batchStopBtn">
                            <i class="mdi mdi-stop"></i>
                            <span>批量停止</span>
                        </button>
                        <button class="btn warning" id="batchDeleteBtn">
                            <i class="mdi mdi-delete"></i>
                            <span>批量删除</span>
                        </button>
                    </div>
                    
                    <button class="btn primary" id="createTunnelBtn">
                        <i class="mdi mdi-plus"></i>
                        <span>创建隧道</span>
                    </button>
                </div>
            </div>
            
            <div class="tunnels-list-container">
                <div class="tunnels-list" id="tunnelsList">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载隧道列表...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染配置标签页
    renderConfigTab() {
        return `
            <div class="config-container">
                <div class="config-status">
                    <div class="status-card">
                        <div class="status-header">
                            <h3>FRP服务状态</h3>
                            <div class="status-indicator" id="frpStatus">
                                <i class="mdi mdi-loading mdi-spin"></i>
                                <span>检查中...</span>
                            </div>
                        </div>
                        <div class="status-actions">
                            <button class="btn success" id="startFrpBtn">
                                <i class="mdi mdi-play"></i>
                                <span>启动FRP</span>
                            </button>
                            <button class="btn warning" id="stopFrpBtn">
                                <i class="mdi mdi-stop"></i>
                                <span>停止FRP</span>
                            </button>
                            <button class="btn" id="restartFrpBtn">
                                <i class="mdi mdi-restart"></i>
                                <span>重启FRP</span>
                            </button>
                        </div>
                    </div>
                </div>
                
                <div class="config-form">
                    <div class="config-section">
                        <h3>基础配置</h3>
                        <div class="form-grid">
                            <div class="form-group">
                                <label for="serverAddr">服务器地址</label>
                                <input type="text" id="serverAddr" class="form-input" placeholder="frp.example.com">
                            </div>
                            <div class="form-group">
                                <label for="serverPort">服务器端口</label>
                                <input type="number" id="serverPort" class="form-input" placeholder="7000" value="7000">
                            </div>
                            <div class="form-group">
                                <label for="authToken">认证令牌</label>
                                <input type="password" id="authToken" class="form-input" placeholder="输入认证令牌">
                            </div>
                            <div class="form-group">
                                <label for="user">用户名</label>
                                <input type="text" id="user" class="form-input" placeholder="用户名（可选）">
                            </div>
                        </div>
                    </div>
                    
                    <div class="config-section">
                        <h3>高级配置</h3>
                        <div class="form-grid">
                            <div class="form-group">
                                <label for="protocol">协议类型</label>
                                <select id="protocol" class="form-input">
                                    <option value="tcp">TCP</option>
                                    <option value="kcp">KCP</option>
                                    <option value="websocket">WebSocket</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label for="logLevel">日志级别</label>
                                <select id="logLevel" class="form-input">
                                    <option value="info">Info</option>
                                    <option value="warn">Warn</option>
                                    <option value="error">Error</option>
                                    <option value="debug">Debug</option>
                                </select>
                            </div>
                            <div class="form-group">
                                <label for="heartbeatInterval">心跳间隔(秒)</label>
                                <input type="number" id="heartbeatInterval" class="form-input" value="30">
                            </div>
                            <div class="form-group">
                                <label for="heartbeatTimeout">心跳超时(秒)</label>
                                <input type="number" id="heartbeatTimeout" class="form-input" value="90">
                            </div>
                        </div>
                    </div>
                    
                    <div class="config-actions">
                        <button class="btn" id="testConfigBtn">
                            <i class="mdi mdi-test-tube"></i>
                            <span>测试连接</span>
                        </button>
                        <button class="btn primary" id="saveConfigBtn">
                            <i class="mdi mdi-content-save"></i>
                            <span>保存配置</span>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染日志标签页
    renderLogsTab() {
        return `
            <div class="logs-container">
                <div class="logs-toolbar">
                    <div class="toolbar-left">
                        <select id="logLevel" class="filter-select">
                            <option value="all">全部级别</option>
                            <option value="debug">Debug</option>
                            <option value="info">Info</option>
                            <option value="warn">Warn</option>
                            <option value="error">Error</option>
                        </select>
                        
                        <div class="log-search">
                            <i class="mdi mdi-magnify"></i>
                            <input type="text" id="logSearch" placeholder="搜索日志...">
                        </div>
                    </div>
                    
                    <div class="toolbar-right">
                        <button class="btn" id="clearLogsBtn">
                            <i class="mdi mdi-delete-sweep"></i>
                            <span>清空日志</span>
                        </button>
                        <button class="btn" id="downloadLogsBtn">
                            <i class="mdi mdi-download"></i>
                            <span>下载日志</span>
                        </button>
                        <button class="btn" id="refreshLogsBtn">
                            <i class="mdi mdi-refresh"></i>
                            <span>刷新</span>
                        </button>
                    </div>
                </div>
                
                <div class="logs-viewer" id="logsViewer">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载日志...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 标签页切换
        const tabBtns = document.querySelectorAll('.tab-btn');
        tabBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.currentTarget.getAttribute('data-tab');
                this.switchTab(tab);
            });
        });
        
        // 搜索框
        const searchInput = document.getElementById('tunnelSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderTunnels();
            });
        }
        
        // 创建隧道按钮
        const createBtn = document.getElementById('createTunnelBtn');
        if (createBtn) {
            createBtn.addEventListener('click', () => this.showCreateTunnelDialog());
        }
        
        // FRP控制按钮
        const startFrpBtn = document.getElementById('startFrpBtn');
        const stopFrpBtn = document.getElementById('stopFrpBtn');
        const restartFrpBtn = document.getElementById('restartFrpBtn');
        
        if (startFrpBtn) {
            startFrpBtn.addEventListener('click', () => this.startFRP());
        }
        if (stopFrpBtn) {
            stopFrpBtn.addEventListener('click', () => this.stopFRP());
        }
        if (restartFrpBtn) {
            restartFrpBtn.addEventListener('click', () => this.restartFRP());
        }
        
        // 配置按钮
        const testConfigBtn = document.getElementById('testConfigBtn');
        const saveConfigBtn = document.getElementById('saveConfigBtn');
        
        if (testConfigBtn) {
            testConfigBtn.addEventListener('click', () => this.testFRPConfig());
        }
        if (saveConfigBtn) {
            saveConfigBtn.addEventListener('click', () => this.saveFRPConfig());
        }
        
        // 日志按钮
        const clearLogsBtn = document.getElementById('clearLogsBtn');
        const downloadLogsBtn = document.getElementById('downloadLogsBtn');
        const refreshLogsBtn = document.getElementById('refreshLogsBtn');
        
        if (clearLogsBtn) {
            clearLogsBtn.addEventListener('click', () => this.clearLogs());
        }
        if (downloadLogsBtn) {
            downloadLogsBtn.addEventListener('click', () => this.downloadLogs());
        }
        if (refreshLogsBtn) {
            refreshLogsBtn.addEventListener('click', () => this.loadLogs());
        }
    }
    
    // 切换标签页
    switchTab(tab) {
        this.currentTab = tab;
        
        // 更新标签按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tab}"]`).classList.add('active');
        
        // 更新标签页内容
        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.remove('active');
        });
        document.getElementById(`${tab}-tab`).classList.add('active');
        
        // 加载对应的数据
        switch (tab) {
            case 'tunnels':
                this.loadTunnels();
                break;
            case 'config':
                this.loadFRPConfig();
                break;
            case 'logs':
                this.loadLogs();
                break;
        }
    }
    
    // 加载数据
    async loadData() {
        await this.loadTunnels();
        await this.loadFRPConfig();
        await this.loadFRPStatus();
    }
    
    // 加载隧道列表
    async loadTunnels() {
        try {
            const response = await fetch('/api/frp/tunnels');
            if (response.ok) {
                const tunnels = await response.json();
                this.tunnels.clear();
                
                tunnels.forEach(tunnel => {
                    this.tunnels.set(tunnel.id, tunnel);
                });
                
                this.filterAndRenderTunnels();
            } else {
                this.showTunnelsError('加载隧道列表失败');
            }
        } catch (error) {
            console.error('Failed to load tunnels:', error);
            this.showTunnelsError('网络错误，请重试');
        }
    }
    
    // 过滤和渲染隧道
    filterAndRenderTunnels() {
        const tunnelsList = document.getElementById('tunnelsList');
        if (!tunnelsList) return;
        
        // 获取过滤后的隧道
        let filteredTunnels = Array.from(this.tunnels.values());
        
        // 应用搜索过滤
        if (this.searchQuery) {
            const query = this.searchQuery.toLowerCase();
            filteredTunnels = filteredTunnels.filter(tunnel => 
                tunnel.name.toLowerCase().includes(query) ||
                tunnel.local_addr?.toLowerCase().includes(query) ||
                tunnel.remote_addr?.toLowerCase().includes(query)
            );
        }
        
        // 渲染隧道卡片
        if (filteredTunnels.length === 0) {
            tunnelsList.innerHTML = this.renderTunnelsEmptyState();
        } else {
            tunnelsList.innerHTML = filteredTunnels.map(tunnel => 
                this.renderTunnelCard(tunnel)
            ).join('');
        }
        
        // 重新绑定事件
        this.bindTunnelEvents();
    }
    
    // 渲染隧道空状态
    renderTunnelsEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-tunnel-outline"></i>
                <h3>没有找到隧道</h3>
                <p>您还没有创建任何隧道，或者搜索条件没有匹配的结果。</p>
                <button class="btn primary" onclick="window.getFRPPageManager().showCreateTunnelDialog()">
                    <i class="mdi mdi-plus"></i>
                    <span>创建第一个隧道</span>
                </button>
            </div>
        `;
    }
    
    // 渲染隧道卡片
    renderTunnelCard(tunnel) {
        const statusIcons = {
            'online': 'mdi-check-circle',
            'offline': 'mdi-close-circle',
            'connecting': 'mdi-loading mdi-spin',
            'error': 'mdi-alert-circle'
        };
        
        const statusColors = {
            'online': 'success',
            'offline': 'secondary',
            'connecting': 'warning',
            'error': 'error'
        };
        
        const statusTexts = {
            'online': '在线',
            'offline': '离线',
            'connecting': '连接中',
            'error': '错误'
        };
        
        return `
            <div class="tunnel-card" data-tunnel-id="${tunnel.id}">
                <div class="tunnel-card-header">
                    <div class="tunnel-select">
                        <input type="checkbox" class="tunnel-checkbox" data-tunnel-id="${tunnel.id}">
                    </div>
                    <div class="tunnel-info">
                        <h3 class="tunnel-name">${tunnel.name}</h3>
                        <p class="tunnel-type">${tunnel.type.toUpperCase()} 隧道</p>
                    </div>
                    <div class="tunnel-status">
                        <span class="status-badge ${statusColors[tunnel.status]}">
                            <i class="mdi ${statusIcons[tunnel.status]}"></i>
                            <span>${statusTexts[tunnel.status]}</span>
                        </span>
                    </div>
                </div>
                
                <div class="tunnel-card-body">
                    <div class="tunnel-details">
                        <div class="detail-row">
                            <span class="detail-label">本地地址</span>
                            <span class="detail-value">${tunnel.local_ip}:${tunnel.local_port}</span>
                        </div>
                        <div class="detail-row">
                            <span class="detail-label">远程地址</span>
                            <span class="detail-value">${tunnel.remote_port ? `${tunnel.subdomain || 'auto'}.${tunnel.custom_domain || 'frp.example.com'}:${tunnel.remote_port}` : tunnel.custom_domain || '自动分配'}</span>
                        </div>
                        <div class="detail-row">
                            <span class="detail-label">协议类型</span>
                            <span class="detail-value">${tunnel.type}</span>
                        </div>
                        <div class="detail-row">
                            <span class="detail-label">创建时间</span>
                            <span class="detail-value">${new Date(tunnel.created_at).toLocaleDateString()}</span>
                        </div>
                    </div>
                </div>
                
                <div class="tunnel-card-footer">
                    <div class="tunnel-actions">
                        ${tunnel.status === 'offline' ? 
                            `<button class="btn success tunnel-action-btn" data-action="start" data-tunnel-id="${tunnel.id}">
                                <i class="mdi mdi-play"></i>
                                <span>启动</span>
                            </button>` :
                            `<button class="btn warning tunnel-action-btn" data-action="stop" data-tunnel-id="${tunnel.id}">
                                <i class="mdi mdi-stop"></i>
                                <span>停止</span>
                            </button>`
                        }
                        
                        <button class="btn tunnel-action-btn" data-action="restart" data-tunnel-id="${tunnel.id}">
                            <i class="mdi mdi-restart"></i>
                            <span>重启</span>
                        </button>
                        
                        ${tunnel.status === 'online' && tunnel.type === 'http' ? 
                            `<button class="btn tunnel-action-btn" data-action="open" data-tunnel-id="${tunnel.id}">
                                <i class="mdi mdi-open-in-new"></i>
                                <span>访问</span>
                            </button>` : ''
                        }
                        
                        <div class="dropdown">
                            <button class="btn tunnel-action-btn dropdown-toggle" data-tunnel-id="${tunnel.id}">
                                <i class="mdi mdi-dots-vertical"></i>
                            </button>
                            <div class="dropdown-menu">
                                <a href="#" class="dropdown-item" data-action="edit" data-tunnel-id="${tunnel.id}">
                                    <i class="mdi mdi-pencil"></i>
                                    <span>编辑隧道</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="copy" data-tunnel-id="${tunnel.id}">
                                    <i class="mdi mdi-content-copy"></i>
                                    <span>复制配置</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="logs" data-tunnel-id="${tunnel.id}">
                                    <i class="mdi mdi-text-box"></i>
                                    <span>查看日志</span>
                                </a>
                                <div class="dropdown-divider"></div>
                                <a href="#" class="dropdown-item danger" data-action="delete" data-tunnel-id="${tunnel.id}">
                                    <i class="mdi mdi-delete"></i>
                                    <span>删除隧道</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 绑定隧道事件
    bindTunnelEvents() {
        // 复选框事件
        const checkboxes = document.querySelectorAll('.tunnel-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                const tunnelId = e.target.getAttribute('data-tunnel-id');
                if (e.target.checked) {
                    this.selectedTunnels.add(tunnelId);
                } else {
                    this.selectedTunnels.delete(tunnelId);
                }
                this.updateBatchActions();
            });
        });
        
        // 操作按钮事件
        const actionBtns = document.querySelectorAll('.tunnel-action-btn');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.getAttribute('data-action');
                const tunnelId = e.currentTarget.getAttribute('data-tunnel-id');
                if (action && tunnelId) {
                    this.handleTunnelAction(action, tunnelId);
                }
            });
        });
        
        // 下拉菜单项事件
        const dropdownItems = document.querySelectorAll('.dropdown-item');
        dropdownItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const action = e.currentTarget.getAttribute('data-action');
                const tunnelId = e.currentTarget.getAttribute('data-tunnel-id');
                if (action && tunnelId) {
                    this.handleTunnelAction(action, tunnelId);
                }
            });
        });
    }
    
    // 更新批量操作显示
    updateBatchActions() {
        const batchActions = document.getElementById('batchActions');
        const selectedCount = document.getElementById('selectedCount');
        
        if (this.selectedTunnels.size > 0) {
            batchActions.style.display = 'flex';
            selectedCount.textContent = `已选择 ${this.selectedTunnels.size} 个隧道`;
        } else {
            batchActions.style.display = 'none';
        }
    }
    
    // 处理隧道操作
    async handleTunnelAction(action, tunnelId) {
        const tunnel = this.tunnels.get(tunnelId);
        if (!tunnel) return;
        
        const uiManager = window.getUIManager();
        
        try {
            switch (action) {
                case 'start':
                    await this.startTunnel(tunnelId);
                    break;
                case 'stop':
                    await this.stopTunnel(tunnelId);
                    break;
                case 'restart':
                    await this.restartTunnel(tunnelId);
                    break;
                case 'open':
                    this.openTunnel(tunnelId);
                    break;
                case 'edit':
                    this.showEditTunnelDialog(tunnelId);
                    break;
                case 'copy':
                    this.copyTunnelConfig(tunnelId);
                    break;
                case 'logs':
                    this.showTunnelLogs(tunnelId);
                    break;
                case 'delete':
                    this.showDeleteTunnelDialog(tunnelId);
                    break;
            }
        } catch (error) {
            console.error(`Tunnel action ${action} failed:`, error);
            uiManager?.showNotification('操作失败', `${action} 操作失败: ${error.message}`, 'error');
        }
    }
    
    // 显示创建隧道对话框
    showCreateTunnelDialog() {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="create-tunnel-form">
                <div class="form-group">
                    <label for="tunnelName">隧道名称</label>
                    <input type="text" id="tunnelName" class="form-input" placeholder="输入隧道名称" required>
                </div>
                
                <div class="form-group">
                    <label for="tunnelType">隧道类型</label>
                    <select id="tunnelType" class="form-input">
                        <option value="tcp">TCP</option>
                        <option value="udp">UDP</option>
                        <option value="http">HTTP</option>
                        <option value="https">HTTPS</option>
                    </select>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="localIp">本地IP</label>
                        <input type="text" id="localIp" class="form-input" value="127.0.0.1">
                    </div>
                    <div class="form-group">
                        <label for="localPort">本地端口</label>
                        <input type="number" id="localPort" class="form-input" placeholder="25565" required>
                    </div>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="remotePort">远程端口</label>
                        <input type="number" id="remotePort" class="form-input" placeholder="留空自动分配">
                    </div>
                    <div class="form-group">
                        <label for="subdomain">子域名</label>
                        <input type="text" id="subdomain" class="form-input" placeholder="可选">
                    </div>
                </div>
                
                <div class="form-group">
                    <label for="customDomain">自定义域名</label>
                    <input type="text" id="customDomain" class="form-input" placeholder="example.com（可选）">
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('创建隧道', content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn primary" id="confirmCreateTunnelBtn">
                    <i class="mdi mdi-check"></i>
                    <span>创建隧道</span>
                </button>
            `
        });
        
        if (modal) {
            const confirmBtn = modal.querySelector('#confirmCreateTunnelBtn');
            confirmBtn.addEventListener('click', () => this.createTunnel());
        }
    }
    
    // 创建隧道
    async createTunnel() {
        const uiManager = window.getUIManager();
        
        const tunnelData = {
            name: document.getElementById('tunnelName').value.trim(),
            type: document.getElementById('tunnelType').value,
            local_ip: document.getElementById('localIp').value.trim(),
            local_port: parseInt(document.getElementById('localPort').value),
            remote_port: document.getElementById('remotePort').value ? parseInt(document.getElementById('remotePort').value) : null,
            subdomain: document.getElementById('subdomain').value.trim(),
            custom_domain: document.getElementById('customDomain').value.trim()
        };
        
        if (!tunnelData.name || !tunnelData.local_port) {
            uiManager?.showNotification('创建失败', '请填写必要信息', 'warning');
            return;
        }
        
        try {
            const response = await fetch('/api/frp/tunnels', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(tunnelData)
            });
            
            if (response.ok) {
                uiManager?.closeModal();
                uiManager?.showNotification('创建成功', '隧道创建成功', 'success');
                await this.loadTunnels(); // 刷新列表
            } else {
                const error = await response.json();
                uiManager?.showNotification('创建失败', error.message || '创建隧道失败', 'error');
            }
        } catch (error) {
            console.error('Create tunnel failed:', error);
            uiManager?.showNotification('创建失败', '网络错误，请重试', 'error');
        }
    }
    
    // 加载FRP配置
    async loadFRPConfig() {
        try {
            const response = await fetch('/api/frp/config');
            if (response.ok) {
                this.frpConfig = await response.json();
                this.updateConfigForm();
            }
        } catch (error) {
            console.error('Failed to load FRP config:', error);
        }
    }
    
    // 更新配置表单
    updateConfigForm() {
        if (!this.frpConfig) return;
        
        const fields = ['serverAddr', 'serverPort', 'authToken', 'user', 'protocol', 'logLevel', 'heartbeatInterval', 'heartbeatTimeout'];
        fields.forEach(field => {
            const element = document.getElementById(field);
            if (element && this.frpConfig[field] !== undefined) {
                element.value = this.frpConfig[field];
            }
        });
    }
    
    // 加载FRP状态
    async loadFRPStatus() {
        try {
            const response = await fetch('/api/frp/status');
            if (response.ok) {
                const status = await response.json();
                this.updateFRPStatus(status);
            }
        } catch (error) {
            console.error('Failed to load FRP status:', error);
        }
    }
    
    // 更新FRP状态显示
    updateFRPStatus(status) {
        const statusElement = document.getElementById('frpStatus');
        if (!statusElement) return;
        
        const statusIcon = statusElement.querySelector('i');
        const statusText = statusElement.querySelector('span');
        
        statusElement.className = 'status-indicator';
        
        if (status.is_running) {
            statusElement.classList.add('success');
            statusIcon.className = 'mdi mdi-check-circle';
            statusText.textContent = '运行中';
        } else {
            statusElement.classList.add('error');
            statusIcon.className = 'mdi mdi-close-circle';
            statusText.textContent = '已停止';
        }
    }
    
    // 显示隧道错误
    showTunnelsError(message) {
        const tunnelsList = document.getElementById('tunnelsList');
        if (tunnelsList) {
            tunnelsList.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h3>加载失败</h3>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getFRPPageManager().loadTunnels()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 其他操作方法（占位符）
    async startTunnel(tunnelId) {
        console.log('Start tunnel:', tunnelId);
    }
    
    async stopTunnel(tunnelId) {
        console.log('Stop tunnel:', tunnelId);
    }
    
    async restartTunnel(tunnelId) {
        console.log('Restart tunnel:', tunnelId);
    }
    
    openTunnel(tunnelId) {
        console.log('Open tunnel:', tunnelId);
    }
    
    showEditTunnelDialog(tunnelId) {
        console.log('Edit tunnel:', tunnelId);
    }
    
    copyTunnelConfig(tunnelId) {
        console.log('Copy tunnel config:', tunnelId);
    }
    
    showTunnelLogs(tunnelId) {
        console.log('Show tunnel logs:', tunnelId);
    }
    
    showDeleteTunnelDialog(tunnelId) {
        console.log('Delete tunnel:', tunnelId);
    }
    
    async startFRP() {
        console.log('Start FRP');
    }
    
    async stopFRP() {
        console.log('Stop FRP');
    }
    
    async restartFRP() {
        console.log('Restart FRP');
    }
    
    async testFRPConfig() {
        console.log('Test FRP config');
    }
    
    async saveFRPConfig() {
        console.log('Save FRP config');
    }
    
    async loadLogs() {
        console.log('Load logs');
    }
    
    async clearLogs() {
        console.log('Clear logs');
    }
    
    async downloadLogs() {
        console.log('Download logs');
    }
}

// 全局FRP页面管理器实例
let frpPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    frpPageManager = new FRPPageManager();
});

// 导出到全局作用域
window.FRPPageManager = FRPPageManager;
window.getFRPPageManager = () => frpPageManager;
