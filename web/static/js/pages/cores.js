// 核心下载页面管理器
class CoresPageManager {
    constructor() {
        this.currentCategory = 'server'; // server, proxy
        this.searchQuery = '';
        this.sortBy = 'name'; // name, version, downloads, date
        this.sortOrder = 'asc';
        this.cores = new Map();
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
        const coresPage = document.getElementById('cores-page');
        if (!coresPage) return;

        coresPage.innerHTML = `
            <div class="page-header">
                <h2>核心下载</h2>
                <p>Minecraft 服务端核心下载中心</p>
            </div>

            <div class="cores-container">
                ${this.renderCategoryTabs()}
                ${this.renderToolbar()}
                ${this.renderCoresGrid()}
            </div>
        `;
    }
    
    // 渲染分类标签
    renderCategoryTabs() {
        const categories = [
            { id: 'server', name: '服务端核心', icon: 'mdi-server' },
            { id: 'proxy', name: '代理核心', icon: 'mdi-network' }
        ];
        
        return `
            <div class="category-tabs">
                ${categories.map(category => `
                    <button class="category-tab ${this.currentCategory === category.id ? 'active' : ''}" 
                            data-category="${category.id}">
                        <i class="${category.icon}"></i>
                        <span>${category.name}</span>
                    </button>
                `).join('')}
            </div>
        `;
    }
    
    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="cores-toolbar">
                <div class="search-box">
                    <i class="mdi mdi-magnify"></i>
                    <input type="text" id="coreSearch" placeholder="搜索核心..." value="${this.searchQuery}">
                </div>
                <div class="sort-controls">
                    <select id="sortBy" class="sort-select">
                        <option value="name" ${this.sortBy === 'name' ? 'selected' : ''}>按名称排序</option>
                        <option value="downloads" ${this.sortBy === 'downloads' ? 'selected' : ''}>按下载量排序</option>
                        <option value="date" ${this.sortBy === 'date' ? 'selected' : ''}>按更新时间排序</option>
                    </select>
                    <button class="sort-order-btn" id="sortOrderBtn" title="排序方向">
                        <i class="mdi ${this.sortOrder === 'asc' ? 'mdi-sort-ascending' : 'mdi-sort-descending'}"></i>
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
    
    // 渲染核心网格
    renderCoresGrid() {
        return `
            <div class="cores-grid" id="coresGrid">
                ${this.isLoading ? this.renderLoadingState() : this.renderCoreCards()}
            </div>
        `;
    }
    
    // 渲染核心卡片
    renderCoreCards() {
        const cores = this.getFilteredAndSortedCores();
        
        if (cores.length === 0) {
            return this.renderEmptyState();
        }
        
        return cores.map(core => `
            <div class="core-card" data-core-id="${core.id}">
                <div class="core-header">
                    <div class="core-icon">
                        <i class="mdi mdi-cube-outline"></i>
                    </div>
                    <div class="core-info">
                        <h3 class="core-name">${core.name}</h3>
                        <p class="core-type">${core.tag || 'Server'}</p>
                    </div>
                    ${core.recommend ? '<div class="recommend-badge">推荐</div>' : ''}
                </div>
                
                <div class="core-description">
                    <p>${core.description || '高性能的Minecraft服务端核心'}</p>
                </div>
                
                <div class="core-meta">
                    <div class="core-versions">
                        <i class="mdi mdi-package-variant"></i>
                        <span>${core.mc_versions ? core.mc_versions.length + ' 个版本' : '多个版本'}</span>
                    </div>
                    <div class="core-downloads">
                        <i class="mdi mdi-download"></i>
                        <span>${this.formatNumber(core.downloads || 0)}</span>
                    </div>
                </div>
                
                <div class="core-actions">
                    <button class="btn secondary" onclick="window.getCoresPageManager().showCoreInfo('${core.id}')">
                        <i class="mdi mdi-information"></i>
                        <span>详情</span>
                    </button>
                    <button class="btn primary" onclick="window.getCoresPageManager().downloadCore('${core.id}')">
                        <i class="mdi mdi-download"></i>
                        <span>下载</span>
                    </button>
                </div>
            </div>
        `).join('');
    }
    
    // 渲染加载状态
    renderLoadingState() {
        return `
            <div class="loading-state">
                <div class="loading-spinner"></div>
                <p>正在加载核心列表...</p>
            </div>
        `;
    }
    
    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-cube-off-outline"></i>
                <h3>暂无核心</h3>
                <p>当前分类下没有找到匹配的核心，请尝试其他搜索条件。</p>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 搜索功能
        const searchInput = document.getElementById('coreSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderCores();
            });
        }
        
        // 排序控件
        const sortBy = document.getElementById('sortBy');
        if (sortBy) {
            sortBy.addEventListener('change', (e) => {
                this.sortBy = e.target.value;
                this.filterAndRenderCores();
            });
        }
        
        const sortOrderBtn = document.getElementById('sortOrderBtn');
        if (sortOrderBtn) {
            sortOrderBtn.addEventListener('click', () => {
                this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
                this.updateSortOrderButton();
                this.filterAndRenderCores();
            });
        }
        
        // 刷新按钮
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => {
                this.loadData();
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
    
    // 切换分类
    async switchCategory(category) {
        if (this.currentCategory === category) return;
        
        this.currentCategory = category;
        
        // 更新分类标签状态
        document.querySelectorAll('.category-tab').forEach(tab => {
            tab.classList.toggle('active', tab.getAttribute('data-category') === category);
        });
        
        await this.loadData();
    }
    
    // 加载数据
    async loadData() {
        if (this.isLoading) return;
        
        this.isLoading = true;
        this.updateCoresGrid();
        
        try {
            const response = await fetch(`/api/cores/list?category=${this.currentCategory}`);
            const result = await response.json();
            
            if (response.ok && result.success) {
                this.cores.clear();
                const cores = result.data || [];
                cores.forEach(core => {
                    this.cores.set(core.id || core.name, core);
                });
                
                this.filterAndRenderCores();
            } else {
                this.showError(result.message || '加载失败');
            }
        } catch (error) {
            console.error('Failed to load cores:', error);
            this.showError('网络错误，请重试');
        } finally {
            this.isLoading = false;
        }
    }
    
    // 过滤和排序核心
    getFilteredAndSortedCores() {
        let cores = Array.from(this.cores.values());
        
        // 搜索过滤
        if (this.searchQuery) {
            const query = this.searchQuery.toLowerCase();
            cores = cores.filter(core => 
                core.name.toLowerCase().includes(query) ||
                (core.description && core.description.toLowerCase().includes(query))
            );
        }
        
        // 排序
        cores.sort((a, b) => {
            let aVal, bVal;
            
            switch (this.sortBy) {
                case 'downloads':
                    aVal = a.downloads || 0;
                    bVal = b.downloads || 0;
                    break;
                case 'date':
                    aVal = new Date(a.last_update || 0);
                    bVal = new Date(b.last_update || 0);
                    break;
                default: // name
                    aVal = a.name.toLowerCase();
                    bVal = b.name.toLowerCase();
            }
            
            if (this.sortOrder === 'desc') {
                return aVal < bVal ? 1 : -1;
            } else {
                return aVal > bVal ? 1 : -1;
            }
        });
        
        return cores;
    }
    
    // 过滤并渲染核心
    filterAndRenderCores() {
        this.updateCoresGrid();
    }
    
    // 更新核心网格
    updateCoresGrid() {
        const coresGrid = document.getElementById('coresGrid');
        if (coresGrid) {
            coresGrid.innerHTML = this.isLoading ? this.renderLoadingState() : this.renderCoreCards();
        }
    }
    
    // 更新排序按钮
    updateSortOrderButton() {
        const sortOrderBtn = document.getElementById('sortOrderBtn');
        if (sortOrderBtn) {
            const icon = sortOrderBtn.querySelector('i');
            icon.className = `mdi ${this.sortOrder === 'asc' ? 'mdi-sort-ascending' : 'mdi-sort-descending'}`;
        }
    }
    
    // 显示核心详情
    showCoreInfo(coreId) {
        const core = this.cores.get(coreId);
        if (!core) return;
        
        const uiManager = window.getUIManager();
        if (!uiManager) return;
        
        const content = `
            <div class="core-detail">
                <div class="core-detail-header">
                    <div class="core-icon">
                        <i class="mdi mdi-cube-outline"></i>
                    </div>
                    <div class="core-info">
                        <h3>${core.name}</h3>
                        <p>类型: ${core.tag || 'Server'}</p>
                        <p>下载量: ${this.formatNumber(core.downloads || 0)}</p>
                    </div>
                </div>
                
                <div class="core-description">
                    <h4>核心描述</h4>
                    <p>${core.description || '高性能的Minecraft服务端核心'}</p>
                </div>
                
                ${core.mc_versions ? `
                    <div class="core-versions">
                        <h4>支持版本</h4>
                        <div class="version-list">
                            ${core.mc_versions.slice(0, 10).map(version => 
                                `<span class="version-tag">${version}</span>`
                            ).join('')}
                            ${core.mc_versions.length > 10 ? `<span class="version-more">+${core.mc_versions.length - 10} 更多</span>` : ''}
                        </div>
                    </div>
                ` : ''}
                
                ${core.homepage ? `
                    <div class="core-links">
                        <h4>相关链接</h4>
                        <a href="${core.homepage}" target="_blank" class="btn secondary">
                            <i class="mdi mdi-open-in-new"></i>
                            <span>官方主页</span>
                        </a>
                    </div>
                ` : ''}
            </div>
        `;
        
        uiManager.showModal(`${core.name} - 详细信息`, content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>关闭</span>
                </button>
                <button class="btn primary" onclick="window.getCoresPageManager().downloadCore('${core.id || core.name}'); window.getUIManager().closeModal();">
                    <i class="mdi mdi-download"></i>
                    <span>下载</span>
                </button>
            `
        });
    }
    
    // 下载核心
    downloadCore(coreId) {
        const core = this.cores.get(coreId);
        if (!core) return;
        
        const uiManager = window.getUIManager();
        uiManager?.showNotification('功能开发中', '核心下载功能正在开发中', 'info');
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

// 全局核心页面管理器实例
let coresPageManager = null;

// 获取核心页面管理器实例
function getCoresPageManager() {
    if (!coresPageManager) {
        coresPageManager = new CoresPageManager();
    }
    return coresPageManager;
}

// 导出到全局
window.getCoresPageManager = getCoresPageManager;
