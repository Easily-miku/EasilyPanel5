// 下载中心页面管理器
class DownloadPageManager {
    constructor() {
        this.currentCategory = 'server'; // server, plugin, mod, tool
        this.searchQuery = '';
        this.sortBy = 'name'; // name, version, downloads, date
        this.sortOrder = 'asc';
        this.downloads = new Map();
        this.categories = new Map();
        this.isLoading = false;
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadData();
    }
    
    // 渲染页面
    renderPage() {
        const downloadPage = document.getElementById('download-page');
        if (!downloadPage) return;

        downloadPage.innerHTML = `
            <div class="page-header">
                <h2>下载中心</h2>
                <p>Minecraft 服务端核心、插件、模组等资源下载</p>
            </div>

            <div class="download-container">
                ${this.renderCategoryTabs()}
                ${this.renderToolbar()}
                ${this.renderDownloadGrid()}
            </div>
        `;
    }
    
    // 渲染分类标签页
    renderCategoryTabs() {
        const categories = [
            { id: 'server', name: '服务端核心', icon: 'mdi-server', description: 'Paper, Spigot, Bukkit 等服务端核心' },
            { id: 'plugin', name: '插件', icon: 'mdi-puzzle', description: '各种功能插件和扩展' },
            { id: 'mod', name: '模组', icon: 'mdi-cube', description: 'Forge, Fabric 模组' },
            { id: 'tool', name: '工具', icon: 'mdi-tools', description: '服务器管理和开发工具' }
        ];

        return `
            <div class="category-tabs">
                ${categories.map(category => `
                    <div class="category-tab ${this.currentCategory === category.id ? 'active' : ''}" data-category="${category.id}">
                        <div class="tab-icon">
                            <i class="mdi ${category.icon}"></i>
                        </div>
                        <div class="tab-content">
                            <h3>${category.name}</h3>
                            <p>${category.description}</p>
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    }
    
    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="download-toolbar">
                <div class="toolbar-left">
                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="downloadSearch" placeholder="搜索资源..." value="${this.searchQuery}">
                    </div>
                    
                    <div class="version-filter">
                        <select id="versionFilter" class="filter-select">
                            <option value="all">全部版本</option>
                            <option value="1.20">1.20.x</option>
                            <option value="1.19">1.19.x</option>
                            <option value="1.18">1.18.x</option>
                            <option value="1.17">1.17.x</option>
                            <option value="1.16">1.16.x</option>
                            <option value="1.12">1.12.x</option>
                            <option value="1.8">1.8.x</option>
                        </select>
                    </div>
                </div>
                
                <div class="toolbar-right">
                    <div class="sort-controls">
                        <select id="sortBy" class="filter-select">
                            <option value="name">按名称</option>
                            <option value="version">按版本</option>
                            <option value="downloads">按下载量</option>
                            <option value="date">按更新时间</option>
                        </select>
                        
                        <button class="btn-icon sort-order" id="sortOrderBtn" title="排序方向">
                            <i class="mdi ${this.sortOrder === 'asc' ? 'mdi-sort-ascending' : 'mdi-sort-descending'}"></i>
                        </button>
                    </div>
                    
                    <button class="btn secondary" id="refreshBtn">
                        <i class="mdi mdi-refresh"></i>
                        <span>刷新</span>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 渲染下载网格
    renderDownloadGrid() {
        return `
            <div class="download-grid-container">
                <div class="download-grid" id="downloadGrid">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载资源列表...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 分类标签切换
        const categoryTabs = document.querySelectorAll('.category-tab');
        categoryTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                const category = e.currentTarget.getAttribute('data-category');
                this.switchCategory(category);
            });
        });
        
        // 搜索框
        const searchInput = document.getElementById('downloadSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderDownloads();
            });
        }
        
        // 版本过滤
        const versionFilter = document.getElementById('versionFilter');
        if (versionFilter) {
            versionFilter.addEventListener('change', () => {
                this.filterAndRenderDownloads();
            });
        }
        
        // 排序控制
        const sortBy = document.getElementById('sortBy');
        const sortOrderBtn = document.getElementById('sortOrderBtn');
        
        if (sortBy) {
            sortBy.addEventListener('change', (e) => {
                this.sortBy = e.target.value;
                this.filterAndRenderDownloads();
            });
        }
        
        if (sortOrderBtn) {
            sortOrderBtn.addEventListener('click', () => {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
                const icon = sortOrderBtn.querySelector('i');
                icon.className = `mdi ${this.sortOrder === 'asc' ? 'mdi-sort-ascending' : 'mdi-sort-descending'}`;
                this.filterAndRenderDownloads();
            });
        }
        
        // 刷新按钮
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadData();
            });
        }
    }
    
    // 切换分类
    switchCategory(category) {
        this.currentCategory = category;
        
        // 更新标签状态
        document.querySelectorAll('.category-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-category="${category}"]`).classList.add('active');
        
        // 重新加载数据
        this.loadData();
    }
    
    // 加载数据
    async loadData() {
        if (this.isLoading) return;
        
        this.isLoading = true;
        
        try {
            const response = await fetch(`/api/cores/list?category=${this.currentCategory}`);
            const result = await response.json();
            
            if (response.ok && result.success) {
                this.downloads.clear();
                const downloads = result.data || [];
                downloads.forEach(download => {
                    this.downloads.set(download.id, download);
                });
                
                this.filterAndRenderDownloads();
            } else {
                this.showError(result.message || '加载失败');
            }
        } catch (error) {
            console.error('Failed to load downloads:', error);
            this.showError('网络错误，请重试');
        } finally {
            this.isLoading = false;
        }
    }
    
    // 过滤和渲染下载项
    filterAndRenderDownloads() {
        const downloadGrid = document.getElementById('downloadGrid');
        if (!downloadGrid) return;
        
        let filteredDownloads = Array.from(this.downloads.values());
        
        // 应用搜索过滤
        if (this.searchQuery) {
            const query = this.searchQuery.toLowerCase();
            filteredDownloads = filteredDownloads.filter(download => 
                download.name.toLowerCase().includes(query) ||
                download.description.toLowerCase().includes(query)
            );
        }
        
        // 应用版本过滤
        const versionFilter = document.getElementById('versionFilter');
        if (versionFilter && versionFilter.value !== 'all') {
            const version = versionFilter.value;
            filteredDownloads = filteredDownloads.filter(download => 
                download.versions && download.versions.some(v => v.startsWith(version))
            );
        }
        
        // 排序
        filteredDownloads.sort((a, b) => {
            let result = 0;
            switch (this.sortBy) {
                case 'name':
                    result = a.name.localeCompare(b.name);
                    break;
                case 'version':
                    result = (a.latest_version || '').localeCompare(b.latest_version || '');
                    break;
                case 'downloads':
                    result = (a.downloads || 0) - (b.downloads || 0);
                    break;
                case 'date':
                    result = new Date(a.updated_at || 0) - new Date(b.updated_at || 0);
                    break;
            }
            
            return this.sortOrder === 'desc' ? -result : result;
        });
        
        // 渲染
        if (filteredDownloads.length === 0) {
            downloadGrid.innerHTML = this.renderEmptyState();
        } else {
            downloadGrid.innerHTML = filteredDownloads.map(download => 
                this.renderDownloadCard(download)
            ).join('');
        }
        
        // 重新绑定事件
        this.bindDownloadEvents();
    }
    
    // 渲染下载卡片
    renderDownloadCard(download) {
        return `
            <div class="download-card" data-download-id="${download.id}">
                <div class="card-header">
                    <div class="download-icon">
                        <i class="mdi ${this.getDownloadIcon(download.type)}"></i>
                    </div>
                    <div class="download-info">
                        <h3 class="download-name">${download.name}</h3>
                        <p class="download-description">${download.description || '暂无描述'}</p>
                    </div>
                </div>
                
                <div class="card-content">
                    <div class="download-stats">
                        <div class="stat-item">
                            <span class="stat-label">最新版本</span>
                            <span class="stat-value">${download.latest_version || 'N/A'}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">下载次数</span>
                            <span class="stat-value">${this.formatNumber(download.downloads || 0)}</span>
                        </div>
                        <div class="stat-item">
                            <span class="stat-label">更新时间</span>
                            <span class="stat-value">${this.formatDate(download.updated_at)}</span>
                        </div>
                    </div>
                    
                    <div class="download-versions">
                        <label>支持版本:</label>
                        <div class="version-tags">
                            ${(download.versions || []).slice(0, 3).map(version => 
                                `<span class="version-tag">${version}</span>`
                            ).join('')}
                            ${(download.versions || []).length > 3 ? 
                                `<span class="version-more">+${(download.versions || []).length - 3}</span>` : ''
                            }
                        </div>
                    </div>
                </div>
                
                <div class="card-actions">
                    <button class="btn secondary" data-action="info" data-download-id="${download.id}">
                        <i class="mdi mdi-information"></i>
                        <span>详情</span>
                    </button>
                    <button class="btn primary" data-action="download" data-download-id="${download.id}">
                        <i class="mdi mdi-download"></i>
                        <span>下载</span>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-download-off"></i>
                <h3>暂无资源</h3>
                <p>当前分类下没有找到匹配的资源，请尝试其他搜索条件。</p>
            </div>
        `;
    }
    
    // 绑定下载事件
    bindDownloadEvents() {
        const actionBtns = document.querySelectorAll('[data-action]');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.getAttribute('data-action');
                const downloadId = e.currentTarget.getAttribute('data-download-id');
                this.handleDownloadAction(action, downloadId);
            });
        });
    }
    
    // 处理下载操作
    handleDownloadAction(action, downloadId) {
        const download = this.downloads.get(downloadId);
        if (!download) return;
        
        switch (action) {
            case 'info':
                this.showDownloadInfo(download);
                break;
            case 'download':
                this.startDownload(download);
                break;
        }
    }
    
    // 显示下载信息
    showDownloadInfo(download) {
        const uiManager = window.getUIManager();
        if (!uiManager) return;
        
        const content = `
            <div class="download-info-modal">
                <div class="info-header">
                    <div class="download-icon">
                        <i class="mdi ${this.getDownloadIcon(download.type)}"></i>
                    </div>
                    <div class="download-details">
                        <h3>${download.name}</h3>
                        <p>${download.description || '暂无描述'}</p>
                    </div>
                </div>
                
                <div class="info-content">
                    <div class="info-section">
                        <h4>基本信息</h4>
                        <div class="info-grid">
                            <div class="info-item">
                                <span class="label">类型:</span>
                                <span class="value">${download.type || 'Unknown'}</span>
                            </div>
                            <div class="info-item">
                                <span class="label">最新版本:</span>
                                <span class="value">${download.latest_version || 'N/A'}</span>
                            </div>
                            <div class="info-item">
                                <span class="label">文件大小:</span>
                                <span class="value">${this.formatFileSize(download.file_size || 0)}</span>
                            </div>
                            <div class="info-item">
                                <span class="label">下载次数:</span>
                                <span class="value">${this.formatNumber(download.downloads || 0)}</span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="info-section">
                        <h4>支持版本</h4>
                        <div class="version-list">
                            ${(download.versions || []).map(version => 
                                `<span class="version-tag">${version}</span>`
                            ).join('')}
                        </div>
                    </div>
                    
                    ${download.changelog ? `
                        <div class="info-section">
                            <h4>更新日志</h4>
                            <div class="changelog">
                                ${download.changelog}
                            </div>
                        </div>
                    ` : ''}
                </div>
            </div>
        `;
        
        uiManager.showModal(`${download.name} - 详细信息`, content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>关闭</span>
                </button>
                <button class="btn primary" onclick="window.getDownloadPageManager().startDownload('${download.id}'); window.getUIManager().closeModal();">
                    <i class="mdi mdi-download"></i>
                    <span>下载</span>
                </button>
            `
        });
    }
    
    // 开始下载
    startDownload(downloadId) {
        const download = typeof downloadId === 'string' ? this.downloads.get(downloadId) : downloadId;
        if (!download) return;
        
        const uiManager = window.getUIManager();
        
        // 创建下载链接并触发下载
        const downloadUrl = `/api/cores/download?id=${download.id}&version=${download.latest_version}`;
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = `${download.name}-${download.latest_version}.jar`;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        uiManager?.showNotification('下载开始', `正在下载 ${download.name}`, 'success');
    }
    
    // 显示错误
    showError(message) {
        const downloadGrid = document.getElementById('downloadGrid');
        if (downloadGrid) {
            downloadGrid.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h3>加载失败</h3>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getDownloadPageManager().loadData()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 工具方法
    getDownloadIcon(type) {
        const icons = {
            'server': 'mdi-server',
            'plugin': 'mdi-puzzle',
            'mod': 'mdi-cube',
            'tool': 'mdi-tools'
        };
        return icons[type] || 'mdi-download';
    }
    
    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }
    
    formatDate(dateString) {
        if (!dateString) return 'N/A';
        const date = new Date(dateString);
        return date.toLocaleDateString();
    }
    
    formatFileSize(bytes) {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
}

// 全局下载页面管理器实例
let downloadPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    downloadPageManager = new DownloadPageManager();
});

// 导出到全局作用域
window.DownloadPageManager = DownloadPageManager;
window.getDownloadPageManager = () => downloadPageManager;
