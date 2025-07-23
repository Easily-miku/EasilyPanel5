// 服务器模板页面管理器
class TemplatesPageManager {
    constructor() {
        this.templates = new Map();
        this.selectedTemplates = new Set();
        this.currentCategory = 'all';
        this.searchQuery = '';
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadTemplates();
    }
    
    // 渲染页面
    renderPage() {
        const templatesPage = document.getElementById('templates-page');
        if (!templatesPage) return;
        
        templatesPage.innerHTML = `
            <div class="page-header">
                <h2>服务端市场</h2>
                <p>浏览和下载各种Minecraft服务端</p>
            </div>

            <div class="templates-container">
                ${this.renderDevelopmentNotice()}
            </div>
        `;
    }
    
    // 渲染开发中提示
    renderDevelopmentNotice() {
        return `
            <div class="development-notice">
                <div class="notice-icon">
                    <i class="mdi mdi-hammer-wrench"></i>
                </div>
                <div class="notice-content">
                    <h3>功能开发中</h3>
                    <p>服务端市场功能正在紧张开发中，敬请期待！</p>
                    <div class="notice-features">
                        <h4>即将推出的功能：</h4>
                        <ul>
                            <li><i class="mdi mdi-check-circle"></i> 官方服务端下载（Vanilla、Paper、Spigot等）</li>
                            <li><i class="mdi mdi-check-circle"></i> 模组包一键安装（Forge、Fabric）</li>
                            <li><i class="mdi mdi-check-circle"></i> 插件市场集成</li>
                            <li><i class="mdi mdi-check-circle"></i> 自定义服务端模板</li>
                            <li><i class="mdi mdi-check-circle"></i> 版本自动更新</li>
                        </ul>
                    </div>
                    <div class="notice-actions">
                        <button class="btn primary" onclick="window.getUIManager().navigateToPage('servers')">
                            <i class="mdi mdi-server"></i>
                            <span>前往实例管理</span>
                        </button>
                        <button class="btn" onclick="window.getUIManager().navigateToPage('dashboard')">
                            <i class="mdi mdi-view-dashboard"></i>
                            <span>返回仪表盘</span>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }

    // 渲染工具栏
    renderToolbar() {
        return `
            <div class="templates-toolbar">
                <div class="toolbar-left">
                    <div class="search-box">
                        <i class="mdi mdi-magnify"></i>
                        <input type="text" id="templateSearch" placeholder="搜索模板..." value="${this.searchQuery}">
                    </div>
                    
                    <div class="category-tabs">
                        <button class="category-tab ${this.currentCategory === 'all' ? 'active' : ''}" data-category="all">
                            <i class="mdi mdi-all-inclusive"></i>
                            <span>全部</span>
                        </button>
                        <button class="category-tab ${this.currentCategory === 'vanilla' ? 'active' : ''}" data-category="vanilla">
                            <i class="mdi mdi-minecraft"></i>
                            <span>原版</span>
                        </button>
                        <button class="category-tab ${this.currentCategory === 'modded' ? 'active' : ''}" data-category="modded">
                            <i class="mdi mdi-puzzle"></i>
                            <span>模组</span>
                        </button>
                        <button class="category-tab ${this.currentCategory === 'plugin' ? 'active' : ''}" data-category="plugin">
                            <i class="mdi mdi-power-plug"></i>
                            <span>插件</span>
                        </button>
                        <button class="category-tab ${this.currentCategory === 'custom' ? 'active' : ''}" data-category="custom">
                            <i class="mdi mdi-account"></i>
                            <span>自定义</span>
                        </button>
                    </div>
                </div>
                
                <div class="toolbar-right">
                    <button class="btn" id="importTemplateBtn">
                        <i class="mdi mdi-import"></i>
                        <span>导入模板</span>
                    </button>
                    <button class="btn primary" id="createTemplateBtn">
                        <i class="mdi mdi-plus"></i>
                        <span>创建模板</span>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 渲染模板列表
    renderTemplatesList() {
        return `
            <div class="templates-list-container">
                <div class="templates-list" id="templatesList">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载模板列表...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 搜索框
        const searchInput = document.getElementById('templateSearch');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => {
                this.searchQuery = e.target.value;
                this.filterAndRenderTemplates();
            });
        }
        
        // 分类标签
        const categoryTabs = document.querySelectorAll('.category-tab');
        categoryTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                const category = e.currentTarget.getAttribute('data-category');
                this.setActiveCategory(category);
            });
        });
        
        // 创建模板按钮
        const createBtn = document.getElementById('createTemplateBtn');
        if (createBtn) {
            createBtn.addEventListener('click', () => this.showCreateTemplateDialog());
        }
        
        // 导入模板按钮
        const importBtn = document.getElementById('importTemplateBtn');
        if (importBtn) {
            importBtn.addEventListener('click', () => this.showImportTemplateDialog());
        }
    }
    
    // 设置活动分类
    setActiveCategory(category) {
        this.currentCategory = category;
        
        // 更新标签状态
        document.querySelectorAll('.category-tab').forEach(tab => {
            tab.classList.remove('active');
        });
        document.querySelector(`[data-category="${category}"]`).classList.add('active');
        
        this.filterAndRenderTemplates();
    }
    
    // 加载模板列表
    async loadTemplates() {
        try {
            const response = await fetch('/api/templates');
            if (response.ok) {
                const templates = await response.json();
                this.templates.clear();
                
                templates.forEach(template => {
                    this.templates.set(template.id, template);
                });
                
                this.filterAndRenderTemplates();
            } else {
                this.showError('加载模板列表失败');
            }
        } catch (error) {
            console.error('Failed to load templates:', error);
            this.showError('网络错误，请重试');
        }
    }
    
    // 过滤和渲染模板
    filterAndRenderTemplates() {
        const templatesList = document.getElementById('templatesList');
        if (!templatesList) return;
        
        // 获取过滤后的模板
        let filteredTemplates = Array.from(this.templates.values());
        
        // 应用搜索过滤
        if (this.searchQuery) {
            const query = this.searchQuery.toLowerCase();
            filteredTemplates = filteredTemplates.filter(template => 
                template.name.toLowerCase().includes(query) ||
                template.description?.toLowerCase().includes(query) ||
                template.tags?.some(tag => tag.toLowerCase().includes(query))
            );
        }
        
        // 应用分类过滤
        if (this.currentCategory !== 'all') {
            filteredTemplates = filteredTemplates.filter(template => 
                template.category === this.currentCategory
            );
        }
        
        // 渲染模板卡片
        if (filteredTemplates.length === 0) {
            templatesList.innerHTML = this.renderEmptyState();
        } else {
            templatesList.innerHTML = filteredTemplates.map(template => 
                this.renderTemplateCard(template)
            ).join('');
        }
        
        // 重新绑定事件
        this.bindTemplateCardEvents();
    }
    
    // 渲染空状态
    renderEmptyState() {
        return `
            <div class="empty-state">
                <i class="mdi mdi-file-document-multiple-outline"></i>
                <h3>没有找到模板</h3>
                <p>当前分类下没有模板，或者搜索条件没有匹配的结果。</p>
                <button class="btn primary" onclick="window.getTemplatesPageManager().showCreateTemplateDialog()">
                    <i class="mdi mdi-plus"></i>
                    <span>创建第一个模板</span>
                </button>
            </div>
        `;
    }
    
    // 渲染模板卡片
    renderTemplateCard(template) {
        const categoryIcons = {
            'vanilla': 'mdi-minecraft',
            'modded': 'mdi-puzzle',
            'plugin': 'mdi-power-plug',
            'custom': 'mdi-account'
        };
        
        const categoryNames = {
            'vanilla': '原版',
            'modded': '模组',
            'plugin': '插件',
            'custom': '自定义'
        };
        
        return `
            <div class="template-card" data-template-id="${template.id}">
                <div class="template-card-header">
                    <div class="template-icon">
                        <i class="mdi ${categoryIcons[template.category] || 'mdi-file-document'}"></i>
                    </div>
                    <div class="template-info">
                        <h3 class="template-name">${template.name}</h3>
                        <p class="template-description">${template.description || '无描述'}</p>
                    </div>
                    <div class="template-category">
                        <span class="category-badge ${template.category}">
                            ${categoryNames[template.category] || template.category}
                        </span>
                    </div>
                </div>
                
                <div class="template-card-body">
                    <div class="template-details">
                        <div class="detail-item">
                            <span class="detail-label">版本</span>
                            <span class="detail-value">${template.version || 'Unknown'}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">核心</span>
                            <span class="detail-value">${template.core || 'Unknown'}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">内存</span>
                            <span class="detail-value">${template.memory || '1G'}</span>
                        </div>
                        <div class="detail-item">
                            <span class="detail-label">创建时间</span>
                            <span class="detail-value">${new Date(template.created_at).toLocaleDateString()}</span>
                        </div>
                    </div>
                    
                    ${template.tags && template.tags.length > 0 ? `
                        <div class="template-tags">
                            ${template.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
                        </div>
                    ` : ''}
                </div>
                
                <div class="template-card-footer">
                    <div class="template-actions">
                        <button class="btn primary template-action-btn" data-action="use" data-template-id="${template.id}">
                            <i class="mdi mdi-rocket-launch"></i>
                            <span>使用模板</span>
                        </button>
                        
                        <button class="btn template-action-btn" data-action="preview" data-template-id="${template.id}">
                            <i class="mdi mdi-eye"></i>
                            <span>预览</span>
                        </button>
                        
                        <div class="dropdown">
                            <button class="btn template-action-btn dropdown-toggle" data-template-id="${template.id}">
                                <i class="mdi mdi-dots-vertical"></i>
                            </button>
                            <div class="dropdown-menu">
                                <a href="#" class="dropdown-item" data-action="edit" data-template-id="${template.id}">
                                    <i class="mdi mdi-pencil"></i>
                                    <span>编辑模板</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="duplicate" data-template-id="${template.id}">
                                    <i class="mdi mdi-content-copy"></i>
                                    <span>复制模板</span>
                                </a>
                                <a href="#" class="dropdown-item" data-action="export" data-template-id="${template.id}">
                                    <i class="mdi mdi-export"></i>
                                    <span>导出模板</span>
                                </a>
                                <div class="dropdown-divider"></div>
                                <a href="#" class="dropdown-item danger" data-action="delete" data-template-id="${template.id}">
                                    <i class="mdi mdi-delete"></i>
                                    <span>删除模板</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 绑定模板卡片事件
    bindTemplateCardEvents() {
        // 操作按钮事件
        const actionBtns = document.querySelectorAll('.template-action-btn');
        actionBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const action = e.currentTarget.getAttribute('data-action');
                const templateId = e.currentTarget.getAttribute('data-template-id');
                if (action && templateId) {
                    this.handleTemplateAction(action, templateId);
                }
            });
        });
        
        // 下拉菜单项事件
        const dropdownItems = document.querySelectorAll('.dropdown-item');
        dropdownItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const action = e.currentTarget.getAttribute('data-action');
                const templateId = e.currentTarget.getAttribute('data-template-id');
                if (action && templateId) {
                    this.handleTemplateAction(action, templateId);
                }
            });
        });
    }
    
    // 处理模板操作
    async handleTemplateAction(action, templateId) {
        const template = this.templates.get(templateId);
        if (!template) return;
        
        const uiManager = window.getUIManager();
        
        try {
            switch (action) {
                case 'use':
                    this.showUseTemplateDialog(templateId);
                    break;
                case 'preview':
                    this.showTemplatePreview(templateId);
                    break;
                case 'edit':
                    this.showEditTemplateDialog(templateId);
                    break;
                case 'duplicate':
                    await this.duplicateTemplate(templateId);
                    break;
                case 'export':
                    await this.exportTemplate(templateId);
                    break;
                case 'delete':
                    this.showDeleteTemplateDialog(templateId);
                    break;
            }
        } catch (error) {
            console.error(`Template action ${action} failed:`, error);
            uiManager?.showNotification('操作失败', `${action} 操作失败: ${error.message}`, 'error');
        }
    }
    
    // 显示使用模板对话框
    showUseTemplateDialog(templateId) {
        const template = this.templates.get(templateId);
        if (!template) return;
        
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="use-template-form">
                <div class="template-preview">
                    <h4>${template.name}</h4>
                    <p>${template.description || '无描述'}</p>
                    <div class="template-specs">
                        <span>版本: ${template.version}</span>
                        <span>核心: ${template.core}</span>
                        <span>内存: ${template.memory}</span>
                    </div>
                </div>
                
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
                        <label for="serverPort">端口</label>
                        <input type="number" id="serverPort" class="form-input" value="25565" min="1024" max="65535">
                    </div>
                    
                    <div class="form-group">
                        <label for="maxPlayers">最大玩家数</label>
                        <input type="number" id="maxPlayers" class="form-input" value="20" min="1" max="1000">
                    </div>
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('使用模板创建服务器', content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn primary" id="confirmUseBtn">
                    <i class="mdi mdi-rocket-launch"></i>
                    <span>创建服务器</span>
                </button>
            `
        });
        
        if (modal) {
            const confirmBtn = modal.querySelector('#confirmUseBtn');
            confirmBtn.addEventListener('click', () => this.createServerFromTemplate(templateId));
        }
    }
    
    // 从模板创建服务器
    async createServerFromTemplate(templateId) {
        const template = this.templates.get(templateId);
        if (!template) return;
        
        const uiManager = window.getUIManager();
        
        const serverData = {
            name: document.getElementById('serverName').value.trim(),
            description: document.getElementById('serverDescription').value.trim(),
            port: parseInt(document.getElementById('serverPort').value),
            max_players: parseInt(document.getElementById('maxPlayers').value),
            template_id: templateId
        };
        
        if (!serverData.name) {
            uiManager?.showNotification('创建失败', '请输入服务器名称', 'warning');
            return;
        }
        
        try {
            const response = await fetch('/api/servers/from-template', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(serverData)
            });
            
            if (response.ok) {
                uiManager?.closeModal();
                uiManager?.showNotification('创建成功', '服务器创建成功', 'success');
                
                // 切换到服务器管理页面
                const mainUIManager = window.getUIManager();
                if (mainUIManager) {
                    mainUIManager.navigateToPage('servers');
                }
            } else {
                const error = await response.json();
                uiManager?.showNotification('创建失败', error.message || '创建服务器失败', 'error');
            }
        } catch (error) {
            console.error('Create server from template failed:', error);
            uiManager?.showNotification('创建失败', '网络错误，请重试', 'error');
        }
    }
    
    // 显示模板预览
    showTemplatePreview(templateId) {
        const template = this.templates.get(templateId);
        if (!template) return;
        
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="template-preview-detail">
                <div class="preview-header">
                    <h3>${template.name}</h3>
                    <span class="category-badge ${template.category}">${template.category}</span>
                </div>
                
                <div class="preview-description">
                    <p>${template.description || '无描述'}</p>
                </div>
                
                <div class="preview-specs">
                    <h4>配置规格</h4>
                    <div class="specs-grid">
                        <div class="spec-item">
                            <span class="spec-label">Minecraft版本</span>
                            <span class="spec-value">${template.version || 'Unknown'}</span>
                        </div>
                        <div class="spec-item">
                            <span class="spec-label">服务器核心</span>
                            <span class="spec-value">${template.core || 'Unknown'}</span>
                        </div>
                        <div class="spec-item">
                            <span class="spec-label">推荐内存</span>
                            <span class="spec-value">${template.memory || '1G'}</span>
                        </div>
                        <div class="spec-item">
                            <span class="spec-label">创建时间</span>
                            <span class="spec-value">${new Date(template.created_at).toLocaleString()}</span>
                        </div>
                    </div>
                </div>
                
                ${template.tags && template.tags.length > 0 ? `
                    <div class="preview-tags">
                        <h4>标签</h4>
                        <div class="tags-list">
                            ${template.tags.map(tag => `<span class="tag">${tag}</span>`).join('')}
                        </div>
                    </div>
                ` : ''}
                
                ${template.features && template.features.length > 0 ? `
                    <div class="preview-features">
                        <h4>特性</h4>
                        <ul class="features-list">
                            ${template.features.map(feature => `<li>${feature}</li>`).join('')}
                        </ul>
                    </div>
                ` : ''}
            </div>
        `;
        
        uiManager?.showModal('模板预览', content, {
            width: '700px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>关闭</span>
                </button>
                <button class="btn primary" onclick="window.getTemplatesPageManager().showUseTemplateDialog('${templateId}'); window.getUIManager().closeModal();">
                    <i class="mdi mdi-rocket-launch"></i>
                    <span>使用此模板</span>
                </button>
            `
        });
    }
    
    // 显示创建模板对话框
    showCreateTemplateDialog() {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="create-template-form">
                <div class="form-group">
                    <label for="templateName">模板名称</label>
                    <input type="text" id="templateName" class="form-input" placeholder="输入模板名称" required>
                </div>
                
                <div class="form-group">
                    <label for="templateDescription">模板描述</label>
                    <textarea id="templateDescription" class="form-input" placeholder="输入模板描述" rows="3"></textarea>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="templateCategory">分类</label>
                        <select id="templateCategory" class="form-input">
                            <option value="vanilla">原版</option>
                            <option value="modded">模组</option>
                            <option value="plugin">插件</option>
                            <option value="custom">自定义</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="templateVersion">Minecraft版本</label>
                        <select id="templateVersion" class="form-input">
                            <option value="1.20.1">1.20.1</option>
                            <option value="1.19.4">1.19.4</option>
                            <option value="1.18.2">1.18.2</option>
                            <option value="1.16.5">1.16.5</option>
                        </select>
                    </div>
                </div>
                
                <div class="form-row">
                    <div class="form-group">
                        <label for="templateCore">服务器核心</label>
                        <select id="templateCore" class="form-input">
                            <option value="paper">Paper</option>
                            <option value="spigot">Spigot</option>
                            <option value="bukkit">Bukkit</option>
                            <option value="vanilla">Vanilla</option>
                        </select>
                    </div>
                    
                    <div class="form-group">
                        <label for="templateMemory">推荐内存</label>
                        <select id="templateMemory" class="form-input">
                            <option value="512M">512MB</option>
                            <option value="1G" selected>1GB</option>
                            <option value="2G">2GB</option>
                            <option value="4G">4GB</option>
                            <option value="8G">8GB</option>
                        </select>
                    </div>
                </div>
                
                <div class="form-group">
                    <label for="templateTags">标签（用逗号分隔）</label>
                    <input type="text" id="templateTags" class="form-input" placeholder="例如: PVP, 生存, 创造">
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('创建模板', content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn primary" id="confirmCreateTemplateBtn">
                    <i class="mdi mdi-check"></i>
                    <span>创建模板</span>
                </button>
            `
        });
        
        if (modal) {
            const confirmBtn = modal.querySelector('#confirmCreateTemplateBtn');
            confirmBtn.addEventListener('click', () => this.createTemplate());
        }
    }
    
    // 创建模板
    async createTemplate() {
        const uiManager = window.getUIManager();
        
        const templateData = {
            name: document.getElementById('templateName').value.trim(),
            description: document.getElementById('templateDescription').value.trim(),
            category: document.getElementById('templateCategory').value,
            version: document.getElementById('templateVersion').value,
            core: document.getElementById('templateCore').value,
            memory: document.getElementById('templateMemory').value,
            tags: document.getElementById('templateTags').value.split(',').map(tag => tag.trim()).filter(tag => tag)
        };
        
        if (!templateData.name) {
            uiManager?.showNotification('创建失败', '请输入模板名称', 'warning');
            return;
        }
        
        try {
            const response = await fetch('/api/templates', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(templateData)
            });
            
            if (response.ok) {
                uiManager?.closeModal();
                uiManager?.showNotification('创建成功', '模板创建成功', 'success');
                await this.loadTemplates(); // 刷新列表
            } else {
                const error = await response.json();
                uiManager?.showNotification('创建失败', error.message || '创建模板失败', 'error');
            }
        } catch (error) {
            console.error('Create template failed:', error);
            uiManager?.showNotification('创建失败', '网络错误，请重试', 'error');
        }
    }
    
    // 显示导入模板对话框
    showImportTemplateDialog() {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="import-template-form">
                <div class="import-methods">
                    <div class="import-method active" data-method="file">
                        <i class="mdi mdi-file-upload"></i>
                        <span>从文件导入</span>
                    </div>
                    <div class="import-method" data-method="url">
                        <i class="mdi mdi-link"></i>
                        <span>从URL导入</span>
                    </div>
                </div>
                
                <div class="import-content">
                    <div class="import-file-content active">
                        <div class="file-drop-zone" id="fileDropZone">
                            <i class="mdi mdi-cloud-upload"></i>
                            <h4>拖拽文件到此处或点击选择</h4>
                            <p>支持 .json 格式的模板文件</p>
                            <input type="file" id="templateFile" accept=".json" style="display: none;">
                        </div>
                    </div>
                    
                    <div class="import-url-content">
                        <div class="form-group">
                            <label for="templateUrl">模板URL</label>
                            <input type="url" id="templateUrl" class="form-input" placeholder="输入模板文件的URL">
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('导入模板', content, {
            width: '500px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn primary" id="confirmImportBtn">
                    <i class="mdi mdi-import"></i>
                    <span>导入模板</span>
                </button>
            `
        });
        
        if (modal) {
            // 设置导入方法切换
            const importMethods = modal.querySelectorAll('.import-method');
            importMethods.forEach(method => {
                method.addEventListener('click', (e) => {
                    const methodType = e.currentTarget.getAttribute('data-method');
                    
                    // 更新方法状态
                    importMethods.forEach(m => m.classList.remove('active'));
                    e.currentTarget.classList.add('active');
                    
                    // 更新内容显示
                    modal.querySelectorAll('.import-file-content, .import-url-content').forEach(content => {
                        content.classList.remove('active');
                    });
                    modal.querySelector(`.import-${methodType}-content`).classList.add('active');
                });
            });
            
            // 文件拖拽处理
            const fileDropZone = modal.querySelector('#fileDropZone');
            const fileInput = modal.querySelector('#templateFile');
            
            fileDropZone.addEventListener('click', () => fileInput.click());
            fileDropZone.addEventListener('dragover', (e) => {
                e.preventDefault();
                fileDropZone.classList.add('dragover');
            });
            fileDropZone.addEventListener('dragleave', () => {
                fileDropZone.classList.remove('dragover');
            });
            fileDropZone.addEventListener('drop', (e) => {
                e.preventDefault();
                fileDropZone.classList.remove('dragover');
                const files = e.dataTransfer.files;
                if (files.length > 0) {
                    fileInput.files = files;
                }
            });
            
            const confirmBtn = modal.querySelector('#confirmImportBtn');
            confirmBtn.addEventListener('click', () => this.importTemplate());
        }
    }
    
    // 导入模板
    async importTemplate() {
        // TODO: 实现模板导入功能
        const uiManager = window.getUIManager();
        uiManager?.showNotification('功能开发中', '模板导入功能正在开发中', 'info');
    }
    
    // 显示错误信息
    showError(message) {
        const templatesList = document.getElementById('templatesList');
        if (templatesList) {
            templatesList.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h3>加载失败</h3>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getTemplatesPageManager().loadTemplates()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 其他操作方法（占位符）
    showEditTemplateDialog(templateId) {
        console.log('Edit template:', templateId);
    }
    
    async duplicateTemplate(templateId) {
        console.log('Duplicate template:', templateId);
    }
    
    async exportTemplate(templateId) {
        console.log('Export template:', templateId);
    }
    
    showDeleteTemplateDialog(templateId) {
        console.log('Delete template:', templateId);
    }
}

// 全局模板页面管理器实例
let templatesPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    templatesPageManager = new TemplatesPageManager();
});

// 导出到全局作用域
window.TemplatesPageManager = TemplatesPageManager;
window.getTemplatesPageManager = () => templatesPageManager;
