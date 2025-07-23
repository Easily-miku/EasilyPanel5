// OpenFRP内网穿透页面管理器
class FRPPageManager {
    constructor() {
        this.userInfo = null;
        this.nodes = new Map();
        this.tunnels = new Map();
        this.selectedTunnels = new Set();
        this.searchQuery = '';
        this.currentTab = 'auth'; // auth, tunnels, nodes, logs
        this.isAuthenticated = false;
        this.authorization = null; // OpenFrp Authorization token
        this.apiBase = 'https://api.openfrp.net';
        this.userAgent = 'EasilyPanel5/1.1.0';
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
                <h2>OpenFRP 内网穿透</h2>
                <p>基于 OpenFRP 的内网穿透服务管理</p>
                ${this.isAuthenticated ? `
                    <div class="user-info">
                        <span class="user-status">
                            <i class="mdi mdi-account-check"></i>
                            已认证用户
                        </span>
                    </div>
                ` : ''}
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
                <button class="tab-btn ${this.currentTab === 'auth' ? 'active' : ''}" data-tab="auth">
                    <i class="mdi mdi-key"></i>
                    <span>身份认证</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'tunnels' ? 'active' : ''}" data-tab="tunnels" ${!this.isAuthenticated ? 'disabled' : ''}>
                    <i class="mdi mdi-tunnel"></i>
                    <span>隧道管理</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'nodes' ? 'active' : ''}" data-tab="nodes" ${!this.isAuthenticated ? 'disabled' : ''}>
                    <i class="mdi mdi-server-network"></i>
                    <span>节点列表</span>
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
                <div class="tab-pane ${this.currentTab === 'auth' ? 'active' : ''}" id="auth-tab">
                    ${this.renderAuthTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'tunnels' ? 'active' : ''}" id="tunnels-tab">
                    ${this.renderTunnelsTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'nodes' ? 'active' : ''}" id="nodes-tab">
                    ${this.renderNodesTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'logs' ? 'active' : ''}" id="logs-tab">
                    ${this.renderLogsTab()}
                </div>
            </div>
        `;
    }

    // 渲染身份认证标签页
    renderAuthTab() {
        if (this.isAuthenticated && this.userInfo) {
            return `
                <div class="auth-success">
                    <div class="user-card">
                        <div class="user-avatar">
                            <i class="mdi mdi-account-circle"></i>
                        </div>
                        <div class="user-details">
                            <h3>${this.userInfo.username}</h3>
                            <p class="user-email">${this.userInfo.email}</p>
                            <p class="user-group">${this.userInfo.friendlyGroup}</p>
                            <div class="user-stats">
                                <div class="stat-item">
                                    <span class="stat-label">隧道数量</span>
                                    <span class="stat-value">${this.userInfo.used}/${this.userInfo.proxies}</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">剩余流量</span>
                                    <span class="stat-value">${this.formatTraffic(this.userInfo.traffic)}</span>
                                </div>
                                <div class="stat-item">
                                    <span class="stat-label">带宽限制</span>
                                    <span class="stat-value">${this.userInfo.inLimit}/${this.userInfo.outLimit} Kbps</span>
                                </div>
                            </div>
                        </div>
                        <div class="user-actions">
                            <button class="btn secondary" id="logoutBtn">
                                <i class="mdi mdi-logout"></i>
                                <span>退出登录</span>
                            </button>
                        </div>
                    </div>
                </div>
            `;
        } else {
            return `
                <div class="auth-form">
                    <div class="auth-header">
                        <h3>OpenFrp 身份认证</h3>
                        <p>请选择认证方式登录到 OpenFrp 服务</p>
                    </div>

                    <div class="auth-methods">
                        <div class="auth-method recommended">
                            <div class="method-header">
                                <h4>
                                    <i class="mdi mdi-shield-check"></i>
                                    Authorization 登录（推荐）
                                </h4>
                                <span class="method-badge">安全</span>
                            </div>
                            <p class="method-description">
                                在 OpenFrp 管理面板的个人中心获取 Authorization 密钥，安全便捷。
                            </p>
                            <div class="auth-input-group">
                                <label for="authToken">Authorization 密钥</label>
                                <textarea id="authToken" placeholder="请粘贴从 OpenFrp 面板获取的 Authorization 密钥" rows="3"></textarea>
                                <div class="input-help">
                                    <a href="https://console.openfrp.net/usercenter" target="_blank">
                                        <i class="mdi mdi-open-in-new"></i>
                                        前往 OpenFrp 面板获取
                                    </a>
                                </div>
                            </div>
                            <button class="btn primary" id="authLoginBtn">
                                <i class="mdi mdi-login"></i>
                                <span>登录</span>
                            </button>
                        </div>

                        <div class="auth-method">
                            <div class="method-header">
                                <h4>
                                    <i class="mdi mdi-web"></i>
                                    远程安全登录
                                </h4>
                                <span class="method-badge">高级</span>
                            </div>
                            <p class="method-description">
                                通过浏览器授权登录，更加安全但需要额外步骤。
                            </p>
                            <button class="btn secondary" id="remoteLoginBtn">
                                <i class="mdi mdi-launch"></i>
                                <span>启动远程登录</span>
                            </button>
                        </div>
                    </div>
                </div>
            `;
        }
    }

    // 渲染隧道管理标签页
    renderTunnelsTab() {
        if (!this.isAuthenticated) {
            return `
                <div class="auth-required">
                    <div class="auth-prompt">
                        <i class="mdi mdi-lock"></i>
                        <h3>需要身份认证</h3>
                        <p>请先在身份认证标签页完成登录</p>
                        <button class="btn primary" onclick="window.getFRPPageManager().switchTab('auth')">
                            前往认证
                        </button>
                    </div>
                </div>
            `;
        }

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
                        </select>
                    </div>

                    <div class="type-filter">
                        <select id="typeFilter" class="filter-select">
                            <option value="all">全部类型</option>
                            <option value="tcp">TCP</option>
                            <option value="udp">UDP</option>
                            <option value="http">HTTP</option>
                            <option value="https">HTTPS</option>
                            <option value="stcp">STCP</option>
                            <option value="xtcp">XTCP</option>
                        </select>
                    </div>
                </div>

                <div class="toolbar-right">
                    <div class="batch-actions" id="batchActions" style="display: none;">
                        <span class="selected-count" id="selectedCount">已选择 0 个隧道</span>
                        <button class="btn warning" id="batchDeleteBtn">
                            <i class="mdi mdi-delete"></i>
                            <span>批量删除</span>
                        </button>
                    </div>

                    <button class="btn secondary" id="refreshTunnelsBtn">
                        <i class="mdi mdi-refresh"></i>
                        <span>刷新</span>
                    </button>

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

    // 渲染节点列表标签页
    renderNodesTab() {
        if (!this.isAuthenticated) {
            return `
                <div class="auth-required">
                    <div class="auth-prompt">
                        <i class="mdi mdi-lock"></i>
                        <h3>需要身份认证</h3>
                        <p>请先在身份认证标签页完成登录</p>
                        <button class="btn primary" onclick="window.getFRPPageManager().switchTab('auth')">
                            前往认证
                        </button>
                    </div>
                </div>
            `;
        }

        return `
            <div class="nodes-toolbar">
                <div class="toolbar-left">
                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="nodeSearch" placeholder="搜索节点..." value="">
                    </div>

                    <div class="region-filter">
                        <select id="regionFilter" class="filter-select">
                            <option value="all">全部地区</option>
                            <option value="1">中国大陆</option>
                            <option value="2">港澳台地区</option>
                            <option value="3">海外地区</option>
                        </select>
                    </div>

                    <div class="status-filter">
                        <select id="nodeStatusFilter" class="filter-select">
                            <option value="all">全部状态</option>
                            <option value="200">正常</option>
                            <option value="other">异常</option>
                        </select>
                    </div>
                </div>

                <div class="toolbar-right">
                    <button class="btn secondary" id="refreshNodesBtn">
                        <i class="mdi mdi-refresh"></i>
                        <span>刷新节点</span>
                    </button>
                </div>
            </div>

            <div class="nodes-list-container">
                <div class="nodes-list" id="nodesList">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载节点列表...</span>
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
            const result = await response.json();

            if (response.ok && result.success) {
                this.tunnels.clear();

                // 处理新的API响应格式
                const tunnelsData = result.data || [];
                tunnelsData.forEach(tunnel => {
                    this.tunnels.set(tunnel.id, tunnel);
                });

                this.filterAndRenderTunnels();
            } else {
                const errorMessage = result.message || result.error || '加载隧道列表失败';
                this.showTunnelsError(errorMessage);
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

        // 获取表单数据
        const name = document.getElementById('tunnelName').value.trim();
        const type = document.getElementById('tunnelType').value;
        const localIp = document.getElementById('localIp').value.trim() || '127.0.0.1';
        const localPort = parseInt(document.getElementById('localPort').value);
        const remotePort = document.getElementById('remotePort').value ? parseInt(document.getElementById('remotePort').value) : null;
        const subdomain = document.getElementById('subdomain').value.trim();
        const customDomain = document.getElementById('customDomain').value.trim();

        // 客户端验证
        if (!name) {
            uiManager?.showNotification('创建失败', '请输入隧道名称', 'warning');
            return;
        }

        if (!type) {
            uiManager?.showNotification('创建失败', '请选择隧道类型', 'warning');
            return;
        }

        if (!localPort || localPort < 1 || localPort > 65535) {
            uiManager?.showNotification('创建失败', '请输入有效的本地端口（1-65535）', 'warning');
            return;
        }

        if (remotePort && (remotePort < 1 || remotePort > 65535)) {
            uiManager?.showNotification('创建失败', '请输入有效的远程端口（1-65535）', 'warning');
            return;
        }

        // 构建隧道数据（匹配后端API格式）
        const tunnelData = {
            name: name,
            type: type,
            token: 'default-token', // 这里应该从配置中获取
            local_ip: localIp,
            local_port: localPort,
            remote_port: remotePort,
            subdomain: subdomain,
            custom_domain: customDomain
        };

        try {
            const response = await fetch('/api/frp/tunnels', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(tunnelData)
            });

            const result = await response.json();

            if (response.ok && result.success) {
                uiManager?.closeModal();
                uiManager?.showNotification('创建成功', '隧道创建成功', 'success');
                await this.loadTunnels(); // 刷新列表
            } else {
                const errorMessage = result.message || result.error || '创建隧道失败';
                uiManager?.showNotification('创建失败', errorMessage, 'error');
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
    
    // OpenFrp API 调用方法

    // Authorization 登录
    async loginWithAuthorization(authToken) {
        try {
            // 验证 Authorization 是否有效
            const response = await this.apiCall('/frp/api/getUserInfo', 'POST', null, authToken);

            if (response.flag) {
                this.authorization = authToken;
                this.userInfo = response.data;
                this.isAuthenticated = true;

                // 保存到本地存储
                localStorage.setItem('openfrp_authorization', authToken);

                return { success: true, data: response.data };
            } else {
                return { success: false, message: response.msg || '认证失败' };
            }
        } catch (error) {
            console.error('Authorization login failed:', error);
            return { success: false, message: '网络错误，请重试' };
        }
    }

    // 退出登录
    logout() {
        this.authorization = null;
        this.userInfo = null;
        this.isAuthenticated = false;
        this.tunnels.clear();
        this.nodes.clear();

        // 清除本地存储
        localStorage.removeItem('openfrp_authorization');

        // 重新渲染页面
        this.renderPage();
    }

    // 获取用户隧道列表
    async getUserTunnels() {
        try {
            const response = await this.apiCall('/frp/api/getUserProxies', 'POST');

            if (response.flag) {
                return { success: true, data: response.data };
            } else {
                return { success: false, message: response.msg || '获取隧道列表失败' };
            }
        } catch (error) {
            console.error('Get user tunnels failed:', error);
            return { success: false, message: '网络错误，请重试' };
        }
    }

    // 获取节点列表
    async getNodeList() {
        try {
            const response = await this.apiCall('/frp/api/getNodeList', 'POST');

            if (response.flag) {
                return { success: true, data: response.data };
            } else {
                return { success: false, message: response.msg || '获取节点列表失败' };
            }
        } catch (error) {
            console.error('Get node list failed:', error);
            return { success: false, message: '网络错误，请重试' };
        }
    }

    // 创建新隧道
    async createNewTunnel(tunnelData) {
        try {
            const response = await this.apiCall('/frp/api/newProxy', 'POST', tunnelData);

            if (response.flag) {
                return { success: true, message: response.msg || '创建成功' };
            } else {
                return { success: false, message: response.msg || '创建隧道失败' };
            }
        } catch (error) {
            console.error('Create tunnel failed:', error);
            return { success: false, message: '网络错误，请重试' };
        }
    }

    // 删除隧道
    async deleteTunnel(proxyId) {
        try {
            const response = await this.apiCall('/frp/api/removeProxy', 'POST', { proxy_id: proxyId });

            if (response.flag) {
                return { success: true, message: response.msg || '删除成功' };
            } else {
                return { success: false, message: response.msg || '删除隧道失败' };
            }
        } catch (error) {
            console.error('Delete tunnel failed:', error);
            return { success: false, message: '网络错误，请重试' };
        }
    }

    // 编辑隧道
    async editTunnel(tunnelData) {
        try {
            const response = await this.apiCall('/frp/api/editProxy', 'POST', tunnelData);

            if (response.flag) {
                return { success: true, message: response.msg || '保存成功' };
            } else {
                return { success: false, message: response.msg || '保存失败' };
            }
        } catch (error) {
            console.error('Edit tunnel failed:', error);
            return { success: false, message: '网络错误，请重试' };
        }
    }

    // 通用 API 调用方法
    async apiCall(endpoint, method = 'GET', data = null, authToken = null) {
        const url = this.apiBase + endpoint;
        const headers = {
            'Content-Type': 'application/json',
            'User-Agent': this.userAgent
        };

        // 添加 Authorization 头
        const token = authToken || this.authorization;
        if (token) {
            headers['Authorization'] = token;
        }

        const options = {
            method: method,
            headers: headers
        };

        if (data && (method === 'POST' || method === 'PUT')) {
            options.body = JSON.stringify(data);
        }

        const response = await fetch(url, options);

        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }

        return await response.json();
    }

    // 工具方法
    formatTraffic(traffic) {
        if (traffic < 1024) {
            return `${traffic} MB`;
        } else {
            return `${(traffic / 1024).toFixed(2)} GB`;
        }
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';

        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));

        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    getRegionName(classify) {
        const regions = {
            1: '中国大陆',
            2: '港澳台地区',
            3: '海外地区'
        };
        return regions[classify] || '未知地区';
    }

    getTunnelTypeIcon(type) {
        const icons = {
            'tcp': 'mdi-ethernet',
            'udp': 'mdi-ethernet',
            'http': 'mdi-web',
            'https': 'mdi-web',
            'stcp': 'mdi-ethernet-cable',
            'xtcp': 'mdi-ethernet-cable'
        };
        return icons[type] || 'mdi-help-circle';
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
