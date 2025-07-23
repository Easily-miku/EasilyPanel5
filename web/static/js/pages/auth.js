// 2FA认证页面管理器
class AuthPageManager {
    constructor() {
        this.authManager = null;
        this.currentStep = 'overview';
        this.totpSecret = '';
        this.backupCodes = [];
    }
    
    // 初始化页面
    async init() {
        this.authManager = window.getAuthManager();
        if (!this.authManager) {
            console.error('Auth manager not available');
            return;
        }
        
        // 等待认证管理器初始化
        await this.authManager.init();
        
        this.renderPage();
        this.setupEventListeners();
    }
    
    // 渲染页面
    renderPage() {
        const authPage = document.getElementById('auth-page');
        if (!authPage) return;
        
        const user = this.authManager.getCurrentUser();
        if (!user) {
            authPage.innerHTML = this.renderNotAuthenticated();
            return;
        }
        
        authPage.innerHTML = this.renderAuthenticatedPage(user);
    }
    
    // 渲染未认证状态
    renderNotAuthenticated() {
        return `
            <div class="page-header">
                <h2>双因素认证</h2>
                <p>请先登录以管理双因素认证设置</p>
            </div>
            <div class="card">
                <div class="card-content">
                    <div class="auth-required">
                        <i class="mdi mdi-shield-lock" style="font-size: 4rem; color: var(--text-secondary); margin-bottom: 1rem;"></i>
                        <h3>需要登录</h3>
                        <p>您需要先登录才能访问双因素认证设置。</p>
                        <button class="btn primary" onclick="window.location.href='/login'">
                            <i class="mdi mdi-login"></i>
                            <span>前往登录</span>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染已认证页面
    renderAuthenticatedPage(user) {
        return `
            <div class="page-header">
                <h2>双因素认证</h2>
                <p>增强您的账户安全性</p>
            </div>
            
            <div class="auth-container">
                ${this.renderCurrentStatus(user)}
                ${this.renderSecurityLogs()}
            </div>
        `;
    }
    
    // 渲染当前状态
    renderCurrentStatus(user) {
        if (user.two_factor_enabled) {
            return this.renderEnabledStatus(user);
        } else {
            return this.renderDisabledStatus(user);
        }
    }
    
    // 渲染已启用状态
    renderEnabledStatus(user) {
        return `
            <div class="card">
                <div class="card-header">
                    <h3>双因素认证状态</h3>
                    <div class="status-badge success">
                        <i class="mdi mdi-shield-check"></i>
                        <span>已启用</span>
                    </div>
                </div>
                <div class="card-content">
                    <div class="auth-status-grid">
                        <div class="status-item">
                            <div class="status-icon success">
                                <i class="mdi mdi-shield-check"></i>
                            </div>
                            <div class="status-content">
                                <h4>双因素认证已启用</h4>
                                <p>您的账户受到额外的安全保护</p>
                            </div>
                        </div>
                        
                        <div class="auth-actions">
                            <button class="btn" id="regenerateBackupBtn">
                                <i class="mdi mdi-refresh"></i>
                                <span>重新生成备用码</span>
                            </button>
                            <button class="btn warning" id="disableTotpBtn">
                                <i class="mdi mdi-shield-off"></i>
                                <span>禁用双因素认证</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染未启用状态
    renderDisabledStatus(user) {
        return `
            <div class="card">
                <div class="card-header">
                    <h3>双因素认证状态</h3>
                    <div class="status-badge warning">
                        <i class="mdi mdi-shield-alert"></i>
                        <span>未启用</span>
                    </div>
                </div>
                <div class="card-content">
                    <div class="auth-status-grid">
                        <div class="status-item">
                            <div class="status-icon warning">
                                <i class="mdi mdi-shield-alert"></i>
                            </div>
                            <div class="status-content">
                                <h4>双因素认证未启用</h4>
                                <p>启用双因素认证以增强账户安全性</p>
                            </div>
                        </div>
                        
                        <div class="auth-setup-info">
                            <h4>什么是双因素认证？</h4>
                            <ul>
                                <li>在密码基础上增加额外的安全层</li>
                                <li>使用手机应用生成时间敏感的验证码</li>
                                <li>即使密码泄露，账户仍然安全</li>
                                <li>支持Google Authenticator、Authy等应用</li>
                            </ul>
                            
                            <button class="btn primary" id="setupTotpBtn">
                                <i class="mdi mdi-shield-plus"></i>
                                <span>启用双因素认证</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染安全日志
    renderSecurityLogs() {
        return `
            <div class="card">
                <div class="card-header">
                    <h3>安全日志</h3>
                    <button class="btn" id="refreshLogsBtn">
                        <i class="mdi mdi-refresh"></i>
                        <span>刷新</span>
                    </button>
                </div>
                <div class="card-content">
                    <div class="security-logs" id="securityLogs">
                        <div class="loading">
                            <i class="mdi mdi-loading mdi-spin"></i>
                            <span>加载中...</span>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 启用TOTP按钮
        const setupTotpBtn = document.getElementById('setupTotpBtn');
        if (setupTotpBtn) {
            setupTotpBtn.addEventListener('click', () => this.startTOTPSetup());
        }
        
        // 禁用TOTP按钮
        const disableTotpBtn = document.getElementById('disableTotpBtn');
        if (disableTotpBtn) {
            disableTotpBtn.addEventListener('click', () => this.showDisableTOTPDialog());
        }
        
        // 重新生成备用码按钮
        const regenerateBackupBtn = document.getElementById('regenerateBackupBtn');
        if (regenerateBackupBtn) {
            regenerateBackupBtn.addEventListener('click', () => this.regenerateBackupCodes());
        }
        
        // 刷新日志按钮
        const refreshLogsBtn = document.getElementById('refreshLogsBtn');
        if (refreshLogsBtn) {
            refreshLogsBtn.addEventListener('click', () => this.loadSecurityLogs());
        }
        
        // 加载安全日志
        this.loadSecurityLogs();
    }
    
    // 开始TOTP设置
    async startTOTPSetup() {
        const uiManager = window.getUIManager();
        
        try {
            const result = await this.authManager.setupTOTP();
            
            if (result.success) {
                this.totpSecret = result.secret;
                this.showTOTPSetupDialog(result);
            } else {
                uiManager?.showNotification('设置失败', result.message, 'error');
            }
        } catch (error) {
            console.error('TOTP setup failed:', error);
            uiManager?.showNotification('设置失败', '网络错误，请重试', 'error');
        }
    }
    
    // 显示TOTP设置对话框
    showTOTPSetupDialog(setupData) {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="totp-setup">
                <div class="setup-step">
                    <h4>步骤 1: 扫描二维码</h4>
                    <p>使用您的认证应用（如Google Authenticator、Authy）扫描下方二维码：</p>
                    <div class="qr-code-container">
                        <img src="https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(setupData.qr_code_url)}" 
                             alt="TOTP QR Code" class="qr-code">
                    </div>
                    <div class="manual-entry">
                        <p>或手动输入密钥：</p>
                        <code class="secret-key">${setupData.secret}</code>
                        <button class="btn-copy" onclick="navigator.clipboard.writeText('${setupData.secret}')">
                            <i class="mdi mdi-content-copy"></i>
                        </button>
                    </div>
                </div>
                
                <div class="setup-step">
                    <h4>步骤 2: 输入验证码</h4>
                    <p>输入认证应用显示的6位验证码：</p>
                    <div class="totp-input-group">
                        <input type="text" id="totpCode" class="totp-input" placeholder="000000" maxlength="6" pattern="[0-9]{6}">
                        <button class="btn primary" id="confirmTotpBtn">
                            <i class="mdi mdi-check"></i>
                            <span>确认</span>
                        </button>
                    </div>
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('设置双因素认证', content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
            `
        });
        
        if (modal) {
            // 设置确认按钮事件
            const confirmBtn = modal.querySelector('#confirmTotpBtn');
            const totpInput = modal.querySelector('#totpCode');
            
            const confirmSetup = async () => {
                const code = totpInput.value.trim();
                if (code.length !== 6) {
                    uiManager?.showNotification('验证失败', '请输入6位验证码', 'warning');
                    return;
                }
                
                try {
                    const result = await this.authManager.confirmTOTP(this.totpSecret, code);
                    
                    if (result.success) {
                        uiManager?.closeModal();
                        this.showBackupCodesDialog(result.backup_codes);
                        this.renderPage(); // 重新渲染页面
                        uiManager?.showNotification('设置成功', '双因素认证已启用', 'success');
                    } else {
                        uiManager?.showNotification('验证失败', result.message, 'error');
                    }
                } catch (error) {
                    console.error('TOTP confirmation failed:', error);
                    uiManager?.showNotification('验证失败', '网络错误，请重试', 'error');
                }
            };
            
            confirmBtn.addEventListener('click', confirmSetup);
            totpInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    confirmSetup();
                }
            });
            
            // 自动聚焦输入框
            totpInput.focus();
        }
    }
    
    // 显示备用码对话框
    showBackupCodesDialog(backupCodes) {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="backup-codes-display">
                <div class="warning-notice">
                    <i class="mdi mdi-alert"></i>
                    <div>
                        <h4>重要提醒</h4>
                        <p>请将这些备用码保存在安全的地方。每个备用码只能使用一次，当您无法使用认证应用时可以用它们登录。</p>
                    </div>
                </div>
                
                <div class="backup-codes-grid">
                    ${backupCodes.map(code => `
                        <div class="backup-code">
                            <code>${code}</code>
                        </div>
                    `).join('')}
                </div>
                
                <div class="backup-actions">
                    <button class="btn" onclick="window.print()">
                        <i class="mdi mdi-printer"></i>
                        <span>打印</span>
                    </button>
                    <button class="btn" onclick="navigator.clipboard.writeText('${backupCodes.join('\\n')}')">
                        <i class="mdi mdi-content-copy"></i>
                        <span>复制全部</span>
                    </button>
                </div>
            </div>
        `;
        
        uiManager?.showModal('备用码', content, {
            width: '500px',
            footer: `
                <button class="btn primary" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-check"></i>
                    <span>我已保存</span>
                </button>
            `
        });
    }
    
    // 显示禁用TOTP对话框
    showDisableTOTPDialog() {
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="disable-totp">
                <div class="warning-notice">
                    <i class="mdi mdi-alert"></i>
                    <div>
                        <h4>确认禁用双因素认证</h4>
                        <p>禁用双因素认证将降低您的账户安全性。请输入您的密码以确认此操作。</p>
                    </div>
                </div>
                
                <div class="form-group">
                    <label for="confirmPassword">当前密码</label>
                    <input type="password" id="confirmPassword" class="form-input" placeholder="请输入当前密码">
                </div>
            </div>
        `;
        
        const modal = uiManager?.showModal('禁用双因素认证', content, {
            width: '400px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>取消</span>
                </button>
                <button class="btn error" id="confirmDisableBtn">
                    <i class="mdi mdi-shield-off"></i>
                    <span>确认禁用</span>
                </button>
            `
        });
        
        if (modal) {
            const confirmBtn = modal.querySelector('#confirmDisableBtn');
            const passwordInput = modal.querySelector('#confirmPassword');
            
            const disableTOTP = async () => {
                const password = passwordInput.value.trim();
                if (!password) {
                    uiManager?.showNotification('验证失败', '请输入密码', 'warning');
                    return;
                }
                
                try {
                    const result = await this.authManager.disableTOTP(password);
                    
                    if (result.success) {
                        uiManager?.closeModal();
                        this.renderPage(); // 重新渲染页面
                        uiManager?.showNotification('禁用成功', '双因素认证已禁用', 'success');
                    } else {
                        uiManager?.showNotification('禁用失败', result.message, 'error');
                    }
                } catch (error) {
                    console.error('TOTP disable failed:', error);
                    uiManager?.showNotification('禁用失败', '网络错误，请重试', 'error');
                }
            };
            
            confirmBtn.addEventListener('click', disableTOTP);
            passwordInput.addEventListener('keypress', (e) => {
                if (e.key === 'Enter') {
                    disableTOTP();
                }
            });
            
            // 自动聚焦输入框
            passwordInput.focus();
        }
    }
    
    // 重新生成备用码
    async regenerateBackupCodes() {
        const uiManager = window.getUIManager();
        
        try {
            const result = await this.authManager.regenerateBackupCodes();
            
            if (result.success) {
                this.showBackupCodesDialog(result.backup_codes);
                uiManager?.showNotification('生成成功', '备用码已重新生成', 'success');
            } else {
                uiManager?.showNotification('生成失败', result.message, 'error');
            }
        } catch (error) {
            console.error('Backup codes regeneration failed:', error);
            uiManager?.showNotification('生成失败', '网络错误，请重试', 'error');
        }
    }
    
    // 加载安全日志
    async loadSecurityLogs() {
        const logsContainer = document.getElementById('securityLogs');
        if (!logsContainer) return;
        
        logsContainer.innerHTML = `
            <div class="loading">
                <i class="mdi mdi-loading mdi-spin"></i>
                <span>加载中...</span>
            </div>
        `;
        
        try {
            const result = await this.authManager.getSecurityLogs(20);
            
            if (result.success) {
                this.renderSecurityLogsContent(result.logs);
            } else {
                logsContainer.innerHTML = `
                    <div class="error-message">
                        <i class="mdi mdi-alert"></i>
                        <span>加载失败: ${result.message}</span>
                    </div>
                `;
            }
        } catch (error) {
            console.error('Failed to load security logs:', error);
            logsContainer.innerHTML = `
                <div class="error-message">
                    <i class="mdi mdi-alert"></i>
                    <span>网络错误，请重试</span>
                </div>
            `;
        }
    }
    
    // 渲染安全日志内容
    renderSecurityLogsContent(logs) {
        const logsContainer = document.getElementById('securityLogs');
        if (!logsContainer) return;
        
        if (logs.length === 0) {
            logsContainer.innerHTML = `
                <div class="empty-state">
                    <i class="mdi mdi-history"></i>
                    <span>暂无安全日志</span>
                </div>
            `;
            return;
        }
        
        const logItems = logs.map(log => {
            const date = new Date(log.created_at);
            const actionIcons = {
                'login': 'mdi-login',
                'logout': 'mdi-logout',
                'login_failed': 'mdi-alert',
                'password_changed': 'mdi-key',
                '2fa_enabled': 'mdi-shield-plus',
                '2fa_disabled': 'mdi-shield-minus',
                'backup_code_used': 'mdi-key-variant'
            };
            
            const actionNames = {
                'login': '登录成功',
                'logout': '登出',
                'login_failed': '登录失败',
                'password_changed': '密码修改',
                '2fa_enabled': '启用双因素认证',
                '2fa_disabled': '禁用双因素认证',
                'backup_code_used': '使用备用码'
            };
            
            return `
                <div class="log-item">
                    <div class="log-icon">
                        <i class="mdi ${actionIcons[log.action] || 'mdi-information'}"></i>
                    </div>
                    <div class="log-content">
                        <div class="log-action">${actionNames[log.action] || log.action}</div>
                        <div class="log-details">
                            <span class="log-time">${date.toLocaleString()}</span>
                            <span class="log-ip">${log.ip_address}</span>
                        </div>
                    </div>
                </div>
            `;
        }).join('');
        
        logsContainer.innerHTML = `
            <div class="logs-list">
                ${logItems}
            </div>
        `;
    }
}

// 全局认证页面管理器实例
let authPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    authPageManager = new AuthPageManager();
});

// 导出到全局作用域
window.AuthPageManager = AuthPageManager;
window.getAuthPageManager = () => authPageManager;
