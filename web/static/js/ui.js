// UI管理器 - 负责界面交互和状态管理
class UIManager {
    constructor() {
        this.currentPage = 'dashboard';
        this.sidebarPinned = false;
        this.theme = localStorage.getItem('theme') || 'light';
        this.notifications = [];
        this.modals = [];
        
        this.init();
    }
    
    init() {
        this.initTheme();
        this.initSidebar();
        this.initNavigation();
        this.initNotifications();
        this.initModals();
        this.initQuickActions();
        
        console.log('UI管理器已初始化');
    }
    
    // 主题管理
    initTheme() {
        document.documentElement.setAttribute('data-theme', this.theme);
        
        const themeToggle = document.getElementById('themeToggle');
        if (themeToggle) {
            themeToggle.addEventListener('click', () => {
                this.toggleTheme();
            });
        }
    }
    
    toggleTheme() {
        this.theme = this.theme === 'light' ? 'dark' : 'light';
        document.documentElement.setAttribute('data-theme', this.theme);
        localStorage.setItem('theme', this.theme);
        
        this.showNotification('主题已切换', `已切换到${this.theme === 'light' ? '浅色' : '深色'}主题`, 'info');
    }
    
    // 侧边栏管理
    initSidebar() {
        const sidebar = document.getElementById('sidebar');
        const sidebarToggle = document.getElementById('sidebarToggle');
        const sidebarOverlay = document.getElementById('sidebarOverlay');
        
        if (sidebarToggle) {
            sidebarToggle.addEventListener('click', () => {
                this.toggleSidebarPin();
            });
        }
        
        if (sidebarOverlay) {
            sidebarOverlay.addEventListener('click', () => {
                this.closeSidebar();
            });
        }
        
        // 移动端处理
        if (window.innerWidth <= 768) {
            this.setupMobileSidebar();
        }
        
        window.addEventListener('resize', () => {
            if (window.innerWidth <= 768) {
                this.setupMobileSidebar();
            } else {
                this.setupDesktopSidebar();
            }
        });
    }
    
    toggleSidebarPin() {
        const sidebar = document.getElementById('sidebar');
        const sidebarToggle = document.getElementById('sidebarToggle');
        
        this.sidebarPinned = !this.sidebarPinned;
        
        if (this.sidebarPinned) {
            sidebar.classList.add('pinned');
            sidebarToggle.innerHTML = '<i class="mdi mdi-pin-off"></i>';
            sidebarToggle.title = '取消固定侧边栏';
        } else {
            sidebar.classList.remove('pinned');
            sidebarToggle.innerHTML = '<i class="mdi mdi-pin"></i>';
            sidebarToggle.title = '固定侧边栏';
        }
        
        localStorage.setItem('sidebarPinned', this.sidebarPinned);
    }
    
    setupMobileSidebar() {
        const sidebar = document.getElementById('sidebar');
        const sidebarOverlay = document.getElementById('sidebarOverlay');
        
        // 移动端点击导航项后自动关闭侧边栏
        const navItems = sidebar.querySelectorAll('.nav-item');
        navItems.forEach(item => {
            item.addEventListener('click', () => {
                this.closeSidebar();
            });
        });
    }
    
    setupDesktopSidebar() {
        const sidebar = document.getElementById('sidebar');
        const sidebarOverlay = document.getElementById('sidebarOverlay');
        
        sidebar.classList.remove('expanded');
        sidebarOverlay.classList.remove('active');
    }
    
    openSidebar() {
        const sidebar = document.getElementById('sidebar');
        const sidebarOverlay = document.getElementById('sidebarOverlay');
        
        sidebar.classList.add('expanded');
        if (window.innerWidth <= 768) {
            sidebarOverlay.classList.add('active');
        }
    }
    
    closeSidebar() {
        const sidebar = document.getElementById('sidebar');
        const sidebarOverlay = document.getElementById('sidebarOverlay');
        
        sidebar.classList.remove('expanded');
        sidebarOverlay.classList.remove('active');
    }
    
    // 导航管理
    initNavigation() {
        const navItems = document.querySelectorAll('.nav-item');
        
        navItems.forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const page = item.getAttribute('data-page');
                if (page) {
                    this.navigateToPage(page);
                }
            });
        });
        
        // 处理浏览器前进后退
        window.addEventListener('popstate', (e) => {
            const page = e.state?.page || 'dashboard';
            this.navigateToPage(page, false);
        });
        
        // 初始页面
        const hash = window.location.hash.slice(1);
        if (hash) {
            this.navigateToPage(hash, false);
        }
    }
    
    navigateToPage(page, pushState = true) {
        if (page === this.currentPage) return;
        
        // 更新导航状态
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        
        const activeNavItem = document.querySelector(`[data-page="${page}"]`);
        if (activeNavItem) {
            activeNavItem.classList.add('active');
        }
        
        // 更新页面内容
        document.querySelectorAll('.page').forEach(pageEl => {
            pageEl.classList.remove('active');
        });
        
        const targetPage = document.getElementById(`${page}-page`);
        if (targetPage) {
            targetPage.classList.add('active');
        }
        
        // 更新面包屑
        this.updateBreadcrumb(page);
        
        // 更新URL
        if (pushState) {
            history.pushState({ page }, '', `#${page}`);
        }
        
        this.currentPage = page;
        
        // 加载页面内容
        this.loadPageContent(page);
        
        console.log(`导航到页面: ${page}`);
    }
    
    updateBreadcrumb(page) {
        const breadcrumb = document.getElementById('breadcrumb');
        const pageNames = {
            'dashboard': '仪表盘',
            'servers': '服务器管理',
            'groups': '服务器分组',
            'frp': '内网穿透',
            'files': '文件管理',
            'monitoring': '监控面板',
            'download': '下载中心',
            'settings': '系统设置',
            'auth': '双因素认证'
        };
        
        if (breadcrumb) {
            breadcrumb.innerHTML = `<span class="breadcrumb-item active">${pageNames[page] || page}</span>`;
        }
    }
    
    loadPageContent(page) {
        // 根据页面类型加载相应内容
        switch (page) {
            case 'dashboard':
                this.loadDashboard();
                break;
            case 'servers':
                this.loadServersPage();
                break;
            case 'groups':
                this.loadGroupsPage();
                break;
            case 'frp':
                this.loadFRPPage();
                break;
            case 'files':
                this.loadFilesPage();
                break;
            case 'monitoring':
                this.loadMonitoringPage();
                break;
            case 'download':
                this.loadDownloadPage();
                break;
            case 'settings':
                this.loadSettingsPage();
                break;
            case 'auth':
                this.loadAuthPage();
                break;
        }
    }
    
    // 通知系统
    initNotifications() {
        this.notificationContainer = document.getElementById('notificationContainer');
    }
    
    showNotification(title, message, type = 'info', duration = 5000) {
        const notification = this.createNotification(title, message, type);
        this.notificationContainer.appendChild(notification);
        
        // 自动移除
        if (duration > 0) {
            setTimeout(() => {
                this.removeNotification(notification);
            }, duration);
        }
        
        return notification;
    }
    
    createNotification(title, message, type) {
        const notification = document.createElement('div');
        notification.className = `notification ${type}`;
        
        const icons = {
            success: 'mdi-check-circle',
            warning: 'mdi-alert',
            error: 'mdi-alert-circle',
            info: 'mdi-information'
        };
        
        notification.innerHTML = `
            <div class="notification-icon">
                <i class="mdi ${icons[type] || icons.info}"></i>
            </div>
            <div class="notification-content">
                <div class="notification-title">${title}</div>
                <div class="notification-message">${message}</div>
            </div>
            <button class="notification-close">
                <i class="mdi mdi-close"></i>
            </button>
        `;
        
        const closeBtn = notification.querySelector('.notification-close');
        closeBtn.addEventListener('click', () => {
            this.removeNotification(notification);
        });
        
        return notification;
    }
    
    removeNotification(notification) {
        notification.style.animation = 'slideOutRight 0.3s ease forwards';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }
    
    // 模态框系统
    initModals() {
        this.modalContainer = document.getElementById('modalContainer');
        
        // 点击遮罩层关闭模态框
        this.modalContainer.addEventListener('click', (e) => {
            if (e.target === this.modalContainer) {
                this.closeModal();
            }
        });
        
        // ESC键关闭模态框
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.modalContainer.classList.contains('active')) {
                this.closeModal();
            }
        });
    }
    
    showModal(title, content, options = {}) {
        const modal = this.createModal(title, content, options);
        this.modalContainer.innerHTML = '';
        this.modalContainer.appendChild(modal);
        this.modalContainer.classList.add('active');
        
        return modal;
    }
    
    createModal(title, content, options) {
        const modal = document.createElement('div');
        modal.className = 'modal';
        modal.style.width = options.width || '600px';
        
        modal.innerHTML = `
            <div class="modal-header">
                <h3 class="modal-title">${title}</h3>
                <button class="modal-close">
                    <i class="mdi mdi-close"></i>
                </button>
            </div>
            <div class="modal-body">
                ${content}
            </div>
            ${options.footer ? `<div class="modal-footer">${options.footer}</div>` : ''}
        `;
        
        const closeBtn = modal.querySelector('.modal-close');
        closeBtn.addEventListener('click', () => {
            this.closeModal();
        });
        
        return modal;
    }
    
    closeModal() {
        this.modalContainer.classList.remove('active');
        setTimeout(() => {
            this.modalContainer.innerHTML = '';
        }, 300);
    }
    
    // 快速操作
    initQuickActions() {
        const quickActionBtns = document.querySelectorAll('.quick-action-btn');
        
        quickActionBtns.forEach(btn => {
            btn.addEventListener('click', () => {
                const action = btn.getAttribute('data-action');
                this.handleQuickAction(action);
            });
        });
    }
    
    handleQuickAction(action) {
        switch (action) {
            case 'create-server':
                this.navigateToPage('servers');
                // TODO: 打开创建服务器对话框
                break;
            case 'start-all':
                this.showNotification('批量操作', '正在启动所有服务器...', 'info');
                // TODO: 实现批量启动
                break;
            case 'backup-all':
                this.showNotification('备份操作', '正在备份所有服务器...', 'info');
                // TODO: 实现批量备份
                break;
            case 'system-info':
                this.showSystemInfo();
                break;
        }
    }
    
    showSystemInfo() {
        const content = `
            <div class="system-info">
                <h4>系统信息</h4>
                <p>版本: EasilyPanel5 v1.1.0</p>
                <p>运行时间: ${this.getUptime()}</p>
                <p>浏览器: ${navigator.userAgent}</p>
            </div>
        `;
        
        this.showModal('系统信息', content);
    }
    
    getUptime() {
        // 简单的运行时间计算
        const startTime = new Date(document.readyState === 'complete' ? Date.now() : performance.timing.loadEventEnd);
        const uptime = Date.now() - startTime;
        const minutes = Math.floor(uptime / 60000);
        const seconds = Math.floor((uptime % 60000) / 1000);
        return `${minutes}分${seconds}秒`;
    }
    
    // 页面加载方法（占位符，将在后续实现）
    loadDashboard() {
        console.log('加载仪表盘');
        // 仪表盘内容已在HTML中定义，数据由app.js中的数据更新方法处理

        // 如果应用实例存在，触发数据刷新
        if (window.app) {
            window.app.refreshData();
        }
    }
    
    loadServersPage() {
        console.log('加载服务器管理页面');

        // 等待服务器页面管理器初始化
        if (window.getServersPageManager) {
            const serversPageManager = window.getServersPageManager();
            if (serversPageManager) {
                serversPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const serversPageManager = window.getServersPageManager();
                    if (serversPageManager) {
                        serversPageManager.init();
                    }
                }, 100);
            }
        }
    }
    

    loadFRPPage() {
        console.log('加载FRP页面');

        // 等待FRP页面管理器初始化
        if (window.getFRPPageManager) {
            const frpPageManager = window.getFRPPageManager();
            if (frpPageManager) {
                frpPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const frpPageManager = window.getFRPPageManager();
                    if (frpPageManager) {
                        frpPageManager.init();
                    }
                }, 100);
            }
        }
    }
    
    loadFilesPage() {
        console.log('加载文件管理页面');

        // 等待文件页面管理器初始化
        if (window.getFilesPageManager) {
            const filesPageManager = window.getFilesPageManager();
            if (filesPageManager) {
                filesPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const filesPageManager = window.getFilesPageManager();
                    if (filesPageManager) {
                        filesPageManager.init();
                    }
                }, 100);
            }
        }
    }
    
    loadMonitoringPage() {
        console.log('加载监控页面');

        // 等待监控页面管理器初始化
        if (window.getMonitoringPageManager) {
            const monitoringPageManager = window.getMonitoringPageManager();
            if (monitoringPageManager) {
                monitoringPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const monitoringPageManager = window.getMonitoringPageManager();
                    if (monitoringPageManager) {
                        monitoringPageManager.init();
                    }
                }, 100);
            }
        }
    }
    
    loadDownloadPage() {
        console.log('加载下载页面');

        // 等待下载页面管理器初始化
        if (window.getDownloadPageManager) {
            const downloadPageManager = window.getDownloadPageManager();
            if (downloadPageManager) {
                downloadPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const downloadPageManager = window.getDownloadPageManager();
                    if (downloadPageManager) {
                        downloadPageManager.init();
                    }
                }, 100);
            }
        }
    }

    loadSettingsPage() {
        console.log('加载设置页面');

        // 等待设置页面管理器初始化
        if (window.getSettingsPageManager) {
            const settingsPageManager = window.getSettingsPageManager();
            if (settingsPageManager) {
                settingsPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const settingsPageManager = window.getSettingsPageManager();
                    if (settingsPageManager) {
                        settingsPageManager.init();
                    }
                }, 100);
            }
        }
    }
    
    loadAuthPage() {
        console.log('加载认证页面');

        // 等待认证页面管理器初始化
        if (window.getAuthPageManager) {
            const authPageManager = window.getAuthPageManager();
            if (authPageManager) {
                authPageManager.init();
            } else {
                // 如果还没有初始化，等待一下再试
                setTimeout(() => {
                    const authPageManager = window.getAuthPageManager();
                    if (authPageManager) {
                        authPageManager.init();
                    }
                }, 100);
            }
        }
    }
}

// 全局UI管理器实例
let uiManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    uiManager = new UIManager();
    
    // 添加CSS动画
    const style = document.createElement('style');
    style.textContent = `
        @keyframes slideOutRight {
            from {
                transform: translateX(0);
                opacity: 1;
            }
            to {
                transform: translateX(100%);
                opacity: 0;
            }
        }
    `;
    document.head.appendChild(style);
});

// 导出到全局作用域
window.UIManager = UIManager;
window.getUIManager = () => uiManager;
