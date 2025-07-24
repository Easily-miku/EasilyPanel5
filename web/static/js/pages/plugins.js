// 插件管理页面管理器
class PluginsPageManager {
    constructor() {
        this.currentCategory = 'all';
        this.currentPage = 1;
        this.searchQuery = '';
        this.plugins = new Map();
        this.isLoading = false;
        this.categories = [
            { id: 'all', name: '全部插件', icon: 'mdi-puzzle' },
            { id: 'admin', name: '管理工具', icon: 'mdi-shield-account' },
            { id: 'economy', name: '经济系统', icon: 'mdi-currency-usd' },
            { id: 'protection', name: '保护插件', icon: 'mdi-shield' },
            { id: 'fun', name: '娱乐插件', icon: 'mdi-gamepad-variant' },
            { id: 'utility', name: '实用工具', icon: 'mdi-tools' },
            { id: 'world', name: '世界管理', icon: 'mdi-earth' }
        ];
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadPlugins();
    }
    
    // 渲染页面
    renderPage() {
        const pluginsPage = document.getElementById('plugins-page');
        if (!pluginsPage) return;

        pluginsPage.innerHTML = `
            <div class="page-header">
                <h2>插件管理</h2>
                <p>浏览、搜索和下载Minecraft服务器插件</p>
            </div>

            <div class="plugins-container">
                ${this.renderToolbar()}
                ${this.renderCategoryTabs()}
                ${this.renderPluginGrid()}
                ${this.renderPagination()}
            </div>
        `;
    }
    
    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="plugins-toolbar">
                <div class="search-box">
                    <i class="mdi mdi-magnify"></i>
                    <input type="text" id="pluginSearch" placeholder="搜索插件..." value="${this.searchQuery}">
                    <button class="search-btn" id="searchBtn">
                        <i class="mdi mdi-magnify"></i>
                    </button>
                </div>
                <div class="toolbar-actions">
                    <button class="btn secondary" id="refreshBtn">
                        <i class="mdi mdi-refresh"></i>
                        <span>刷新</span>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 渲染分类标签
    renderCategoryTabs() {
        return `
            <div class="category-tabs">
                ${this.categories.map(category => `
                    <button class="category-tab ${this.currentCategory === category.id ? 'active' : ''}" 
                            data-category="${category.id}">
                        <i class="${category.icon}"></i>
                        <span>${category.name}</span>
                    </button>
                `).join('')}
            </div>
        `;
    }
    
    // 渲染插件网格
    renderPluginGrid() {
        return `
            <div class="plugins-grid" id="pluginsGrid">
                ${this.isLoading ? this.renderLoadingState() : this.renderPluginCards()}
            </div>
        `;
    }
    
    // 渲染插件卡片
    renderPluginCards() {
        const plugins = Array.from(this.plugins.values());
        
        if (plugins.length === 0) {
            return this.renderEmptyState();
        }
        
        return plugins.map(plugin => `
            <div class="plugin-card" data-plugin-id="${plugin.id}">
                <div class="plugin-header">
                    <div class="plugin-icon">
                        ${plugin.icon ? `<img src="${plugin.icon}" alt="${plugin.name}">` : 
                          '<i class="mdi mdi-puzzle"></i>'}
                    </div>
                    <div class="plugin-info">
                        <h3 class="plugin-name">${plugin.name}</h3>
                        <p class="plugin-author">by ${plugin.author}</p>
                    </div>
                    <div class="plugin-rating">
                        <i class="mdi mdi-star"></i>
                        <span>${plugin.rating ? plugin.rating.toFixed(1) : 'N/A'}</span>
                    </div>
                </div>
                
                <div class="plugin-description">
                    <p>${plugin.description || '暂无描述'}</p>
                </div>
                
                <div class="plugin-meta">
                    <div class="plugin-category">
                        <i class="mdi mdi-tag"></i>
                        <span>${plugin.category}</span>
                    </div>
                    <div class="plugin-downloads">
                        <i class="mdi mdi-download"></i>
                        <span>${this.formatNumber(plugin.downloads)}</span>
                    </div>
                    <div class="plugin-version">
                        <i class="mdi mdi-package-variant"></i>
                        <span>${plugin.latest_version || 'Unknown'}</span>
                    </div>
                </div>
                
                <div class="plugin-actions">
                    <button class="btn secondary" onclick="window.getPluginsPageManager().showPluginInfo('${plugin.id}')">
                        <i class="mdi mdi-information"></i>
                        <span>详情</span>
                    </button>
                    <button class="btn primary" onclick="window.getPluginsPageManager().downloadPlugin('${plugin.id}')">
                        <i class="mdi mdi-download"></i>
                        <span>下载</span>
                    </button>
                </div>
            </div>
        `).join('');
    }
    
    // 渲染分页
    renderPagination() {
        return `
            <div class="pagination" id="pluginsPagination">
                <!-- 分页控件将在这里动态生成 -->
            </div>
        `;
    }
    
    // 渲染加载状态
    renderLoadingState() {
        return `
            <div class="loading-state">
                <div class="loading-spinner"></div>
                <p>正在加载插件...</p>
            </div>
        `;
    }
    
    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-puzzle-outline"></i>
                <h3>暂无插件</h3>
                <p>当前分类下没有找到匹配的插件，请尝试其他搜索条件。</p>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 搜索功能
        const searchInput = document.getElementById('pluginSearch');
        const searchBtn = document.getElementById('searchBtn');
        
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.debounceSearch();
            });
            
            searchInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    this.performSearch();
                }
            });
        }
        
        if (searchBtn) {
            searchBtn.addEventListener('click', () => {
                this.performSearch();
            });
        }
        
        // 刷新按钮
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadPlugins();
            });
        }
        
        // 分类标签
        const categoryTabs = document.querySelectorAll('.category-tab');
        categoryTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                const category = e.currentTarget.getAttribute('data-category');
                this.switchCategory(category);
            });
        });
    }
    
    // 防抖搜索
    debounceSearch() {
        clearTimeout(this.searchTimeout);
        this.searchTimeout = setTimeout(() => {
            this.performSearch();
        }, 500);
    }
    
    // 执行搜索
    async performSearch() {
        if (this.searchQuery.trim()) {
            await this.searchPlugins(this.searchQuery);
        } else {
            await this.loadPlugins();
        }
    }
    
    // 切换分类
    async switchCategory(category) {
        if (this.currentCategory === category) return;
        
        this.currentCategory = category;
        this.currentPage = 1;
        
        // 更新分类标签状态
        document.querySelectorAll('.category-tab').forEach(tab => {
            tab.classList.toggle('active', tab.getAttribute('data-category') === category);
        });
        
        await this.loadPlugins();
    }
    
    // 加载插件列表
    async loadPlugins() {
        if (this.isLoading) return;
        
        this.isLoading = true;
        this.updatePluginGrid();
        
        try {
            const url = `/api/plugins/list?category=${this.currentCategory}&page=${this.currentPage}`;
            const response = await fetch(url);
            const result = await response.json();
            
            if (response.ok && result.success) {
                this.plugins.clear();
                const plugins = result.data.plugins || [];
                plugins.forEach(plugin => {
                    this.plugins.set(plugin.id, plugin);
                });
                
                this.updatePluginGrid();
                this.updatePagination(result.data);
            } else {
                this.showError(result.message || '加载插件失败');
            }
        } catch (error) {
            console.error('Failed to load plugins:', error);
            this.showError('网络错误，请重试');
        } finally {
            this.isLoading = false;
        }
    }
    
    // 搜索插件
    async searchPlugins(query) {
        if (this.isLoading) return;
        
        this.isLoading = true;
        this.updatePluginGrid();
        
        try {
            const url = `/api/plugins/search?q=${encodeURIComponent(query)}&category=${this.currentCategory}`;
            const response = await fetch(url);
            const result = await response.json();
            
            if (response.ok && result.success) {
                this.plugins.clear();
                const plugins = result.data.plugins || [];
                plugins.forEach(plugin => {
                    this.plugins.set(plugin.id, plugin);
                });
                
                this.updatePluginGrid();
            } else {
                this.showError(result.message || '搜索失败');
            }
        } catch (error) {
            console.error('Failed to search plugins:', error);
            this.showError('搜索失败，请重试');
        } finally {
            this.isLoading = false;
        }
    }
    
    // 更新插件网格
    updatePluginGrid() {
        const pluginsGrid = document.getElementById('pluginsGrid');
        if (pluginsGrid) {
            pluginsGrid.innerHTML = this.isLoading ? this.renderLoadingState() : this.renderPluginCards();
        }
    }
    
    // 更新分页
    updatePagination(data) {
        const pagination = document.getElementById('pluginsPagination');
        if (!pagination || !data.total_pages || data.total_pages <= 1) {
            pagination.innerHTML = '';
            return;
        }
        
        // 这里可以添加分页控件的实现
        pagination.innerHTML = `
            <div class="pagination-info">
                第 ${data.page} 页，共 ${data.total_pages} 页 (${data.total} 个插件)
            </div>
        `;
    }
    
    // 显示插件详情
    showPluginInfo(pluginId) {
        const plugin = this.plugins.get(pluginId);
        if (!plugin) return;
        
        const uiManager = window.getUIManager();
        if (!uiManager) return;
        
        const content = `
            <div class="plugin-detail">
                <div class="plugin-detail-header">
                    <div class="plugin-icon">
                        ${plugin.icon ? `<img src="${plugin.icon}" alt="${plugin.name}">` : 
                          '<i class="mdi mdi-puzzle"></i>'}
                    </div>
                    <div class="plugin-info">
                        <h3>${plugin.name}</h3>
                        <p>作者: ${plugin.author}</p>
                        <p>分类: ${plugin.category}</p>
                        <p>下载量: ${this.formatNumber(plugin.downloads)}</p>
                    </div>
                </div>
                
                <div class="plugin-description">
                    <h4>插件描述</h4>
                    <p>${plugin.description || '暂无描述'}</p>
                </div>
                
                ${plugin.homepage ? `
                    <div class="plugin-links">
                        <h4>相关链接</h4>
                        <a href="${plugin.homepage}" target="_blank" class="btn secondary">
                            <i class="mdi mdi-open-in-new"></i>
                            <span>官方主页</span>
                        </a>
                    </div>
                ` : ''}
            </div>
        `;
        
        uiManager.showModal(`${plugin.name} - 详细信息`, content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>关闭</span>
                </button>
                <button class="btn primary" onclick="window.getPluginsPageManager().downloadPlugin('${plugin.id}'); window.getUIManager().closeModal();">
                    <i class="mdi mdi-download"></i>
                    <span>下载</span>
                </button>
            `
        });
    }
    
    // 下载插件
    async downloadPlugin(pluginId) {
        const plugin = this.plugins.get(pluginId);
        if (!plugin) return;
        
        const uiManager = window.getUIManager();
        
        try {
            const response = await fetch('/api/plugins/download', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    plugin_id: pluginId,
                    version: plugin.latest_version
                })
            });
            
            const result = await response.json();
            
            if (response.ok && result.success) {
                // 打开下载链接
                window.open(result.data.download_url, '_blank');
                uiManager?.showNotification('下载开始', `正在下载 ${plugin.name}`, 'success');
            } else {
                uiManager?.showNotification('下载失败', result.message || '下载失败', 'error');
            }
        } catch (error) {
            console.error('Download failed:', error);
            uiManager?.showNotification('下载失败', '网络错误，请重试', 'error');
        }
    }
    
    // 显示错误信息
    showError(message) {
        const uiManager = window.getUIManager();
        uiManager?.showNotification('错误', message, 'error');
    }
    
    // 格式化数字
    formatNumber(num) {
        if (!num) return '0';
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        }
        if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }
}

// 全局插件页面管理器实例
let pluginsPageManager = null;

// 获取插件页面管理器实例
function getPluginsPageManager() {
    if (!pluginsPageManager) {
        pluginsPageManager = new PluginsPageManager();
    }
    return pluginsPageManager;
}

// 导出到全局
window.getPluginsPageManager = getPluginsPageManager;
