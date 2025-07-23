// 服务器管理页面管理器
class ServersPageManager {
    constructor() {
        this.servers = new Map();
        this.selectedServers = new Set();
        this.currentFilter = 'all';
        this.currentSort = 'name';
        this.searchQuery = '';
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadServers();
    }
    
    // 渲染页面
    renderPage() {
        const serversPage = document.getElementById('servers-page');
        if (!serversPage) return;
        
        serversPage.innerHTML = `
            <div class="page-header">
                <h2>服务器管理</h2>
                <p>管理您的Minecraft服务器实例</p>
            </div>
            
            <div class="servers-container">
                ${this.renderToolbar()}
                ${this.renderServersList()}
            </div>
        `;
    }
    
    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="servers-toolbar">
                <div class="toolbar-left">
                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="serverSearch" placeholder="搜索服务器..." value="${this.searchQuery}">
                    </div>
                    
                    <div class="filter-group">
                        <select id="serverFilter" class="filter-select">
                            <option value="all">全部服务器</option>
                            <option value="running">运行中</option>
                            <option value="stopped">已停止</option>
                            <option value="error">错误状态</option>
                        </select>
                        
                        <select id="serverSort" class="filter-select">
                            <option value="name">按名称排序</option>
                            <option value="status">按状态排序</option>
                            <option value="created">按创建时间</option>
                            <option value="memory">按内存使用</option>
                        </select>
                    </div>
                </div>
                
                <div class="toolbar-right">
                    <div class="batch-actions" id="batchActions" style="display: none;">
                        <span class="selected-count" id="selectedCount">已选择 0 个服务器</span>
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
                    
                    <button class="btn primary" id="createServerBtn">
                        <i class="mdi mdi-plus"></i>
                        <span>创建服务器</span>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 渲染服务器列表
    renderServersList() {
        return `
            <div class="servers-list-container">
                <div class="servers-list" id="serversList">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载服务器列表...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 搜索框
        const searchInput = document.getElementById('serverSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderServers();
            });
        }
        
        // 过滤器
        const filterSelect = document.getElementById('serverFilter');
        if (filterSelect) {
            filterSelect.addEventListener('change', (e) => {
                this.currentFilter = e.target.value;
                this.filterAndRenderServers();
            });
        }
        
        // 排序
        const sortSelect = document.getElementById('serverSort');
        if (sortSelect) {
            sortSelect.addEventListener('change', (e) => {
                this.currentSort = e.target.value;
                this.filterAndRenderServers();
            });
        }
        
        // 创建服务器按钮
        const createBtn = document.getElementById('createServerBtn');
        if (createBtn) {
            createBtn.addEventListener('click', () => this.showCreateServerDialog());
        }
        
        // 批量操作按钮
        const batchStartBtn = document.getElementById('batchStartBtn');
        const batchStopBtn = document.getElementById('batchStopBtn');
        const batchDeleteBtn = document.getElementById('batchDeleteBtn');
        
        if (batchStartBtn) {
            batchStartBtn.addEventListener('click', () => this.batchStartServers());
        }
        if (batchStopBtn) {
            batchStopBtn.addEventListener('click', () => this.batchStopServers());
        }
        if (batchDeleteBtn) {
            batchDeleteBtn.addEventListener('click', () => this.batchDeleteServers());
        }
    }
    
    // 加载服务器列表
    async loadServers() {
        try {
            const response = await fetch('/api/servers');
            if (response.ok) {
                const servers = await response.json();
                this.servers.clear();
                
                servers.forEach(server => {
                    this.servers.set(server.id, server);
                });
                
                this.filterAndRenderServers();
            } else {
                this.showError('加载服务器列表失败');
            }
        } catch (error) {
            console.error('Failed to load servers:', error);
            this.showError('网络错误，请重试');
        }
    }
    
    // 过滤和渲染服务器
    filterAndRenderServers() {
        const serversList = document.getElementById('serversList');
        if (!serversList) return;
        
        // 获取过滤后的服务器
        let filteredServers = Array.from(this.servers.values());
        
        // 应用搜索过滤
        if (this.searchQuery) {
            const query = this.searchQuery.toLowerCase();
            filteredServers = filteredServers.filter(server => 
                server.name.toLowerCase().includes(query) ||
                server.description?.toLowerCase().includes(query)
            );
        }
        
        // 应用状态过滤
        if (this.currentFilter !== 'all') {
            filteredServers = filteredServers.filter(server => 
                server.status === this.currentFilter
            );
        }
        
        // 应用排序
        filteredServers.sort((a, b) => {
            switch (this.currentSort) {
                case 'name':
                    return a.name.localeCompare(b.name);
                case 'status':
                    return a.status.localeCompare(b.status);
                case 'created':
                    return new Date(b.created_at) - new Date(a.created_at);
                case 'memory':
                    return (b.memory_usage || 0) - (a.memory_usage || 0);
                default:
                    return 0;
            }
        });
        
        // 渲染服务器卡片
        if (filteredServers.length === 0) {
            serversList.innerHTML = this.renderEmptyState();
        } else {
            serversList.innerHTML = filteredServers.map(server => 
                this.renderServerCard(server)
            ).join('');
        }
        
        // 重新绑定事件
        this.bindServerCardEvents();
    }
    
    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-server-off"></i>
                <h3>没有找到服务器</h3>
                <p>您还没有创建任何服务器，或者当前过滤条件没有匹配的结果。</p>
                <button class="btn primary" onclick="window.getServersPageManager().showCreateServerDialog()">
                    <i class="mdi mdi-plus"></i>
                    <span>创建第一个服务器</span>
                </button>
            </div>
        `;
    }
    
    // 渲染服务器卡片
    renderServerCard(server) {
        const statusIcons = {
            'running': 'mdi-play-circle',
            'stopped': 'mdi-stop-circle',
            'starting': 'mdi-loading mdi-spin',
            'stopping': 'mdi-loading mdi-spin',
            'error': 'mdi-alert-circle'
        };
        
        const statusColors = {
            'running': 'success',
            'stopped': 'secondary',
            'starting': 'warning',
            'stopping': 'warning',
            'error': 'error'
        };
        
        const statusTexts = {
            'running': '运行中',
            'stopped': '已停止',
            'starting': '启动中',
            'stopping': '停止中',
            'error': '错误'
        };
        
        return `
            <div class="server-card" data-server-id="${server.id}">
                <div class="server-card-header">
                    <div class="server-select">
                        <input type="checkbox" class="server-checkbox" data-server-id="${server.id}">
                    </div>
                    <div class="server-info">
                        <h3 class="server-name">${server.name}</h3>
                        <p class="server-description">${server.description || '无描述'}</p>
                    </div>
                    <div class="server-status">
                        <span class="status-badge ${statusColors[server.status]}">
                            <i class="mdi ${statusIcons[server.status]}"></i>
                            <span>${statusTexts[server.status]}</span>
                        </span>
                    </div>
                </div>
                
                <div class="server-card-body">
                    <div class="server-stats">
                        <div class="stat-item">
                            <span class="stat-label">版本</span>
                            <span class="stat-value">${server.version || 'Unknown'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">端口</span>
                            <span class="stat-value">${server.port || 25565}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">内存</span>
                            <span class="stat-value">${server.memory || '1G'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">在线玩家</span>
                            <span class="stat-value">${server.online_players || 0}/${server.max_players || 20}</span>
                        </div>
                    </div>
                </div>
                
                <div class="server-card-footer">
                    <div class="server-actions">
                        ${server.status === 'stopped' ? 
                            `<button class="btn success server-action-btn" data-action="start" data-server-id="${server.id}">
                                <i class="mdi mdi-play"></i>
                                <span>启动</span>
                            </button>` :
                            `<button class="btn warning server-action-btn" data-action="stop" data-server-id="${server.id}">
                                <i class="mdi mdi-stop"></i>
                                <span>停止</span>
                            </button>`
                        }
                        
                        <button class="btn server-action-btn" data-action="restart" data-server-id="${server.id}">
                            <i class="mdi mdi-restart"></i>
                            <span>重启</span>
                        </button>
                        
                        <button class="btn server-action-btn" data-action="console" data-server-id="${server.id}">
                            <i class="mdi mdi-console"></i>
                            <span>控制台</span>
                        </button>
                        
                        <div class="dropdown">
                            <button class="btn server-action-btn dropdown-toggle" data-server-id="${server.id}">
                                <i class="mdi mdi-dots-vertical"></i>
                            </button>
                            <div class="dropdown-menu">
                                <a href="#" class="dropdown-item" data-action="edit" data-server-id="${server.id}">
                                    <i class="mdi mdi-pencil"></i>
                                    <span>编辑配置</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="backup" data-server-id="${server.id}">
                                    <i class="mdi mdi-backup-restore"></i>
                                    <span>备份</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="files" data-server-id="${server.id}">
                                    <i class="mdi mdi-folder"></i>
                                    <span>文件管理</span>
                                </a>
                                <div class="dropdown-divider"></div>
                                <a href="#" class="dropdown-item danger" data-action="delete" data-server-id="${server.id}">
                                    <i class="mdi mdi-delete"></i>
                                    <span>删除服务器</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 绑定服务器卡片事件
    bindServerCardEvents() {
        // 复选框事件
        const checkboxes = document.querySelectorAll('.server-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                const serverId = e.target.getAttribute('data-server-id');
                if (e.target.checked) {
                    this.selectedServers.add(serverId);
                } else {
                    this.selectedServers.delete(serverId);
                }
                this.updateBatchActions();
            });
        });
        
        // 操作按钮事件
        const actionBtns = document.querySelectorAll('.server-action-btn');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.getAttribute('data-action');
                const serverId = e.currentTarget.getAttribute('data-server-id');
                if (action && serverId) {
                    this.handleServerAction(action, serverId);
                }
            });
        });
        
        // 下拉菜单项事件
        const dropdownItems = document.querySelectorAll('.dropdown-item');
        dropdownItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const action = e.currentTarget.getAttribute('data-action');
                const serverId = e.currentTarget.getAttribute('data-server-id');
                if (action && serverId) {
                    this.handleServerAction(action, serverId);
                }
            });
        });
    }
    
    // 更新批量操作显示
    updateBatchActions() {
        const batchActions = document.getElementById('batchActions');
        const selectedCount = document.getElementById('selectedCount');
        
        if (this.selectedServers.size > 0) {
            batchActions.style.display = 'flex';
            selectedCount.textContent = `已选择 ${this.selectedServers.size} 个服务器`;
        } else {
            batchActions.style.display = 'none';
        }
    }
    
    // 处理服务器操作
    async handleServerAction(action, serverId) {
        const server = this.servers.get(serverId);
        if (!server) return;
        
        const uiManager = window.getUIManager();
        
        try {
            switch (action) {
                case 'start':
                    await this.startServer(serverId);
                    break;
                case 'stop':
                    await this.stopServer(serverId);
                    break;
                case 'restart':
                    await this.restartServer(serverId);
                    break;
                case 'console':
                    this.openConsole(serverId);
                    break;
                case 'edit':
                    this.showEditServerDialog(serverId);
                    break;
                case 'backup':
                    await this.backupServer(serverId);
                    break;
                case 'files':
                    this.openFileManager(serverId);
                    break;
                case 'delete':
                    this.showDeleteServerDialog(serverId);
                    break;
            }
        } catch (error) {
            console.error(`Server action ${action} failed:`, error);
            uiManager?.showNotification('操作失败', `${action} 操作失败: ${error.message}`, 'error');
        }
    }
    
    // 启动服务器
    async startServer(serverId) {
        const response = await fetch(`/api/servers/${serverId}/start`, {
            method: 'POST'
        });
        
        if (response.ok) {
            const uiManager = window.getUIManager();
            uiManager?.showNotification('启动成功', '服务器正在启动...', 'success');
            await this.loadServers(); // 刷新列表
        } else {
            throw new Error('启动失败');
        }
    }
    
    // 停止服务器
    async stopServer(serverId) {
        const response = await fetch(`/api/servers/${serverId}/stop`, {
            method: 'POST'
        });
        
        if (response.ok) {
            const uiManager = window.getUIManager();
            uiManager?.showNotification('停止成功', '服务器正在停止...', 'success');
            await this.loadServers(); // 刷新列表
        } else {
            throw new Error('停止失败');
        }
    }
    
    // 重启服务器
    async restartServer(serverId) {
        const response = await fetch(`/api/servers/${serverId}/restart`, {
            method: 'POST'
        });
        
        if (response.ok) {
            const uiManager = window.getUIManager();
            uiManager?.showNotification('重启成功', '服务器正在重启...', 'success');
            await this.loadServers(); // 刷新列表
        } else {
            throw new Error('重启失败');
        }
    }
    
    // 显示创建服务器对话框
    showCreateServerDialog() {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="create-server-form">
                <div class="form-group">
                    <label for="serverName">服务器名称</label>
                    <input type="text" id="serverName" class="form-input" placeholder="输入服务器名称" required>
                </div>
                
                <div class="form-group">
                    <label for="serverDescription">服务器描述</label>
                    <textarea id="serverDescription" class="form-input" placeholder="输入服务器描述（可选）" rows="3"></textarea>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="serverVersion">Minecraft版本</label>
                        <select id="serverVersion" class="form-input">
                            <option value="1.20.1">1.20.1</option>
                            <option value="1.19.4">1.19.4</option>
                            <option value="1.18.2">1.18.2</option>
                            <option value="1.16.5">1.16.5</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="serverCore">服务器核心</label>
                        <select id="serverCore" class="form-input">
                            <option value="paper">Paper</option>
                            <option value="spigot">Spigot</option>
                            <option value="bukkit">Bukkit</option>
                            <option value="vanilla">Vanilla</option>
                        </select>
                    </div>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="serverPort">端口</label>
                        <input type="number" id="serverPort" class="form-input" value="25565" min="1024" max="65535">
                    </div>
                    
                    <div class="form-group">
                        <label for="serverMemory">内存分配</label>
                        <select id="serverMemory" class="form-input">
                            <option value="512M">512MB</option>
                            <option value="1G" selected>1GB</option>
                            <option value="2G">2GB</option>
                            <option value="4G">4GB</option>
                            <option value="8G">8GB</option>
                        </select>
                    </div>
                </div>
                
                <div class="form-group">
                    <label for="maxPlayers">最大玩家数</label>
                    <input type="number" id="maxPlayers" class="form-input" value="20" min="1" max="1000">
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('创建服务器', content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn primary" id="confirmCreateBtn">
                    <i class="mdi mdi-check"></i>
                    <span>创建服务器</span>
                </button>
            `
        });
        
        if (modal) {
            const confirmBtn = modal.querySelector('#confirmCreateBtn');
            confirmBtn.addEventListener('click', () => this.createServer());
        }
    }
    
    // 创建服务器
    async createServer() {
        const uiManager = window.getUIManager();
        
        const serverData = {
            name: document.getElementById('serverName').value.trim(),
            description: document.getElementById('serverDescription').value.trim(),
            version: document.getElementById('serverVersion').value,
            core: document.getElementById('serverCore').value,
            port: parseInt(document.getElementById('serverPort').value),
            memory: document.getElementById('serverMemory').value,
            max_players: parseInt(document.getElementById('maxPlayers').value)
        };
        
        if (!serverData.name) {
            uiManager?.showNotification('创建失败', '请输入服务器名称', 'warning');
            return;
        }
        
        try {
            const response = await fetch('/api/servers', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(serverData)
            });
            
            if (response.ok) {
                uiManager?.closeModal();
                uiManager?.showNotification('创建成功', '服务器创建成功', 'success');
                await this.loadServers(); // 刷新列表
            } else {
                const error = await response.json();
                uiManager?.showNotification('创建失败', error.message || '创建服务器失败', 'error');
            }
        } catch (error) {
            console.error('Create server failed:', error);
            uiManager?.showNotification('创建失败', '网络错误，请重试', 'error');
        }
    }
    
    // 显示错误信息
    showError(message) {
        const serversList = document.getElementById('serversList');
        if (serversList) {
            serversList.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h3>加载失败</h3>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getServersPageManager().loadServers()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 批量启动服务器
    async batchStartServers() {
        // TODO: 实现批量启动
        console.log('Batch start servers:', Array.from(this.selectedServers));
    }
    
    // 批量停止服务器
    async batchStopServers() {
        // TODO: 实现批量停止
        console.log('Batch stop servers:', Array.from(this.selectedServers));
    }
    
    // 批量删除服务器
    async batchDeleteServers() {
        // TODO: 实现批量删除
        console.log('Batch delete servers:', Array.from(this.selectedServers));
    }
    
    // 打开控制台
    openConsole(serverId) {
        // TODO: 实现控制台功能
        console.log('Open console for server:', serverId);
    }
    
    // 显示编辑服务器对话框
    showEditServerDialog(serverId) {
        // TODO: 实现编辑功能
        console.log('Edit server:', serverId);
    }
    
    // 备份服务器
    async backupServer(serverId) {
        // TODO: 实现备份功能
        console.log('Backup server:', serverId);
    }
    
    // 打开文件管理器
    openFileManager(serverId) {
        // TODO: 实现文件管理功能
        console.log('Open file manager for server:', serverId);
    }
    
    // 显示删除服务器对话框
    showDeleteServerDialog(serverId) {
        // TODO: 实现删除功能
        console.log('Delete server:', serverId);
    }
}

// 全局服务器页面管理器实例
let serversPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    serversPageManager = new ServersPageManager();
});

// 导出到全局作用域
window.ServersPageManager = ServersPageManager;
window.getServersPageManager = () => serversPageManager;
