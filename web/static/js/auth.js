// EasilyPanel5 认证管理器
class AuthManager {
    constructor() {
        this.token = null;
        this.user = null;
        this.isAuthenticated = false;
        this.authEnabled = false;
        this.requireTwoFactor = false;
        this.loginCallback = null;
        this.logoutCallback = null;
        
        // 从localStorage恢复会话
        this.restoreSession();
    }
    
    // 初始化认证状态
    async init() {
        try {
            const response = await fetch('/api/auth/status');
            const data = await response.json();
            
            this.authEnabled = data.enabled;
            
            if (data.config) {
                this.requireTwoFactor = data.config.require_two_factor;
            }
            
            if (data.authenticated) {
                this.isAuthenticated = true;
                this.user = data.user;
                return true;
            } else {
                this.isAuthenticated = false;
                this.user = null;
                this.token = null;
                localStorage.removeItem('auth_token');
                return false;
            }
        } catch (error) {
            console.error('Failed to initialize auth status:', error);
            return false;
        }
    }
    
    // 登录
    async login(username, password, totpCode = '', backupCode = '', rememberMe = false) {
        try {
            const response = await fetch('/api/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    username,
                    password,
                    totp_code: totpCode,
                    backup_code: backupCode,
                    remember_me: rememberMe
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                this.token = data.token;
                this.user = data.user;
                this.isAuthenticated = true;
                
                // 保存令牌到localStorage（如果记住我）
                if (rememberMe) {
                    localStorage.setItem('auth_token', data.token);
                }
                
                // 调用登录回调
                if (this.loginCallback) {
                    this.loginCallback(this.user);
                }
                
                return {
                    success: true,
                    user: data.user
                };
            } else if (data.requires_totp) {
                return {
                    success: false,
                    requiresTOTP: true,
                    message: data.message
                };
            } else {
                return {
                    success: false,
                    message: data.message
                };
            }
        } catch (error) {
            console.error('Login failed:', error);
            return {
                success: false,
                message: 'Login failed due to a network error'
            };
        }
    }
    
    // 登出
    async logout() {
        if (!this.isAuthenticated) {
            return true;
        }
        
        try {
            const response = await fetch('/api/auth/logout', {
                method: 'POST',
                headers: this.getAuthHeaders()
            });
            
            const data = await response.json();
            
            if (data.success) {
                this.token = null;
                this.user = null;
                this.isAuthenticated = false;
                localStorage.removeItem('auth_token');
                
                // 调用登出回调
                if (this.logoutCallback) {
                    this.logoutCallback();
                }
                
                return true;
            } else {
                return false;
            }
        } catch (error) {
            console.error('Logout failed:', error);
            return false;
        }
    }
    
    // 设置TOTP
    async setupTOTP() {
        if (!this.isAuthenticated) {
            return {
                success: false,
                message: 'Authentication required'
            };
        }
        
        try {
            const response = await fetch('/api/auth/totp/setup', {
                method: 'POST',
                headers: this.getAuthHeaders()
            });
            
            return await response.json();
        } catch (error) {
            console.error('TOTP setup failed:', error);
            return {
                success: false,
                message: 'TOTP setup failed due to a network error'
            };
        }
    }
    
    // 确认TOTP设置
    async confirmTOTP(secret, code) {
        if (!this.isAuthenticated) {
            return {
                success: false,
                message: 'Authentication required'
            };
        }
        
        try {
            const response = await fetch('/api/auth/totp/confirm', {
                method: 'POST',
                headers: {
                    ...this.getAuthHeaders(),
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    secret,
                    code
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                // 更新用户信息
                this.user.two_factor_enabled = true;
                
                // 刷新用户资料
                this.refreshProfile();
            }
            
            return data;
        } catch (error) {
            console.error('TOTP confirmation failed:', error);
            return {
                success: false,
                message: 'TOTP confirmation failed due to a network error'
            };
        }
    }
    
    // 禁用TOTP
    async disableTOTP(password) {
        if (!this.isAuthenticated) {
            return {
                success: false,
                message: 'Authentication required'
            };
        }
        
        try {
            const response = await fetch('/api/auth/totp/disable', {
                method: 'POST',
                headers: {
                    ...this.getAuthHeaders(),
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    password
                })
            });
            
            const data = await response.json();
            
            if (data.success) {
                // 更新用户信息
                this.user.two_factor_enabled = false;
                
                // 刷新用户资料
                this.refreshProfile();
            }
            
            return data;
        } catch (error) {
            console.error('TOTP disable failed:', error);
            return {
                success: false,
                message: 'TOTP disable failed due to a network error'
            };
        }
    }
    
    // 重新生成备用码
    async regenerateBackupCodes() {
        if (!this.isAuthenticated) {
            return {
                success: false,
                message: 'Authentication required'
            };
        }
        
        try {
            const response = await fetch('/api/auth/backup-codes/regenerate', {
                method: 'POST',
                headers: this.getAuthHeaders()
            });
            
            return await response.json();
        } catch (error) {
            console.error('Backup codes regeneration failed:', error);
            return {
                success: false,
                message: 'Backup codes regeneration failed due to a network error'
            };
        }
    }
    
    // 修改密码
    async changePassword(currentPassword, newPassword, totpCode = '') {
        if (!this.isAuthenticated) {
            return {
                success: false,
                message: 'Authentication required'
            };
        }
        
        try {
            const response = await fetch('/api/auth/password/change', {
                method: 'POST',
                headers: {
                    ...this.getAuthHeaders(),
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify({
                    current_password: currentPassword,
                    new_password: newPassword,
                    totp_code: totpCode
                })
            });
            
            return await response.json();
        } catch (error) {
            console.error('Password change failed:', error);
            return {
                success: false,
                message: 'Password change failed due to a network error'
            };
        }
    }
    
    // 获取用户资料
    async getProfile() {
        if (!this.isAuthenticated) {
            return null;
        }
        
        try {
            const response = await fetch('/api/auth/profile', {
                headers: this.getAuthHeaders()
            });
            
            if (response.ok) {
                const profile = await response.json();
                this.user = profile;
                return profile;
            } else {
                return null;
            }
        } catch (error) {
            console.error('Failed to get profile:', error);
            return null;
        }
    }
    
    // 刷新用户资料
    async refreshProfile() {
        return await this.getProfile();
    }
    
    // 获取安全日志
    async getSecurityLogs(limit = 50) {
        if (!this.isAuthenticated) {
            return {
                success: false,
                message: 'Authentication required'
            };
        }
        
        try {
            const response = await fetch(`/api/auth/security-logs?limit=${limit}`, {
                headers: this.getAuthHeaders()
            });
            
            return await response.json();
        } catch (error) {
            console.error('Failed to get security logs:', error);
            return {
                success: false,
                message: 'Failed to get security logs due to a network error'
            };
        }
    }
    
    // 从localStorage恢复会话
    restoreSession() {
        const token = localStorage.getItem('auth_token');
        if (token) {
            this.token = token;
        }
    }
    
    // 获取认证头
    getAuthHeaders() {
        if (this.token) {
            return {
                'Authorization': `Bearer ${this.token}`
            };
        }
        return {};
    }
    
    // 设置登录回调
    setLoginCallback(callback) {
        this.loginCallback = callback;
    }
    
    // 设置登出回调
    setLogoutCallback(callback) {
        this.logoutCallback = callback;
    }
    
    // 检查是否已认证
    isUserAuthenticated() {
        return this.isAuthenticated;
    }
    
    // 获取当前用户
    getCurrentUser() {
        return this.user;
    }
    
    // 检查是否是管理员
    isAdmin() {
        return this.isAuthenticated && this.user && this.user.role === 'admin';
    }
}

// 全局认证管理器实例
let authManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    authManager = new AuthManager();
    
    // 初始化认证状态
    authManager.init().then(isAuthenticated => {
        console.log('Auth initialized, authenticated:', isAuthenticated);
    });
});

// 导出到全局作用域
window.AuthManager = AuthManager;
window.getAuthManager = () => authManager;
