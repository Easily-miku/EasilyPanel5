// 主应用类
class EasilyPanel {
    constructor() {
        this.currentTab = 'servers';
        this.servers = new Map();
        this.downloadTasks = new Map();
        this.currentServerId = null;
        this.wsManager = null;
        
        this.init();
    }
    
    async init() {
        console.log('开始初始化EasilyPanel...');

        // 初始化UI
        this.setupTabNavigation();
        this.setupModals();
        this.setupEventHandlers();

        // 加载初始数据
        await this.loadInitialData();

        // 初始化WebSocket管理器
        console.log('初始化WebSocket管理器...');
        this.wsManager = initWebSocket();
        this.setupWebSocketHandlers();

        // 延迟启动WebSocket连接
        setTimeout(() => {
            this.startWebSocketIfNeeded();
        }, 2000);

        console.log('EasilyPanel初始化完成');
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
    
    setupWebSocketHandlers() {
        if (!this.wsManager) return;

        this.wsManager.on('connected', (data) => {
            console.log('WebSocket连接已建立');
            // 只在重连成功时显示通知
            if (data && data.reconnected) {
                this.showNotification('连接已恢复', 'success');
            }
        });

        this.wsManager.on('disconnected', (data) => {
            console.log('WebSocket连接已断开:', data);
            // 只在意外断开且之前已连接时显示警告
            if (data && data.wasConnected && data.code !== 1000 && data.code !== 1001) {
                this.showNotification('连接已断开，正在重连...', 'warning');
            }
        });

        this.wsManager.on('reconnect_failed', (data) => {
            this.showNotification(`连接失败 (${data.attempts}次重试)，请检查网络或刷新页面`, 'error');
        });
        
        this.wsManager.on('log_message', (message) => {
            this.handleLogMessage(message);
        });
        
        this.wsManager.on('server_status', (message) => {
            this.handleServerStatus(message);
        });
        
        this.wsManager.on('download_progress', (message) => {
            this.handleDownloadProgress(message);
        });
        
        this.wsManager.on('server_error', (message) => {
            this.showNotification(`服务器错误: ${message.error}`, 'error');
        });
    }
    
    setupTabNavigation() {
        const tabButtons = document.querySelectorAll('.tab-btn');
        const tabContents = document.querySelectorAll('.tab-content');
        
        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                const tabId = button.dataset.tab;

                // 启动WebSocket连接（如果还没有连接）
                this.startWebSocketIfNeeded();

                // 更新按钮状态
                tabButtons.forEach(btn => btn.classList.remove('active'));
                button.classList.add('active');

                // 更新内容显示
                tabContents.forEach(content => content.classList.remove('active'));
                document.getElementById(tabId).classList.add('active');

                this.currentTab = tabId;
                this.onTabChanged(tabId);
            });
        });
    }
    
    setupModals() {
        // 控制台模态框
        const consoleModal = document.getElementById('console-modal');
        const closeConsole = document.getElementById('close-console');
        
        closeConsole.addEventListener('click', () => {
            this.closeConsole();
        });
        
        // 创建服务器模态框
        const createServerModal = document.getElementById('create-server-modal');
        const createServerBtn = document.getElementById('create-server-btn');
        const closeCreateServer = document.getElementById('close-create-server');
        const cancelCreateServer = document.getElementById('cancel-create-server');
        
        createServerBtn.addEventListener('click', () => {
            this.showCreateServerModal();
        });
        
        closeCreateServer.addEventListener('click', () => {
            this.hideCreateServerModal();
        });
        
        cancelCreateServer.addEventListener('click', () => {
            this.hideCreateServerModal();
        });
        
        // 点击模态框外部关闭
        window.addEventListener('click', (event) => {
            if (event.target === consoleModal) {
                this.closeConsole();
            }
            if (event.target === createServerModal) {
                this.hideCreateServerModal();
            }
        });
    }
    
    setupEventHandlers() {
        // Java检测按钮
        document.getElementById('detect-java-btn').addEventListener('click', () => {
            this.detectJava();
        });
        
        // 保存Java配置按钮
        document.getElementById('save-java-config').addEventListener('click', () => {
            this.saveJavaConfig();
        });
        
        // 核心类型选择
        document.getElementById('core-type').addEventListener('change', (e) => {
            this.onCoreTypeChanged(e.target.value);
        });
        
        // MC版本选择
        document.getElementById('mc-version').addEventListener('change', (e) => {
            this.onMCVersionChanged(e.target.value);
        });
        
        // 下载按钮
        document.getElementById('download-btn').addEventListener('click', () => {
            this.startDownload();
        });
        
        // 创建服务器表单
        document.getElementById('create-server-form').addEventListener('submit', (e) => {
            e.preventDefault();
            this.createServer();
        });
        
        // 服务器核心类型选择
        document.getElementById('server-core-type').addEventListener('change', (e) => {
            this.onServerCoreTypeChanged(e.target.value);
        });
        
        // 控制台相关
        this.setupConsoleHandlers();
    }
    
    setupConsoleHandlers() {
        const commandInput = document.getElementById('command-input');
        const sendCommandBtn = document.getElementById('send-command');
        const startServerBtn = document.getElementById('start-server');
        const stopServerBtn = document.getElementById('stop-server');
        const restartServerBtn = document.getElementById('restart-server');
        const clearConsoleBtn = document.getElementById('clear-console');
        
        // 发送命令
        const sendCommand = () => {
            const command = commandInput.value.trim();
            if (command && this.currentServerId) {
                this.wsManager.sendCommand(this.currentServerId, command);
                commandInput.value = '';
            }
        };
        
        sendCommandBtn.addEventListener('click', sendCommand);
        commandInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                sendCommand();
            }
        });
        
        // 服务器控制
        startServerBtn.addEventListener('click', () => {
            this.startServer(this.currentServerId);
        });
        
        stopServerBtn.addEventListener('click', () => {
            this.stopServer(this.currentServerId);
        });
        
        restartServerBtn.addEventListener('click', () => {
            this.restartServer(this.currentServerId);
        });
        
        // 清空控制台
        clearConsoleBtn.addEventListener('click', () => {
            document.getElementById('console-output').innerHTML = '';
        });
    }
    
    async loadInitialData() {
        try {
            // 检测Java环境
            await this.detectJava();
            
            // 加载服务器列表
            await this.loadServers();
            
            // 加载核心列表
            await this.loadCoresList();
            
        } catch (error) {
            console.error('加载初始数据失败:', error);
            this.showNotification('加载初始数据失败', 'error');
        }
    }
    
    onTabChanged(tabId) {
        switch (tabId) {
            case 'servers':
                this.loadServers();
                break;
            case 'download':
                this.loadCoresList();
                break;
            case 'settings':
                this.loadJavaConfig();
                break;
        }
    }
    
    // API调用方法
    async apiCall(url, options = {}) {
        try {
            const response = await fetch(url, {
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                },
                ...options
            });
            
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('API调用失败:', error);
            throw error;
        }
    }
    
    // 通知系统
    showNotification(message, type = 'info', duration = 5000) {
        const container = document.getElementById('notifications');
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        notification.innerHTML = `
            <div class="notification-content">
                <strong>${this.getNotificationTitle(type)}</strong>
                <p>${message}</p>
            </div>
        `;
        
        container.appendChild(notification);
        
        // 自动移除
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, duration);
        
        // 点击移除
        notification.addEventListener('click', () => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        });
    }
    
    getNotificationTitle(type) {
        switch (type) {
            case 'success': return '成功';
            case 'error': return '错误';
            case 'warning': return '警告';
            default: return '信息';
        }
    }

    // Java环境相关方法
    async detectJava() {
        try {
            const javaInfo = await this.apiCall('/api/java/detect');
            this.updateJavaStatus(javaInfo);
            this.displayJavaInfo(javaInfo);
            return javaInfo;
        } catch (error) {
            this.updateJavaStatus(null, error.message);
            throw error;
        }
    }

    updateJavaStatus(javaInfo, error = null) {
        const statusElement = document.getElementById('java-status');
        const statusText = statusElement.querySelector('.status-text');

        if (error) {
            statusText.textContent = `Java检测失败: ${error}`;
            statusElement.style.background = 'rgba(220, 53, 69, 0.1)';
        } else if (javaInfo && javaInfo.length > 0) {
            const validJava = javaInfo.find(info => info.is_valid);
            if (validJava) {
                statusText.textContent = `Java ${validJava.version} (${validJava.vendor})`;
                statusElement.style.background = 'rgba(40, 167, 69, 0.1)';
            } else {
                statusText.textContent = '未找到有效的Java环境';
                statusElement.style.background = 'rgba(255, 193, 7, 0.1)';
            }
        } else {
            statusText.textContent = '未检测到Java环境';
            statusElement.style.background = 'rgba(220, 53, 69, 0.1)';
        }
    }

    displayJavaInfo(javaInfo) {
        const javaInfoDiv = document.getElementById('java-info');
        const javaList = document.getElementById('java-list');

        if (!javaInfo || javaInfo.length === 0) {
            javaInfoDiv.style.display = 'none';
            return;
        }

        javaList.innerHTML = '';
        javaInfo.forEach(info => {
            const item = document.createElement('div');
            item.className = `java-item ${info.is_valid ? 'valid' : 'invalid'}`;
            item.innerHTML = `
                <div class="java-details">
                    <strong>${info.version}</strong> - ${info.vendor}
                    <br>
                    <small>${info.path}</small>
                    ${info.error ? `<br><span class="error">${info.error}</span>` : ''}
                </div>
                <button class="btn btn-sm btn-secondary" onclick="app.selectJava('${info.path}')">
                    选择
                </button>
            `;
            javaList.appendChild(item);
        });

        javaInfoDiv.style.display = 'block';
    }

    selectJava(javaPath) {
        document.getElementById('java-path').value = javaPath;
    }

    async loadJavaConfig() {
        try {
            const config = await this.apiCall('/api/java/config');
            document.getElementById('java-path').value = config.java_path || '';
            document.getElementById('auto-detect-java').checked = config.auto_detect;
        } catch (error) {
            console.error('加载Java配置失败:', error);
        }
    }

    async saveJavaConfig() {
        try {
            const javaPath = document.getElementById('java-path').value;
            const autoDetect = document.getElementById('auto-detect-java').checked;

            await this.apiCall('/api/java/config', {
                method: 'POST',
                body: JSON.stringify({
                    java_path: javaPath,
                    auto_detect: autoDetect
                })
            });

            this.showNotification('Java配置已保存', 'success');

            // 重新检测Java环境
            if (autoDetect) {
                await this.detectJava();
            }
        } catch (error) {
            this.showNotification('保存Java配置失败', 'error');
        }
    }

    // 核心下载相关方法
    async loadCoresList() {
        try {
            const cores = await this.apiCall('/api/cores/list');
            this.populateCoreTypeSelect(cores);
        } catch (error) {
            console.error('加载核心列表失败:', error);
            this.showNotification('加载核心列表失败', 'error');
        }
    }

    populateCoreTypeSelect(cores) {
        const coreTypeSelect = document.getElementById('core-type');
        const serverCoreTypeSelect = document.getElementById('server-core-type');

        // 清空现有选项
        coreTypeSelect.innerHTML = '<option value="">选择核心类型...</option>';
        serverCoreTypeSelect.innerHTML = '<option value="">选择核心类型...</option>';

        cores.forEach(core => {
            const option = document.createElement('option');
            option.value = core.name;
            option.textContent = `${core.name}${core.recommend ? ' (推荐)' : ''}`;

            coreTypeSelect.appendChild(option.cloneNode(true));
            serverCoreTypeSelect.appendChild(option);
        });
    }

    async onCoreTypeChanged(coreType) {
        const mcVersionSelect = document.getElementById('mc-version');
        const coreVersionSelect = document.getElementById('core-version');

        if (!coreType) {
            mcVersionSelect.disabled = true;
            coreVersionSelect.disabled = true;
            mcVersionSelect.innerHTML = '<option value="">请先选择核心类型</option>';
            coreVersionSelect.innerHTML = '<option value="">请先选择MC版本</option>';
            return;
        }

        try {
            const versions = await this.apiCall(`/api/cores/versions?type=${coreType}`);

            mcVersionSelect.innerHTML = '<option value="">选择MC版本...</option>';
            versions.mc_versions.forEach(version => {
                const option = document.createElement('option');
                option.value = version;
                option.textContent = version;
                mcVersionSelect.appendChild(option);
            });

            mcVersionSelect.disabled = false;
            coreVersionSelect.disabled = true;
            coreVersionSelect.innerHTML = '<option value="">请先选择MC版本</option>';
        } catch (error) {
            console.error('加载MC版本失败:', error);
            this.showNotification('加载MC版本失败', 'error');
        }
    }

    async onMCVersionChanged(mcVersion) {
        const coreType = document.getElementById('core-type').value;
        const coreVersionSelect = document.getElementById('core-version');
        const downloadBtn = document.getElementById('download-btn');

        if (!mcVersion || !coreType) {
            coreVersionSelect.disabled = true;
            downloadBtn.disabled = true;
            return;
        }

        try {
            // 这里需要调用获取构建版本的API
            // 暂时使用模拟数据
            coreVersionSelect.innerHTML = '<option value="">选择核心版本...</option>';

            // 添加一些示例版本
            const versions = ['latest', 'build-100', 'build-99', 'build-98'];
            versions.forEach(version => {
                const option = document.createElement('option');
                option.value = version;
                option.textContent = version;
                coreVersionSelect.appendChild(option);
            });

            coreVersionSelect.disabled = false;
            coreVersionSelect.addEventListener('change', () => {
                downloadBtn.disabled = !coreVersionSelect.value;
            });
        } catch (error) {
            console.error('加载核心版本失败:', error);
            this.showNotification('加载核心版本失败', 'error');
        }
    }

    async startDownload() {
        const coreType = document.getElementById('core-type').value;
        const mcVersion = document.getElementById('mc-version').value;
        const coreVersion = document.getElementById('core-version').value;

        if (!coreType || !mcVersion || !coreVersion) {
            this.showNotification('请选择完整的核心信息', 'warning');
            return;
        }

        try {
            const result = await this.apiCall('/api/cores/download', {
                method: 'POST',
                body: JSON.stringify({
                    core_type: coreType,
                    mc_version: mcVersion,
                    core_version: coreVersion
                })
            });

            this.showNotification('下载任务已开始', 'success');
            this.showDownloadProgress(result.task_id);
        } catch (error) {
            this.showNotification('启动下载失败', 'error');
        }
    }

    showDownloadProgress(taskId) {
        const progressDiv = document.getElementById('download-progress');
        progressDiv.style.display = 'block';

        // 这里可以添加轮询逻辑来更新下载进度
        // 或者通过WebSocket接收进度更新
    }

    handleDownloadProgress(message) {
        const progressDiv = document.getElementById('download-progress');
        const progressFill = progressDiv.querySelector('.progress-fill');
        const progressPercent = progressDiv.querySelector('.progress-percent');
        const downloadSpeed = progressDiv.querySelector('.download-speed');
        const downloadSize = progressDiv.querySelector('.download-size');

        if (message.data) {
            const data = message.data;
            progressFill.style.width = `${data.progress}%`;
            progressPercent.textContent = `${Math.round(data.progress)}%`;

            if (data.speed) {
                downloadSpeed.textContent = this.formatBytes(data.speed) + '/s';
            }

            if (data.downloaded && data.total) {
                downloadSize.textContent = `${this.formatBytes(data.downloaded)} / ${this.formatBytes(data.total)}`;
            }

            if (data.status === 'completed') {
                this.showNotification('下载完成', 'success');
                setTimeout(() => {
                    progressDiv.style.display = 'none';
                }, 3000);
            } else if (data.status === 'failed') {
                this.showNotification(`下载失败: ${data.error}`, 'error');
                progressDiv.style.display = 'none';
            }
        }
    }

    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // 服务器管理相关方法
    async loadServers() {
        try {
            const servers = await this.apiCall('/api/servers');
            this.displayServers(servers);
        } catch (error) {
            console.error('加载服务器列表失败:', error);
            this.showNotification('加载服务器列表失败', 'error');
        }
    }

    displayServers(servers) {
        const serverList = document.getElementById('server-list');

        if (!servers || Object.keys(servers).length === 0) {
            serverList.innerHTML = `
                <div class="empty-state">
                    <h3>还没有服务器</h3>
                    <p>点击"创建服务器"按钮来创建您的第一个Minecraft服务器</p>
                </div>
            `;
            return;
        }

        serverList.innerHTML = '';
        Object.values(servers).forEach(server => {
            const serverCard = this.createServerCard(server);
            serverList.appendChild(serverCard);
            this.servers.set(server.id, server);
        });
    }

    createServerCard(server) {
        const card = document.createElement('div');
        card.className = 'server-card';
        card.innerHTML = `
            <div class="server-header">
                <div class="server-name">${server.name}</div>
                <span class="badge ${server.status}">${this.getStatusText(server.status)}</span>
            </div>
            <div class="server-info">
                <div>核心: ${server.core_type}</div>
                <div>版本: ${server.mc_version}</div>
                <div>内存: ${server.memory}MB</div>
                <div>端口: ${server.port}</div>
            </div>
            <div class="server-actions">
                <button class="btn btn-sm btn-primary" onclick="app.openConsole('${server.id}')">
                    控制台
                </button>
                ${server.status === 'stopped' ?
                    `<button class="btn btn-sm btn-success" onclick="app.startServer('${server.id}')">启动</button>` :
                    `<button class="btn btn-sm btn-danger" onclick="app.stopServer('${server.id}')">停止</button>`
                }
                <button class="btn btn-sm btn-secondary" onclick="app.editServer('${server.id}')">
                    编辑
                </button>
                <button class="btn btn-sm btn-danger" onclick="app.deleteServer('${server.id}')">
                    删除
                </button>
            </div>
        `;
        return card;
    }

    getStatusText(status) {
        const statusMap = {
            'running': '运行中',
            'stopped': '已停止',
            'crashed': '已崩溃',
            'starting': '启动中',
            'stopping': '停止中'
        };
        return statusMap[status] || status;
    }

    async showCreateServerModal() {
        // 加载核心列表
        await this.loadCoresList();

        // 显示模态框
        document.getElementById('create-server-modal').classList.add('show');
    }

    hideCreateServerModal() {
        document.getElementById('create-server-modal').classList.remove('show');

        // 重置表单
        document.getElementById('create-server-form').reset();
        document.getElementById('server-mc-version').disabled = true;
        document.getElementById('server-mc-version').innerHTML = '<option value="">请先选择核心类型</option>';
    }

    async onServerCoreTypeChanged(coreType) {
        const mcVersionSelect = document.getElementById('server-mc-version');

        if (!coreType) {
            mcVersionSelect.disabled = true;
            mcVersionSelect.innerHTML = '<option value="">请先选择核心类型</option>';
            return;
        }

        try {
            const versions = await this.apiCall(`/api/cores/versions?type=${coreType}`);

            mcVersionSelect.innerHTML = '<option value="">选择MC版本...</option>';
            versions.mc_versions.forEach(version => {
                const option = document.createElement('option');
                option.value = version;
                option.textContent = version;
                mcVersionSelect.appendChild(option);
            });

            mcVersionSelect.disabled = false;
        } catch (error) {
            console.error('加载MC版本失败:', error);
            this.showNotification('加载MC版本失败', 'error');
        }
    }

    async createServer() {
        const formData = new FormData(document.getElementById('create-server-form'));
        const serverData = {
            name: formData.get('server-name') || document.getElementById('server-name').value,
            core_type: formData.get('server-core-type') || document.getElementById('server-core-type').value,
            mc_version: formData.get('server-mc-version') || document.getElementById('server-mc-version').value,
            memory: parseInt(formData.get('server-memory') || document.getElementById('server-memory').value),
            port: parseInt(formData.get('server-port') || document.getElementById('server-port').value)
        };

        // 验证数据
        if (!serverData.name || !serverData.core_type || !serverData.mc_version) {
            this.showNotification('请填写完整的服务器信息', 'warning');
            return;
        }

        try {
            const result = await this.apiCall('/api/servers', {
                method: 'POST',
                body: JSON.stringify(serverData)
            });

            this.showNotification('服务器创建成功', 'success');
            this.hideCreateServerModal();
            this.loadServers(); // 重新加载服务器列表
        } catch (error) {
            this.showNotification('创建服务器失败', 'error');
        }
    }

    // 服务器控制方法
    async startServer(serverId) {
        try {
            await this.apiCall(`/api/servers/${serverId}/start`, {
                method: 'POST'
            });
            this.showNotification('服务器启动中...', 'success');
            this.updateServerStatus(serverId, 'starting');
        } catch (error) {
            this.showNotification('启动服务器失败', 'error');
        }
    }

    async stopServer(serverId) {
        try {
            await this.apiCall(`/api/servers/${serverId}/stop`, {
                method: 'POST'
            });
            this.showNotification('服务器停止中...', 'success');
            this.updateServerStatus(serverId, 'stopping');
        } catch (error) {
            this.showNotification('停止服务器失败', 'error');
        }
    }

    async restartServer(serverId) {
        try {
            await this.apiCall(`/api/servers/${serverId}/restart`, {
                method: 'POST'
            });
            this.showNotification('服务器重启中...', 'success');
            this.updateServerStatus(serverId, 'stopping');
        } catch (error) {
            this.showNotification('重启服务器失败', 'error');
        }
    }

    async deleteServer(serverId) {
        if (!confirm('确定要删除这个服务器吗？此操作不可撤销。')) {
            return;
        }

        try {
            await this.apiCall(`/api/servers/${serverId}`, {
                method: 'DELETE'
            });
            this.showNotification('服务器已删除', 'success');
            this.loadServers(); // 重新加载服务器列表
        } catch (error) {
            this.showNotification('删除服务器失败', 'error');
        }
    }

    editServer(serverId) {
        // TODO: 实现服务器编辑功能
        this.showNotification('编辑功能正在开发中', 'info');
    }

    updateServerStatus(serverId, status) {
        const server = this.servers.get(serverId);
        if (server) {
            server.status = status;
            this.servers.set(serverId, server);

            // 更新UI中的状态显示
            const serverCards = document.querySelectorAll('.server-card');
            serverCards.forEach(card => {
                const actions = card.querySelector('.server-actions');
                if (actions && actions.innerHTML.includes(serverId)) {
                    const badge = card.querySelector('.badge');
                    badge.className = `badge ${status}`;
                    badge.textContent = this.getStatusText(status);

                    // 更新按钮状态
                    this.updateServerCardActions(card, serverId, status);
                }
            });
        }
    }

    updateServerCardActions(card, serverId, status) {
        const actions = card.querySelector('.server-actions');
        const consoleBtn = `<button class="btn btn-sm btn-primary" onclick="app.openConsole('${serverId}')">控制台</button>`;
        const editBtn = `<button class="btn btn-sm btn-secondary" onclick="app.editServer('${serverId}')">编辑</button>`;
        const deleteBtn = `<button class="btn btn-sm btn-danger" onclick="app.deleteServer('${serverId}')">删除</button>`;

        let controlBtn = '';
        if (status === 'stopped' || status === 'crashed') {
            controlBtn = `<button class="btn btn-sm btn-success" onclick="app.startServer('${serverId}')">启动</button>`;
        } else if (status === 'running') {
            controlBtn = `<button class="btn btn-sm btn-danger" onclick="app.stopServer('${serverId}')">停止</button>`;
        } else {
            controlBtn = `<button class="btn btn-sm btn-secondary" disabled>处理中...</button>`;
        }

        actions.innerHTML = consoleBtn + controlBtn + editBtn + deleteBtn;
    }

    // 控制台相关方法
    openConsole(serverId) {
        this.currentServerId = serverId;
        const server = this.servers.get(serverId);

        if (!server) {
            this.showNotification('服务器不存在', 'error');
            return;
        }

        // 更新控制台标题
        document.querySelector('.modal-title .server-name').textContent = server.name;
        document.querySelector('.modal-title .server-status').textContent = this.getStatusText(server.status);
        document.querySelector('.modal-title .server-status').className = `badge ${server.status}`;

        // 更新控制按钮状态
        this.updateConsoleControls(server.status);

        // 清空控制台输出
        document.getElementById('console-output').innerHTML = '<div class="console-welcome">连接到服务器控制台...</div>';

        // 订阅日志
        this.wsManager.subscribeToLogs(serverId);

        // 显示控制台
        document.getElementById('console-modal').classList.add('show');

        // 获取服务器状态
        this.wsManager.getServerStatus(serverId);
    }

    closeConsole() {
        if (this.currentServerId) {
            this.wsManager.unsubscribeFromLogs(this.currentServerId);
            this.currentServerId = null;
        }

        document.getElementById('console-modal').classList.remove('show');
    }

    updateConsoleControls(status) {
        const startBtn = document.getElementById('start-server');
        const stopBtn = document.getElementById('stop-server');
        const restartBtn = document.getElementById('restart-server');
        const commandInput = document.getElementById('command-input');
        const sendCommandBtn = document.getElementById('send-command');

        const isRunning = status === 'running';
        const isStopped = status === 'stopped' || status === 'crashed';
        const isProcessing = status === 'starting' || status === 'stopping';

        startBtn.disabled = !isStopped;
        stopBtn.disabled = !isRunning;
        restartBtn.disabled = !isRunning;
        commandInput.disabled = !isRunning;
        sendCommandBtn.disabled = !isRunning;

        if (isProcessing) {
            startBtn.disabled = true;
            stopBtn.disabled = true;
            restartBtn.disabled = true;
        }
    }

    // WebSocket消息处理
    handleLogMessage(message) {
        if (message.server_id !== this.currentServerId) {
            return;
        }

        const consoleOutput = document.getElementById('console-output');
        const logData = message.data;

        // 创建日志行
        const logLine = document.createElement('div');
        logLine.className = `console-line log-${logData.level.toLowerCase()}`;

        // 格式化时间戳
        const timestamp = new Date(logData.timestamp).toLocaleTimeString();

        logLine.innerHTML = `
            <span class="log-timestamp">[${timestamp}]</span>
            <span class="log-level">[${logData.level}]</span>
            <span class="log-message">${this.escapeHtml(logData.message)}</span>
        `;

        consoleOutput.appendChild(logLine);

        // 自动滚动
        if (document.getElementById('auto-scroll').checked) {
            consoleOutput.scrollTop = consoleOutput.scrollHeight;
        }

        // 限制日志行数
        const maxLines = 1000;
        while (consoleOutput.children.length > maxLines) {
            consoleOutput.removeChild(consoleOutput.firstChild);
        }
    }

    handleServerStatus(message) {
        if (message.server_id) {
            const statusData = message.data;
            this.updateServerStatus(message.server_id, statusData.status);

            // 如果是当前控制台的服务器，更新控制台状态
            if (message.server_id === this.currentServerId) {
                document.querySelector('.modal-title .server-status').textContent = this.getStatusText(statusData.status);
                document.querySelector('.modal-title .server-status').className = `badge ${statusData.status}`;
                this.updateConsoleControls(statusData.status);
            }
        }
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}

// 工具函数
function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 全局应用实例
let app = null;

// 页面加载完成后初始化应用
document.addEventListener('DOMContentLoaded', () => {
    // 防止重复初始化
    if (app) {
        console.log('应用已存在，跳过重复初始化');
        return;
    }

    console.log('正在初始化EasilyPanel应用...');
    app = new EasilyPanel();

    // 将app实例暴露到全局作用域，供HTML中的onclick使用
    window.app = app;

    console.log('EasilyPanel应用已启动');
});

// 页面卸载时清理资源
window.addEventListener('beforeunload', () => {
    console.log('页面卸载，清理资源...');
    if (app && app.wsManager) {
        app.wsManager.destroy();
    }
    app = null;
});

// 页面隐藏时暂停WebSocket
document.addEventListener('visibilitychange', () => {
    if (app && app.wsManager) {
        if (document.hidden) {
            console.log('页面隐藏，暂停WebSocket');
            // 不关闭连接，只是停止重连
            app.wsManager.maxReconnectAttempts = 0;
        } else {
            console.log('页面显示，恢复WebSocket');
            app.wsManager.maxReconnectAttempts = 5;
            if (!app.wsManager.isConnected()) {
                app.wsManager.connect();
            }
        }
    }
});
