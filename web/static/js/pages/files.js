// 文件管理页面管理器
class FilesPageManager {
    constructor() {
        this.currentPath = '/';
        this.files = [];
        this.selectedFiles = new Set();
        this.viewMode = 'list'; // 'list' or 'grid'
        this.sortBy = 'name';
        this.sortOrder = 'asc';
        this.searchQuery = '';
        this.clipboard = null;
    }

    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadFiles();
    }

    // 销毁页面
    destroy() {
        // 清理事件监听器等
        this.selectedFiles.clear();
        this.clipboard = null;
    }

    // 渲染页面
    renderPage() {
        const filesPage = document.getElementById('files-page');
        if (!filesPage) return;

        filesPage.innerHTML = `
            <div class="page-header">
                <h2>文件管理</h2>
                <p>浏览和管理服务器文件</p>
            </div>

            <div class="files-container">
                ${this.renderBreadcrumb()}
                ${this.renderToolbar()}
                ${this.renderFilesList()}
            </div>
        `;
    }
    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="files-toolbar">
                <div class="toolbar-left">
                    <div class="navigation-buttons">
                        <button class="btn" id="backBtn" title="后退">
                            <i class="mdi mdi-arrow-left"></i>
                        </button>
                        <button class="btn" id="forwardBtn" title="前进">
                            <i class="mdi mdi-arrow-right"></i>
                        </button>
                        <button class="btn" id="upBtn" title="上级目录">
                            <i class="mdi mdi-arrow-up"></i>
                        </button>
                        <button class="btn" id="refreshBtn" title="刷新">
                            <i class="mdi mdi-refresh"></i>
                        </button>
                    </div>

                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="fileSearch" placeholder="搜索文件..." value="${this.searchQuery}">
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
                    <div class="batch-actions" id="batchActions" style="display: none;">
                        <span class="selected-count" id="selectedCount">已选择 0 个文件</span>
                        <button class="btn" id="batchDeleteBtn">
                            <i class="mdi mdi-delete"></i>
                            <span>删除</span>
                        </button>
                    </div>

                    <div class="action-buttons">
                        <button class="btn primary" id="uploadBtn">
                            <i class="mdi mdi-upload"></i>
                            <span>上传文件</span>
                        </button>

                        <button class="btn" id="createFolderBtn">
                            <i class="mdi mdi-folder-plus"></i>
                            <span>新建文件夹</span>
                        </button>
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
            <div class="breadcrumb-container">
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
            <div class="files-list-container">
                <div class="files-list ${this.viewMode}" id="filesList">
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
        // 导航按钮
        const refreshBtn = document.getElementById('refreshBtn');
        const upBtn = document.getElementById('upBtn');

        if (refreshBtn) refreshBtn.addEventListener('click', () => this.loadFiles());
        if (upBtn) upBtn.addEventListener('click', () => this.goUp());

        // 搜索框
        const searchInput = document.getElementById('fileSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderFiles();
            });
        }

        // 视图切换
        const viewBtns = document.querySelectorAll('.view-btn');
        viewBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const view = e.currentTarget.getAttribute('data-view');
                this.setViewMode(view);
            });
        });

        // 批量操作按钮
        const batchDeleteBtn = document.getElementById('batchDeleteBtn');
        if (batchDeleteBtn) batchDeleteBtn.addEventListener('click', () => this.deleteSelectedFiles());

        // 操作按钮
        const uploadBtn = document.getElementById('uploadBtn');
        const createFolderBtn = document.getElementById('createFolderBtn');

        if (uploadBtn) uploadBtn.addEventListener('click', () => this.showUploadDialog());
        if (createFolderBtn) createFolderBtn.addEventListener('click', () => this.showCreateFolderDialog());

        // 面包屑导航
        this.bindBreadcrumbEvents();
    }
    
    // 渲染文件列表
    renderFilesList() {
        return `
            <div class="files-list-container">
                <div class="files-list ${this.viewMode}" id="filesList">
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
        // 导航按钮
        const backBtn = document.getElementById('backBtn');
        const forwardBtn = document.getElementById('forwardBtn');
        const upBtn = document.getElementById('upBtn');
        const refreshBtn = document.getElementById('refreshBtn');
        
        if (backBtn) backBtn.addEventListener('click', () => this.goBack());
        if (forwardBtn) forwardBtn.addEventListener('click', () => this.goForward());
        if (upBtn) upBtn.addEventListener('click', () => this.goUp());
        if (refreshBtn) refreshBtn.addEventListener('click', () => this.loadFiles());
        
        // 搜索框
        const searchInput = document.getElementById('fileSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderFiles();
            });
        }
        
        // 视图切换
        const viewBtns = document.querySelectorAll('.view-btn');
        viewBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const view = e.currentTarget.getAttribute('data-view');
                this.setViewMode(view);
            });
        });
        
        // 批量操作按钮
        const batchCopyBtn = document.getElementById('batchCopyBtn');
        const batchCutBtn = document.getElementById('batchCutBtn');
        const batchDeleteBtn = document.getElementById('batchDeleteBtn');
        const pasteBtn = document.getElementById('pasteBtn');
        
        if (batchCopyBtn) batchCopyBtn.addEventListener('click', () => this.copySelectedFiles());
        if (batchCutBtn) batchCutBtn.addEventListener('click', () => this.cutSelectedFiles());
        if (batchDeleteBtn) batchDeleteBtn.addEventListener('click', () => this.deleteSelectedFiles());
        if (pasteBtn) pasteBtn.addEventListener('click', () => this.pasteFiles());
        
        // 新建按钮下拉菜单
        const dropdownItems = document.querySelectorAll('.dropdown-item');
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
        this.setupDragAndDrop();
    }

    // 绑定面包屑事件
    bindBreadcrumbEvents() {
        const breadcrumbItems = document.querySelectorAll('.breadcrumb-item:not(.active)');
        breadcrumbItems.forEach(item => {
            item.addEventListener('click', (e) => {
                const path = e.currentTarget.getAttribute('data-path');
                this.navigateToPath(path);
            });
        });
    }

    // 设置视图模式
    setViewMode(mode) {
        this.viewMode = mode;

        // 更新按钮状态
        document.querySelectorAll('.view-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        const activeBtn = document.querySelector(`[data-view="${mode}"]`);
        if (activeBtn) activeBtn.classList.add('active');

        // 更新列表样式
        const filesList = document.getElementById('filesList');
        if (filesList) {
            filesList.className = `files-list ${mode}`;
        }

        this.filterAndRenderFiles();
    }

    // 加载文件列表
    async loadFiles() {
        try {
            const response = await fetch(`/api/files?path=${encodeURIComponent(this.currentPath)}`);
            const result = await response.json();

            if (result.success) {
                this.files = result.data || [];
                this.filterAndRenderFiles();
            } else {
                this.showError('加载文件列表失败: ' + (result.message || '未知错误'));
            }
        } catch (error) {
            console.error('Failed to load files:', error);
            this.showError('网络错误，请重试');
        }
    }

    // 过滤和渲染文件
    filterAndRenderFiles() {
        const filesList = document.getElementById('filesList');
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
                    result = new Date(a.mod_time) - new Date(b.mod_time);
                    break;
            }

            return this.sortOrder === 'desc' ? -result : result;
        });

        // 渲染文件
        if (filteredFiles.length === 0) {
            filesList.innerHTML = this.renderEmptyState();
        } else {
            filesList.innerHTML = this.renderFilesTable(filteredFiles);
        }

        // 重新绑定事件
        this.bindFileEvents();
    }

    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-folder-open-outline"></i>
                <h3>文件夹为空</h3>
                <p>此文件夹中没有文件，或者搜索条件没有匹配的结果。</p>
                <button class="btn primary" onclick="window.getFilesPageManager().showUploadDialog()">
                    <i class="mdi mdi-upload"></i>
                    <span>上传文件</span>
                </button>
            </div>
        `;
    }

    // 渲染文件表格
    renderFilesTable(files) {
        return `
            <div class="files-table">
                <div class="table-header">
                    <div class="table-cell select">
                        <input type="checkbox" id="selectAllFiles">
                    </div>
                    <div class="table-cell name">
                        <span>名称</span>
                    </div>
                    <div class="table-cell size">
                        <span>大小</span>
                    </div>
                    <div class="table-cell modified">
                        <span>修改时间</span>
                    </div>
                    <div class="table-cell actions">操作</div>
                </div>

                <div class="table-body">
                    ${files.map(file => this.renderFileRow(file)).join('')}
                </div>
            </div>
        `;
    }

    // 渲染文件行
    renderFileRow(file) {
        const fileIcon = this.getFileIcon(file);
        const fileSize = file.is_dir ? '-' : this.formatFileSize(file.size);
        const modifiedTime = new Date(file.mod_time).toLocaleString();

        return `
            <div class="table-row" data-file-path="${file.path}">
                <div class="table-cell select">
                    <input type="checkbox" class="file-checkbox" data-file-path="${file.path}">
                </div>
                <div class="table-cell name" data-action="open">
                    <div class="file-info">
                        <i class="mdi ${fileIcon}"></i>
                        <span class="file-name">${file.name}</span>
                    </div>
                </div>
                <div class="table-cell size">
                    <span>${fileSize}</span>
                </div>
                <div class="table-cell modified">
                    <span>${modifiedTime}</span>
                </div>
                <div class="table-cell actions">
                    <div class="action-buttons">
                        ${file.is_dir ? '' : `
                            <button class="btn-icon file-action-btn" data-action="download" data-file-path="${file.path}" title="下载">
                                <i class="mdi mdi-download"></i>
                            </button>
                        `}
                        <button class="btn-icon file-action-btn" data-action="delete" data-file-path="${file.path}" title="删除">
                            <i class="mdi mdi-delete"></i>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    // 绑定文件事件
    bindFileEvents() {
        // 文件打开事件
        const openElements = document.querySelectorAll('[data-action="open"]');
        openElements.forEach(element => {
            element.addEventListener('dblclick', (e) => {
                const filePath = e.currentTarget.closest('[data-file-path]').getAttribute('data-file-path');
                this.openFile(filePath);
            });
        });

        // 操作按钮事件
        const actionBtns = document.querySelectorAll('.file-action-btn');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.getAttribute('data-action');
                const filePath = e.currentTarget.getAttribute('data-file-path');
                this.handleFileAction(action, filePath);
            });
        });

        // 复选框事件
        const checkboxes = document.querySelectorAll('.file-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                const filePath = e.target.getAttribute('data-file-path');
                if (e.target.checked) {
                    this.selectedFiles.add(filePath);
                } else {
                    this.selectedFiles.delete(filePath);
                }
                this.updateBatchActions();
            });
        });
    }

    // 更新批量操作显示
    updateBatchActions() {
        const batchActions = document.getElementById('batchActions');
        const selectedCount = document.getElementById('selectedCount');

        if (this.selectedFiles.size > 0) {
            batchActions.style.display = 'flex';
            selectedCount.textContent = `已选择 ${this.selectedFiles.size} 个文件`;
        } else {
            batchActions.style.display = 'none';
        }
    }

    // 获取文件图标
    getFileIcon(file) {
        if (file.is_dir) {
            return 'mdi-folder';
        }

        const ext = file.name.split('.').pop().toLowerCase();
        const iconMap = {
            'txt': 'mdi-file-document',
            'md': 'mdi-file-document',
            'json': 'mdi-code-json',
            'xml': 'mdi-file-xml',
            'yml': 'mdi-file-code',
            'yaml': 'mdi-file-code',
            'properties': 'mdi-cog',
            'conf': 'mdi-cog',
            'cfg': 'mdi-cog',
            'ini': 'mdi-cog',
            'jar': 'mdi-minecraft',
            'zip': 'mdi-folder-zip',
            'rar': 'mdi-folder-zip',
            '7z': 'mdi-folder-zip',
            'jpg': 'mdi-file-image',
            'jpeg': 'mdi-file-image',
            'png': 'mdi-file-image',
            'gif': 'mdi-file-image'
        };

        return iconMap[ext] || 'mdi-file';
    }

    // 格式化文件大小
    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';

        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));

        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }

    // 导航到路径
    navigateToPath(path) {
        this.currentPath = path;
        this.selectedFiles.clear();
        this.loadFiles();
    }

    // 打开文件
    openFile(filePath) {
        const file = this.files.find(f => f.path === filePath);
        if (!file) return;

        if (file.is_dir) {
            this.navigateToPath(filePath);
        } else {
            // 下载文件
            this.downloadFile(filePath);
        }
    }

    // 处理文件操作
    handleFileAction(action, filePath) {
        switch (action) {
            case 'download':
                this.downloadFile(filePath);
                break;
            case 'delete':
                this.showDeleteDialog(filePath);
                break;
        }
    }

    // 显示错误信息
    showError(message) {
        const filesList = document.getElementById('filesList');
        if (filesList) {
            filesList.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h3>加载失败</h3>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getFilesPageManager().loadFiles()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 渲染文件表格
    renderFilesTable(files) {
        return `
            <div class="files-table">
                <div class="table-header">
                    <div class="table-cell select">
                        <input type="checkbox" id="selectAllFiles">
                    </div>
                    <div class="table-cell name sortable" data-sort="name">
                        <span>名称</span>
                        <i class="mdi mdi-chevron-${this.sortBy === 'name' ? (this.sortOrder === 'asc' ? 'up' : 'down') : 'up'}"></i>
                    </div>
                    <div class="table-cell size sortable" data-sort="size">
                        <span>大小</span>
                        <i class="mdi mdi-chevron-${this.sortBy === 'size' ? (this.sortOrder === 'asc' ? 'up' : 'down') : 'up'}"></i>
                    </div>
                    <div class="table-cell modified sortable" data-sort="modified">
                        <span>修改时间</span>
                        <i class="mdi mdi-chevron-${this.sortBy === 'modified' ? (this.sortOrder === 'asc' ? 'up' : 'down') : 'up'}"></i>
                    </div>
                    <div class="table-cell actions">操作</div>
                </div>
                
                <div class="table-body">
                    ${files.map(file => this.renderFileRow(file)).join('')}
                </div>
            </div>
        `;
    }
    
    // 渲染文件行
    renderFileRow(file) {
        const fileIcon = this.getFileIcon(file);
        const fileSize = file.is_dir ? '-' : this.formatFileSize(file.size);
        const modifiedTime = new Date(file.modified_at).toLocaleString();
        
        return `
            <div class="table-row" data-file-path="${file.path}">
                <div class="table-cell select">
                    <input type="checkbox" class="file-checkbox" data-file-path="${file.path}">
                </div>
                <div class="table-cell name" data-action="open">
                    <div class="file-info">
                        <i class="mdi ${fileIcon}"></i>
                        <span class="file-name">${file.name}</span>
                    </div>
                </div>
                <div class="table-cell size">
                    <span>${fileSize}</span>
                </div>
                <div class="table-cell modified">
                    <span>${modifiedTime}</span>
                </div>
                <div class="table-cell actions">
                    <div class="action-buttons">
                        ${file.is_dir ? '' : `
                            <button class="btn-icon file-action-btn" data-action="download" data-file-path="${file.path}" title="下载">
                                <i class="mdi mdi-download"></i>
                            </button>
                            <button class="btn-icon file-action-btn" data-action="edit" data-file-path="${file.path}" title="编辑">
                                <i class="mdi mdi-pencil"></i>
                            </button>
                        `}
                        <div class="dropdown">
                            <button class="btn-icon dropdown-toggle" data-file-path="${file.path}">
                                <i class="mdi mdi-dots-vertical"></i>
                            </button>
                            <div class="dropdown-menu">
                                <a href="#" class="dropdown-item" data-action="rename" data-file-path="${file.path}">
                                    <i class="mdi mdi-rename-box"></i>
                                    <span>重命名</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="copy" data-file-path="${file.path}">
                                    <i class="mdi mdi-content-copy"></i>
                                    <span>复制</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="cut" data-file-path="${file.path}">
                                    <i class="mdi mdi-content-cut"></i>
                                    <span>剪切</span>
                                </a>
                                ${file.is_dir ? '' : `
                                    <a href="#" class="dropdown-item" data-action="properties" data-file-path="${file.path}">
                                        <i class="mdi mdi-information"></i>
                                        <span>属性</span>
                                    </a>
                                `}
                                <div class="dropdown-divider"></div>
                                <a href="#" class="dropdown-item danger" data-action="delete" data-file-path="${file.path}">
                                    <i class="mdi mdi-delete"></i>
                                    <span>删除</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染文件卡片
    renderFileCard(file) {
        const fileIcon = this.getFileIcon(file);
        const fileSize = file.is_dir ? '' : this.formatFileSize(file.size);
        
        return `
            <div class="file-card" data-file-path="${file.path}">
                <div class="file-card-header">
                    <input type="checkbox" class="file-checkbox" data-file-path="${file.path}">
                </div>
                
                <div class="file-card-body" data-action="open">
                    <div class="file-icon">
                        <i class="mdi ${fileIcon}"></i>
                    </div>
                    <div class="file-name">${file.name}</div>
                    ${fileSize ? `<div class="file-size">${fileSize}</div>` : ''}
                </div>
                
                <div class="file-card-footer">
                    <div class="file-actions">
                        ${file.is_dir ? '' : `
                            <button class="btn-icon file-action-btn" data-action="download" data-file-path="${file.path}" title="下载">
                                <i class="mdi mdi-download"></i>
                            </button>
                        `}
                        <div class="dropdown">
                            <button class="btn-icon dropdown-toggle" data-file-path="${file.path}">
                                <i class="mdi mdi-dots-vertical"></i>
                            </button>
                            <div class="dropdown-menu">
                                <a href="#" class="dropdown-item" data-action="rename" data-file-path="${file.path}">
                                    <i class="mdi mdi-rename-box"></i>
                                    <span>重命名</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="copy" data-file-path="${file.path}">
                                    <i class="mdi mdi-content-copy"></i>
                                    <span>复制</span>
                                </a>
                                <a href="#" class="dropdown-item danger" data-action="delete" data-file-path="${file.path}">
                                    <i class="mdi mdi-delete"></i>
                                    <span>删除</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 绑定文件事件
    bindFileEvents() {
        // 复选框事件
        const checkboxes = document.querySelectorAll('.file-checkbox');
        checkboxes.forEach(checkbox => {
            checkbox.addEventListener('change', (e) => {
                const filePath = e.target.getAttribute('data-file-path');
                if (e.target.checked) {
                    this.selectedFiles.add(filePath);
                } else {
                    this.selectedFiles.delete(filePath);
                }
                this.updateBatchActions();
            });
        });
        
        // 全选复选框
        const selectAllCheckbox = document.getElementById('selectAllFiles');
        if (selectAllCheckbox) {
            selectAllCheckbox.addEventListener('change', (e) => {
                const isChecked = e.target.checked;
                checkboxes.forEach(checkbox => {
                    checkbox.checked = isChecked;
                    const filePath = checkbox.getAttribute('data-file-path');
                    if (isChecked) {
                        this.selectedFiles.add(filePath);
                    } else {
                        this.selectedFiles.delete(filePath);
                    }
                });
                this.updateBatchActions();
            });
        }
        
        // 排序事件
        const sortableHeaders = document.querySelectorAll('.sortable');
        sortableHeaders.forEach(header => {
            header.addEventListener('click', (e) => {
                const sortBy = e.currentTarget.getAttribute('data-sort');
                if (this.sortBy === sortBy) {
                    this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
                } else {
                    this.sortBy = sortBy;
                    this.sortOrder = 'asc';
                }
                this.filterAndRenderFiles();
            });
        });
        
        // 文件打开事件
        const openElements = document.querySelectorAll('[data-action="open"]');
        openElements.forEach(element => {
            element.addEventListener('dblclick', (e) => {
                const filePath = e.currentTarget.closest('[data-file-path]').getAttribute('data-file-path');
                this.openFile(filePath);
            });
        });
        
        // 操作按钮事件
        const actionBtns = document.querySelectorAll('.file-action-btn');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.getAttribute('data-action');
                const filePath = e.currentTarget.getAttribute('data-file-path');
                this.handleFileAction(action, filePath);
            });
        });
        
        // 下拉菜单项事件
        const dropdownItems = document.querySelectorAll('.dropdown-item');
        dropdownItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const action = e.currentTarget.getAttribute('data-action');
                const filePath = e.currentTarget.getAttribute('data-file-path');
                if (filePath) {
                    this.handleFileAction(action, filePath);
                }
            });
        });
    }
    
    // 更新批量操作显示
    updateBatchActions() {
        const batchActions = document.getElementById('batchActions');
        const selectedCount = document.getElementById('selectedCount');
        const pasteBtn = document.getElementById('pasteBtn');
        
        if (this.selectedFiles.size > 0) {
            batchActions.style.display = 'flex';
            selectedCount.textContent = `已选择 ${this.selectedFiles.size} 个文件`;
        } else {
            batchActions.style.display = 'none';
        }
        
        // 显示/隐藏粘贴按钮
        if (this.clipboard && this.clipboard.files.length > 0) {
            pasteBtn.style.display = 'inline-flex';
        } else {
            pasteBtn.style.display = 'none';
        }
    }
    
    // 获取文件图标
    getFileIcon(file) {
        if (file.is_dir) {
            return 'mdi-folder';
        }
        
        const ext = file.name.split('.').pop().toLowerCase();
        const iconMap = {
            // 图片
            'jpg': 'mdi-file-image',
            'jpeg': 'mdi-file-image',
            'png': 'mdi-file-image',
            'gif': 'mdi-file-image',
            'bmp': 'mdi-file-image',
            'svg': 'mdi-file-image',
            
            // 文档
            'txt': 'mdi-file-document',
            'md': 'mdi-file-document',
            'pdf': 'mdi-file-pdf-box',
            'doc': 'mdi-file-word',
            'docx': 'mdi-file-word',
            'xls': 'mdi-file-excel',
            'xlsx': 'mdi-file-excel',
            'ppt': 'mdi-file-powerpoint',
            'pptx': 'mdi-file-powerpoint',
            
            // 代码
            'js': 'mdi-language-javascript',
            'html': 'mdi-language-html5',
            'css': 'mdi-language-css3',
            'java': 'mdi-language-java',
            'py': 'mdi-language-python',
            'php': 'mdi-language-php',
            'json': 'mdi-code-json',
            'xml': 'mdi-file-xml',
            'yml': 'mdi-file-code',
            'yaml': 'mdi-file-code',
            
            // 压缩包
            'zip': 'mdi-folder-zip',
            'rar': 'mdi-folder-zip',
            '7z': 'mdi-folder-zip',
            'tar': 'mdi-folder-zip',
            'gz': 'mdi-folder-zip',
            
            // 音视频
            'mp3': 'mdi-file-music',
            'wav': 'mdi-file-music',
            'flac': 'mdi-file-music',
            'mp4': 'mdi-file-video',
            'avi': 'mdi-file-video',
            'mkv': 'mdi-file-video',
            'mov': 'mdi-file-video',
            
            // 配置文件
            'properties': 'mdi-cog',
            'conf': 'mdi-cog',
            'cfg': 'mdi-cog',
            'ini': 'mdi-cog',
            
            // Minecraft相关
            'jar': 'mdi-minecraft',
            'mcmeta': 'mdi-minecraft',
            'mcfunction': 'mdi-minecraft'
        };
        
        return iconMap[ext] || 'mdi-file';
    }
    
    // 格式化文件大小
    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    // 导航到路径
    navigateToPath(path) {
        this.currentPath = path;
        this.selectedFiles.clear();
        this.loadFiles();
    }
    
    // 打开文件
    openFile(filePath) {
        const file = this.files.find(f => f.path === filePath);
        if (!file) return;
        
        if (file.is_dir) {
            this.navigateToPath(filePath);
        } else {
            // 根据文件类型决定打开方式
            const ext = file.name.split('.').pop().toLowerCase();
            const editableExts = ['txt', 'md', 'json', 'xml', 'yml', 'yaml', 'properties', 'conf', 'cfg', 'ini', 'js', 'html', 'css', 'java', 'py', 'php'];
            
            if (editableExts.includes(ext)) {
                this.editFile(filePath);
            } else {
                this.downloadFile(filePath);
            }
        }
    }
    
    // 处理文件操作
    handleFileAction(action, filePath) {
        switch (action) {
            case 'download':
                this.downloadFile(filePath);
                break;
            case 'edit':
                this.editFile(filePath);
                break;
            case 'rename':
                this.showRenameDialog(filePath);
                break;
            case 'copy':
                this.copyFile(filePath);
                break;
            case 'cut':
                this.cutFile(filePath);
                break;
            case 'delete':
                this.showDeleteDialog(filePath);
                break;
            case 'properties':
                this.showFileProperties(filePath);
                break;
        }
    }
    
    // 处理创建操作
    handleCreateAction(action) {
        switch (action) {
            case 'create-folder':
                this.showCreateFolderDialog();
                break;
            case 'create-file':
                this.showCreateFileDialog();
                break;
            case 'upload':
                this.showUploadDialog();
                break;
        }
    }
    
    // 显示错误信息
    showError(message) {
        const filesList = document.getElementById('filesList');
        if (filesList) {
            filesList.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h3>加载失败</h3>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getFilesPageManager().loadFiles()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }

    // 其他操作方法
    goUp() {
        const parentPath = this.currentPath.split('/').slice(0, -1).join('/') || '/';
        this.navigateToPath(parentPath);
    }

    deleteSelectedFiles() {
        if (this.selectedFiles.size === 0) return;

        if (confirm(`确定要删除选中的 ${this.selectedFiles.size} 个文件吗？`)) {
            console.log('Delete selected files:', Array.from(this.selectedFiles));
            // TODO: 实现删除逻辑
        }
    }

    downloadFile(filePath) {
        window.open(`/api/files/download?path=${encodeURIComponent(filePath)}`, '_blank');
    }

    editFile(filePath) {
        console.log('Edit file:', filePath);
        // TODO: 实现文件编辑功能
    }

    showDeleteDialog(filePath) {
        const file = this.files.find(f => f.path === filePath);
        if (!file) return;

        if (confirm(`确定要删除 "${file.name}" 吗？`)) {
            console.log('Delete file:', filePath);
            // TODO: 实现删除逻辑
        }
    }

    showCreateFolderDialog() {
        const folderName = prompt('请输入文件夹名称:');
        if (folderName) {
            console.log('Create folder:', folderName);
            // TODO: 实现创建文件夹逻辑
        }
    }

    showUploadDialog() {
        const input = document.createElement('input');
        input.type = 'file';
        input.multiple = true;
        input.onchange = (e) => {
            const files = Array.from(e.target.files);
            if (files.length > 0) {
                this.uploadFiles(files);
            }
        };
        input.click();
    }

    uploadFiles(files) {
        console.log('Upload files:', files);
        // TODO: 实现文件上传逻辑
    }
}

// 全局文件页面管理器实例
let filesPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    filesPageManager = new FilesPageManager();
});

// 导出到全局作用域
window.FilesPageManager = FilesPageManager;
window.getFilesPageManager = () => filesPageManager;
