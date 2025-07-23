// 可复用的文件管理组件
class FileManagerComponent {
    constructor(containerId, options = {}) {
        this.containerId = containerId;
        this.options = {
            initialPath: '/',
            showToolbar: true,
            showBreadcrumb: true,
            allowUpload: true,
            allowDelete: true,
            allowEdit: true,
            apiEndpoint: '/api/files',
            ...options
        };
        
        this.currentPath = this.options.initialPath;
        this.files = [];
        this.selectedFiles = new Set();
        this.viewMode = 'list'; // list 或 grid
        this.sortBy = 'name'; // name, size, modified
        this.sortOrder = 'asc'; // asc 或 desc
        this.searchQuery = '';
        this.clipboard = null;
        this.pathHistory = [this.currentPath];
        this.historyIndex = 0;
    }
    
    // 初始化组件
    async init() {
        this.render();
        this.setupEventListeners();
        await this.loadFiles();
    }
    
    // 销毁组件
    destroy() {
        const container = document.getElementById(this.containerId);
        if (container) {
            container.innerHTML = '';
        }
    }
    
    // 渲染组件
    render() {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        container.innerHTML = `
            <div class="file-manager-component">
                ${this.options.showToolbar ? this.renderToolbar() : ''}
                ${this.options.showBreadcrumb ? this.renderBreadcrumb() : ''}
                ${this.renderFilesList()}
            </div>
        `;
    }
    
    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="file-manager-toolbar">
                <div class="toolbar-left">
                    <div class="navigation-buttons">
                        <button class="btn btn-sm" id="fm-back-btn" title="后退" ${this.historyIndex <= 0 ? 'disabled' : ''}>
                            <i class="mdi mdi-arrow-left"></i>
                        </button>
                        <button class="btn btn-sm" id="fm-forward-btn" title="前进" ${this.historyIndex >= this.pathHistory.length - 1 ? 'disabled' : ''}>
                            <i class="mdi mdi-arrow-right"></i>
                        </button>
                        <button class="btn btn-sm" id="fm-up-btn" title="上级目录" ${this.currentPath === '/' ? 'disabled' : ''}>
                            <i class="mdi mdi-arrow-up"></i>
                        </button>
                        <button class="btn btn-sm" id="fm-refresh-btn" title="刷新">
                            <i class="mdi mdi-refresh"></i>
                        </button>
                    </div>
                    
                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="fm-search" placeholder="搜索文件..." value="${this.searchQuery}">
                    </div>
                    
                    <div class="view-controls">
                        <button class="view-btn ${this.viewMode === 'list' ? 'active' : ''}" data-view="list" title="列表视图">
                            <i class="mdi mdi-view-list"></i>
                        </button>
                        <button class="view-btn ${this.viewMode === 'grid' ? 'active' : ''}" data-view="grid" title="网格视图">
                            <i class="mdi mdi-view-grid"></i>
                        </button>
                    </div>
                </div>
                
                <div class="toolbar-right">
                    <div class="batch-actions" id="fm-batch-actions" style="display: none;">
                        <span class="selected-count" id="fm-selected-count">已选择 0 个文件</span>
                        <button class="btn btn-sm" id="fm-batch-copy">
                            <i class="mdi mdi-content-copy"></i>
                            <span>复制</span>
                        </button>
                        <button class="btn btn-sm" id="fm-batch-cut">
                            <i class="mdi mdi-content-cut"></i>
                            <span>剪切</span>
                        </button>
                        ${this.options.allowDelete ? `
                            <button class="btn btn-sm warning" id="fm-batch-delete">
                                <i class="mdi mdi-delete"></i>
                                <span>删除</span>
                            </button>
                        ` : ''}
                    </div>
                    
                    <div class="action-buttons">
                        <button class="btn btn-sm" id="fm-paste-btn" style="display: none;">
                            <i class="mdi mdi-content-paste"></i>
                            <span>粘贴</span>
                        </button>
                        
                        ${this.options.allowUpload ? `
                            <button class="btn btn-sm primary" id="fm-upload-btn">
                                <i class="mdi mdi-upload"></i>
                                <span>上传</span>
                            </button>
                        ` : ''}
                        
                        <div class="dropdown">
                            <button class="btn btn-sm dropdown-toggle" id="fm-create-btn">
                                <i class="mdi mdi-plus"></i>
                                <span>新建</span>
                            </button>
                            <div class="dropdown-menu">
                                <a href="#" class="dropdown-item" data-action="create-folder">
                                    <i class="mdi mdi-folder-plus"></i>
                                    <span>新建文件夹</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="create-file">
                                    <i class="mdi mdi-file-plus"></i>
                                    <span>新建文件</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染面包屑导航
    renderBreadcrumb() {
        const pathParts = this.currentPath.split('/').filter(part => part);
        const breadcrumbs = [
            { name: '根目录', path: '/' }
        ];
        
        let currentPath = '';
        pathParts.forEach(part => {
            currentPath += '/' + part;
            breadcrumbs.push({
                name: part,
                path: currentPath
            });
        });
        
        return `
            <div class="file-manager-breadcrumb">
                <nav class="breadcrumb">
                    ${breadcrumbs.map((crumb, index) => `
                        <span class="breadcrumb-item ${index === breadcrumbs.length - 1 ? 'active' : ''}" 
                              data-path="${crumb.path}">
                            ${index === 0 ? '<i class="mdi mdi-home"></i>' : ''}
                            <span>${crumb.name}</span>
                        </span>
                        ${index < breadcrumbs.length - 1 ? '<i class="mdi mdi-chevron-right breadcrumb-separator"></i>' : ''}
                    `).join('')}
                </nav>
                
                <div class="path-info">
                    <span class="current-path">${this.currentPath}</span>
                </div>
            </div>
        `;
    }
    
    // 渲染文件列表
    renderFilesList() {
        return `
            <div class="file-manager-list-container">
                <div class="file-manager-list ${this.viewMode}" id="fm-files-list">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载文件列表...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        // 导航按钮
        const backBtn = container.querySelector('#fm-back-btn');
        const forwardBtn = container.querySelector('#fm-forward-btn');
        const upBtn = container.querySelector('#fm-up-btn');
        const refreshBtn = container.querySelector('#fm-refresh-btn');
        
        if (backBtn) backBtn.addEventListener('click', () => this.goBack());
        if (forwardBtn) forwardBtn.addEventListener('click', () => this.goForward());
        if (upBtn) upBtn.addEventListener('click', () => this.goUp());
        if (refreshBtn) refreshBtn.addEventListener('click', () => this.loadFiles());
        
        // 搜索框
        const searchInput = container.querySelector('#fm-search');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderFiles();
            });
        }
        
        // 视图切换
        const viewBtns = container.querySelectorAll('.view-btn');
        viewBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const view = e.currentTarget.getAttribute('data-view');
                this.setViewMode(view);
            });
        });
        
        // 批量操作按钮
        const batchCopyBtn = container.querySelector('#fm-batch-copy');
        const batchCutBtn = container.querySelector('#fm-batch-cut');
        const batchDeleteBtn = container.querySelector('#fm-batch-delete');
        const pasteBtn = container.querySelector('#fm-paste-btn');
        
        if (batchCopyBtn) batchCopyBtn.addEventListener('click', () => this.copySelectedFiles());
        if (batchCutBtn) batchCutBtn.addEventListener('click', () => this.cutSelectedFiles());
        if (batchDeleteBtn) batchDeleteBtn.addEventListener('click', () => this.deleteSelectedFiles());
        if (pasteBtn) pasteBtn.addEventListener('click', () => this.pasteFiles());
        
        // 上传按钮
        const uploadBtn = container.querySelector('#fm-upload-btn');
        if (uploadBtn) {
            uploadBtn.addEventListener('click', () => this.showUploadDialog());
        }
        
        // 新建按钮下拉菜单
        const dropdownItems = container.querySelectorAll('.dropdown-item');
        dropdownItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const action = e.currentTarget.getAttribute('data-action');
                this.handleCreateAction(action);
            });
        });
        
        // 面包屑导航
        this.bindBreadcrumbEvents();
        
        // 文件拖拽上传
        if (this.options.allowUpload) {
            this.setupDragAndDrop();
        }
    }
    
    // 绑定面包屑事件
    bindBreadcrumbEvents() {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        const breadcrumbItems = container.querySelectorAll('.breadcrumb-item:not(.active)');
        breadcrumbItems.forEach(item => {
            item.addEventListener('click', (e) => {
                const path = e.currentTarget.getAttribute('data-path');
                this.navigateToPath(path);
            });
        });
    }
    
    // 设置拖拽上传
    setupDragAndDrop() {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        container.addEventListener('dragover', (e) => {
            e.preventDefault();
            container.classList.add('drag-over');
        });
        
        container.addEventListener('dragleave', (e) => {
            if (!container.contains(e.relatedTarget)) {
                container.classList.remove('drag-over');
            }
        });
        
        container.addEventListener('drop', (e) => {
            e.preventDefault();
            container.classList.remove('drag-over');
            
            const files = Array.from(e.dataTransfer.files);
            if (files.length > 0) {
                this.uploadFiles(files);
            }
        });
    }
    
    // 设置视图模式
    setViewMode(mode) {
        this.viewMode = mode;
        
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        // 更新按钮状态
        container.querySelectorAll('.view-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        container.querySelector(`[data-view="${mode}"]`).classList.add('active');
        
        // 更新列表样式
        const filesList = container.querySelector('#fm-files-list');
        if (filesList) {
            filesList.className = `file-manager-list ${mode}`;
        }
        
        this.filterAndRenderFiles();
    }
    
    // 导航到路径
    navigateToPath(path) {
        // 添加到历史记录
        if (this.historyIndex < this.pathHistory.length - 1) {
            this.pathHistory = this.pathHistory.slice(0, this.historyIndex + 1);
        }
        
        if (path !== this.currentPath) {
            this.pathHistory.push(path);
            this.historyIndex = this.pathHistory.length - 1;
        }
        
        this.currentPath = path;
        this.selectedFiles.clear();
        this.loadFiles();
    }
    
    // 后退
    goBack() {
        if (this.historyIndex > 0) {
            this.historyIndex--;
            this.currentPath = this.pathHistory[this.historyIndex];
            this.selectedFiles.clear();
            this.loadFiles();
        }
    }
    
    // 前进
    goForward() {
        if (this.historyIndex < this.pathHistory.length - 1) {
            this.historyIndex++;
            this.currentPath = this.pathHistory[this.historyIndex];
            this.selectedFiles.clear();
            this.loadFiles();
        }
    }
    
    // 上级目录
    goUp() {
        const parentPath = this.currentPath.split('/').slice(0, -1).join('/') || '/';
        this.navigateToPath(parentPath);
    }
    
    // 加载文件列表
    async loadFiles() {
        try {
            const url = `${this.options.apiEndpoint}?path=${encodeURIComponent(this.currentPath)}&sort_by=${this.sortBy}&order=${this.sortOrder}`;
            const response = await fetch(url);
            const result = await response.json();

            if (response.ok && result.success) {
                this.files = result.data.files || [];
                this.filterAndRenderFiles();
                this.updateToolbarState();
            } else {
                const errorMessage = result.message || result.error || '加载文件列表失败';
                this.showError(errorMessage);
            }
        } catch (error) {
            console.error('Failed to load files:', error);
            this.showError('网络错误，请重试');
        }
    }
    
    // 更新工具栏状态
    updateToolbarState() {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        // 更新导航按钮状态
        const backBtn = container.querySelector('#fm-back-btn');
        const forwardBtn = container.querySelector('#fm-forward-btn');
        const upBtn = container.querySelector('#fm-up-btn');
        
        if (backBtn) backBtn.disabled = this.historyIndex <= 0;
        if (forwardBtn) forwardBtn.disabled = this.historyIndex >= this.pathHistory.length - 1;
        if (upBtn) upBtn.disabled = this.currentPath === '/';
        
        // 重新渲染面包屑
        if (this.options.showBreadcrumb) {
            const breadcrumbContainer = container.querySelector('.file-manager-breadcrumb');
            if (breadcrumbContainer) {
                breadcrumbContainer.innerHTML = this.renderBreadcrumb().replace(/<div class="file-manager-breadcrumb">|<\/div>$/g, '');
                this.bindBreadcrumbEvents();
            }
        }
    }
    
    // 过滤和渲染文件
    filterAndRenderFiles() {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        const filesList = container.querySelector('#fm-files-list');
        if (!filesList) return;
        
        // 获取过滤后的文件
        let filteredFiles = [...this.files];
        
        // 应用搜索过滤
        if (this.searchQuery) {
            const query = this.searchQuery.toLowerCase();
            filteredFiles = filteredFiles.filter(file => 
                file.name.toLowerCase().includes(query)
            );
        }
        
        // 排序
        filteredFiles.sort((a, b) => {
            // 文件夹优先
            if (a.is_dir !== b.is_dir) {
                return a.is_dir ? -1 : 1;
            }
            
            let result = 0;
            switch (this.sortBy) {
                case 'name':
                    result = a.name.localeCompare(b.name);
                    break;
                case 'size':
                    result = (a.size || 0) - (b.size || 0);
                    break;
                case 'modified':
                    result = new Date(a.modified_at) - new Date(b.modified_at);
                    break;
            }
            
            return this.sortOrder === 'desc' ? -result : result;
        });
        
        // 渲染文件
        if (filteredFiles.length === 0) {
            filesList.innerHTML = this.renderEmptyState();
        } else {
            if (this.viewMode === 'list') {
                filesList.innerHTML = this.renderFilesTable(filteredFiles);
            } else {
                filesList.innerHTML = filteredFiles.map(file => 
                    this.renderFileCard(file)
                ).join('');
            }
        }
        
        // 重新绑定事件
        this.bindFileEvents();
    }
    
    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-folder-open-outline"></i>
                <h4>文件夹为空</h4>
                <p>此文件夹中没有文件，或者搜索条件没有匹配的结果。</p>
                ${this.options.allowUpload ? `
                    <button class="btn primary" onclick="document.getElementById('${this.containerId}').querySelector('#fm-upload-btn').click()">
                        <i class="mdi mdi-upload"></i>
                        <span>上传文件</span>
                    </button>
                ` : ''}
            </div>
        `;
    }
    
    // 显示错误信息
    showError(message) {
        const container = document.getElementById(this.containerId);
        if (!container) return;
        
        const filesList = container.querySelector('#fm-files-list');
        if (filesList) {
            filesList.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h4>加载失败</h4>
                    <p>${message}</p>
                    <button class="btn primary" onclick="document.getElementById('${this.containerId}').fileManager.loadFiles()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 其他方法的占位符
    renderFilesTable(files) {
        // 实现文件表格渲染
        return '<div>文件表格视图</div>';
    }
    
    renderFileCard(file) {
        // 实现文件卡片渲染
        return '<div>文件卡片视图</div>';
    }
    
    bindFileEvents() {
        // 实现文件事件绑定
    }
    
    copySelectedFiles() {
        console.log('Copy selected files');
    }
    
    cutSelectedFiles() {
        console.log('Cut selected files');
    }
    
    deleteSelectedFiles() {
        console.log('Delete selected files');
    }
    
    pasteFiles() {
        console.log('Paste files');
    }
    
    showUploadDialog() {
        console.log('Show upload dialog');
    }
    
    // 处理创建操作
    handleCreateAction(action) {
        const uiManager = window.getUIManager();

        if (action === 'folder') {
            this.showCreateDialog(true);
        } else if (action === 'file') {
            this.showCreateDialog(false);
        }
    }

    // 显示创建对话框
    showCreateDialog(isDir) {
        const uiManager = window.getUIManager();
        const type = isDir ? '文件夹' : '文件';

        const content = `
            <div class="create-dialog">
                <div class="form-group">
                    <label for="itemName">${type}名称</label>
                    <input type="text" id="itemName" class="form-input" placeholder="输入${type}名称" required>
                </div>
            </div>
        `;

        const modal = uiManager?.showModal(`创建${type}`, content, {
            width: '400px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn primary" id="confirmCreateBtn">
                    <i class="mdi mdi-check"></i>
                    <span>创建</span>
                </button>
            `
        });

        if (modal) {
            const confirmBtn = modal.querySelector('#confirmCreateBtn');
            const nameInput = modal.querySelector('#itemName');

            confirmBtn.addEventListener('click', () => {
                const name = nameInput.value.trim();
                if (name) {
                    this.createItem(name, isDir);
                }
            });

            nameInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    confirmBtn.click();
                }
            });

            nameInput.focus();
        }
    }

    // 创建文件或文件夹
    async createItem(name, isDir) {
        const uiManager = window.getUIManager();

        try {
            const path = this.currentPath === '/' ? name : `${this.currentPath}/${name}`;

            const response = await fetch(this.options.apiEndpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    action: 'create',
                    path: path,
                    is_dir: isDir
                })
            });

            const result = await response.json();

            if (response.ok && result.success) {
                uiManager?.closeModal();
                uiManager?.showNotification('创建成功', `${isDir ? '文件夹' : '文件'}创建成功`, 'success');
                await this.loadFiles();
            } else {
                const errorMessage = result.message || result.error || '创建失败';
                uiManager?.showNotification('创建失败', errorMessage, 'error');
            }
        } catch (error) {
            console.error('Create item failed:', error);
            uiManager?.showNotification('创建失败', '网络错误，请重试', 'error');
        }
    }

    // 上传文件
    async uploadFiles(files) {
        const uiManager = window.getUIManager();

        for (const file of files) {
            try {
                const formData = new FormData();
                formData.append('file', file);
                formData.append('path', this.currentPath);

                const response = await fetch(this.options.apiEndpoint, {
                    method: 'PUT',
                    body: formData
                });

                const result = await response.json();

                if (response.ok && result.success) {
                    uiManager?.showNotification('上传成功', `文件 ${file.name} 上传成功`, 'success');
                } else {
                    const errorMessage = result.message || result.error || '上传失败';
                    uiManager?.showNotification('上传失败', `文件 ${file.name}: ${errorMessage}`, 'error');
                }
            } catch (error) {
                console.error('Upload file failed:', error);
                uiManager?.showNotification('上传失败', `文件 ${file.name}: 网络错误`, 'error');
            }
        }

        // 刷新文件列表
        await this.loadFiles();
    }

    // 删除文件或文件夹
    async deleteItem(path) {
        const uiManager = window.getUIManager();

        try {
            const response = await fetch(this.options.apiEndpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    action: 'delete',
                    path: path
                })
            });

            const result = await response.json();

            if (response.ok && result.success) {
                uiManager?.showNotification('删除成功', '删除成功', 'success');
                await this.loadFiles();
            } else {
                const errorMessage = result.message || result.error || '删除失败';
                uiManager?.showNotification('删除失败', errorMessage, 'error');
            }
        } catch (error) {
            console.error('Delete item failed:', error);
            uiManager?.showNotification('删除失败', '网络错误，请重试', 'error');
        }
    }

    // 重命名文件或文件夹
    async renameItem(oldPath, newName) {
        const uiManager = window.getUIManager();

        try {
            const dir = oldPath.substring(0, oldPath.lastIndexOf('/'));
            const newPath = dir ? `${dir}/${newName}` : newName;

            const response = await fetch(this.options.apiEndpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    action: 'rename',
                    path: oldPath,
                    target: newPath
                })
            });

            const result = await response.json();

            if (response.ok && result.success) {
                uiManager?.showNotification('重命名成功', '重命名成功', 'success');
                await this.loadFiles();
            } else {
                const errorMessage = result.message || result.error || '重命名失败';
                uiManager?.showNotification('重命名失败', errorMessage, 'error');
            }
        } catch (error) {
            console.error('Rename item failed:', error);
            uiManager?.showNotification('重命名失败', '网络错误，请重试', 'error');
        }
    }
}

// 导出到全局作用域
window.FileManagerComponent = FileManagerComponent;
