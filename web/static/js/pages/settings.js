// 系统设置页面管理器
class SettingsPageManager {
    constructor() {
        this.settings = {};
        this.currentTab = 'general'; // general, security, notifications, advanced
        this.hasUnsavedChanges = false;
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadSettings();
    }
    
    // 渲染页面
    renderPage() {
        const settingsPage = document.getElementById('settings-page');
        if (!settingsPage) return;
        
        settingsPage.innerHTML = `
            <div class="page-header">
                <h2>系统设置</h2>
                <p>配置系统参数和偏好设置</p>
            </div>
            
            <div class="settings-container">
                ${this.renderSettingsTabs()}
                ${this.renderTabContent()}
                ${this.renderActionBar()}
            </div>
        `;
    }
    
    // 渲染设置标签页
    renderSettingsTabs() {
        return `
            <div class="settings-tabs">
                <button class="tab-btn ${this.currentTab === 'general' ? 'active' : ''}" data-tab="general">
                    <i class="mdi mdi-cog"></i>
                    <span>常规设置</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'security' ? 'active' : ''}" data-tab="security">
                    <i class="mdi mdi-shield-check"></i>
                    <span>安全设置</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'notifications' ? 'active' : ''}" data-tab="notifications">
                    <i class="mdi mdi-bell"></i>
                    <span>通知设置</span>
                </button>
                <button class="tab-btn ${this.currentTab === 'advanced' ? 'active' : ''}" data-tab="advanced">
                    <i class="mdi mdi-tune"></i>
                    <span>高级设置</span>
                </button>
            </div>
        `;
    }
    
    // 渲染标签页内容
    renderTabContent() {
        return `
            <div class="tab-content">
                <div class="tab-pane ${this.currentTab === 'general' ? 'active' : ''}" id="general-tab">
                    ${this.renderGeneralTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'security' ? 'active' : ''}" id="security-tab">
                    ${this.renderSecurityTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'notifications' ? 'active' : ''}" id="notifications-tab">
                    ${this.renderNotificationsTab()}
                </div>
                <div class="tab-pane ${this.currentTab === 'advanced' ? 'active' : ''}" id="advanced-tab">
                    ${this.renderAdvancedTab()}
                </div>
            </div>
        `;
    }
    
    // 渲染常规设置标签页
    renderGeneralTab() {
        return `
            <div class="settings-section">
                <h3>基本信息</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="systemName">系统名称</label>
                        <input type="text" id="systemName" class="form-input" placeholder="EasilyPanel5">
                        <small>显示在页面标题和导航栏中的系统名称</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="systemDescription">系统描述</label>
                        <textarea id="systemDescription" class="form-input" rows="3" placeholder="Minecraft服务器管理面板"></textarea>
                        <small>系统的简短描述</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="adminEmail">管理员邮箱</label>
                        <input type="email" id="adminEmail" class="form-input" placeholder="admin@example.com">
                        <small>用于接收系统通知和重要信息</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="timezone">时区</label>
                        <select id="timezone" class="form-input">
                            <option value="Asia/Shanghai">Asia/Shanghai (UTC+8)</option>
                            <option value="UTC">UTC (UTC+0)</option>
                            <option value="America/New_York">America/New_York (UTC-5)</option>
                            <option value="Europe/London">Europe/London (UTC+0)</option>
                        </select>
                        <small>系统显示时间的时区</small>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>界面设置</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="theme">主题</label>
                        <select id="theme" class="form-input">
                            <option value="light">浅色主题</option>
                            <option value="dark">深色主题</option>
                            <option value="auto">跟随系统</option>
                        </select>
                        <small>选择界面主题</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="language">语言</label>
                        <select id="language" class="form-input">
                            <option value="zh-CN">简体中文</option>
                            <option value="en-US">English</option>
                            <option value="ja-JP">日本語</option>
                        </select>
                        <small>界面显示语言</small>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="compactMode">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">紧凑模式</span>
                                <small>减少界面元素间距，显示更多内容</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="showAnimations">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">界面动画</span>
                                <small>启用界面过渡动画效果</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>默认设置</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="defaultServerMemory">默认服务器内存</label>
                        <select id="defaultServerMemory" class="form-input">
                            <option value="512M">512MB</option>
                            <option value="1G">1GB</option>
                            <option value="2G">2GB</option>
                            <option value="4G">4GB</option>
                            <option value="8G">8GB</option>
                        </select>
                        <small>创建新服务器时的默认内存分配</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="defaultServerCore">默认服务器核心</label>
                        <select id="defaultServerCore" class="form-input">
                            <option value="paper">Paper</option>
                            <option value="spigot">Spigot</option>
                            <option value="bukkit">Bukkit</option>
                            <option value="vanilla">Vanilla</option>
                        </select>
                        <small>创建新服务器时的默认核心类型</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="autoBackupInterval">自动备份间隔</label>
                        <select id="autoBackupInterval" class="form-input">
                            <option value="0">禁用</option>
                            <option value="1">每小时</option>
                            <option value="6">每6小时</option>
                            <option value="24">每天</option>
                            <option value="168">每周</option>
                        </select>
                        <small>自动备份服务器的时间间隔</small>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染安全设置标签页
    renderSecurityTab() {
        return `
            <div class="settings-section">
                <h3>登录安全</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="sessionTimeout">会话超时时间</label>
                        <select id="sessionTimeout" class="form-input">
                            <option value="30">30分钟</option>
                            <option value="60">1小时</option>
                            <option value="240">4小时</option>
                            <option value="480">8小时</option>
                            <option value="1440">24小时</option>
                        </select>
                        <small>用户无操作后自动登出的时间</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="maxLoginAttempts">最大登录尝试次数</label>
                        <input type="number" id="maxLoginAttempts" class="form-input" min="3" max="10" value="5">
                        <small>超过此次数将暂时锁定账户</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="lockoutDuration">锁定持续时间（分钟）</label>
                        <input type="number" id="lockoutDuration" class="form-input" min="5" max="60" value="15">
                        <small>账户被锁定后的解锁时间</small>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="requireStrongPassword">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">强制强密码</span>
                                <small>要求密码包含大小写字母、数字和特殊字符</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableTwoFactor">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">强制双因素认证</span>
                                <small>要求所有用户启用双因素认证</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>访问控制</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="allowedIPs">允许的IP地址</label>
                        <textarea id="allowedIPs" class="form-input" rows="4" placeholder="192.168.1.0/24&#10;10.0.0.0/8&#10;留空表示允许所有IP"></textarea>
                        <small>限制只有指定IP地址可以访问系统，每行一个IP或CIDR</small>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableAuditLog">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用审计日志</span>
                                <small>记录所有用户操作和系统事件</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableRateLimit">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用访问频率限制</span>
                                <small>防止暴力攻击和恶意请求</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>SSL/TLS设置</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="forceHTTPS">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">强制HTTPS</span>
                                <small>自动将HTTP请求重定向到HTTPS</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <label for="sslCertPath">SSL证书路径</label>
                        <input type="text" id="sslCertPath" class="form-input" placeholder="/path/to/cert.pem">
                        <small>SSL证书文件的完整路径</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="sslKeyPath">SSL私钥路径</label>
                        <input type="text" id="sslKeyPath" class="form-input" placeholder="/path/to/key.pem">
                        <small>SSL私钥文件的完整路径</small>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染通知设置标签页
    renderNotificationsTab() {
        return `
            <div class="settings-section">
                <h3>邮件通知</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableEmailNotifications">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用邮件通知</span>
                                <small>通过邮件发送系统通知</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <label for="smtpServer">SMTP服务器</label>
                        <input type="text" id="smtpServer" class="form-input" placeholder="smtp.gmail.com">
                        <small>邮件服务器地址</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="smtpPort">SMTP端口</label>
                        <input type="number" id="smtpPort" class="form-input" value="587">
                        <small>邮件服务器端口（通常为587或465）</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="smtpUsername">SMTP用户名</label>
                        <input type="text" id="smtpUsername" class="form-input" placeholder="your-email@gmail.com">
                        <small>邮件服务器登录用户名</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="smtpPassword">SMTP密码</label>
                        <input type="password" id="smtpPassword" class="form-input" placeholder="应用专用密码">
                        <small>邮件服务器登录密码</small>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="smtpTLS">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用TLS加密</span>
                                <small>使用TLS加密邮件传输</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>通知类型</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="notifyServerStart">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">服务器启动通知</span>
                                <small>服务器启动时发送通知</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="notifyServerStop">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">服务器停止通知</span>
                                <small>服务器停止时发送通知</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="notifyServerCrash">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">服务器崩溃通知</span>
                                <small>服务器异常停止时发送通知</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="notifyHighResource">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">高资源使用通知</span>
                                <small>CPU或内存使用率过高时发送通知</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="notifyBackupComplete">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">备份完成通知</span>
                                <small>自动备份完成时发送通知</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="notifySecurityEvents">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">安全事件通知</span>
                                <small>登录失败、权限变更等安全事件</small>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>Webhook通知</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableWebhook">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用Webhook通知</span>
                                <small>通过Webhook发送通知到第三方服务</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <label for="webhookUrl">Webhook URL</label>
                        <input type="url" id="webhookUrl" class="form-input" placeholder="https://hooks.slack.com/services/...">
                        <small>接收通知的Webhook地址</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="webhookSecret">Webhook密钥</label>
                        <input type="password" id="webhookSecret" class="form-input" placeholder="可选的验证密钥">
                        <small>用于验证Webhook请求的密钥</small>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染高级设置标签页
    renderAdvancedTab() {
        return `
            <div class="settings-section">
                <h3>系统性能</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="maxConcurrentServers">最大并发服务器数</label>
                        <input type="number" id="maxConcurrentServers" class="form-input" min="1" max="100" value="10">
                        <small>同时运行的服务器数量限制</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="systemResourceLimit">系统资源限制（%）</label>
                        <input type="number" id="systemResourceLimit" class="form-input" min="50" max="95" value="80">
                        <small>系统资源使用率上限，超过将限制新服务器启动</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="logRetentionDays">日志保留天数</label>
                        <input type="number" id="logRetentionDays" class="form-input" min="7" max="365" value="30">
                        <small>系统日志文件的保留时间</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="backupRetentionDays">备份保留天数</label>
                        <input type="number" id="backupRetentionDays" class="form-input" min="7" max="365" value="30">
                        <small>自动备份文件的保留时间</small>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>Java设置</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="javaPath">Java路径</label>
                        <input type="text" id="javaPath" class="form-input" placeholder="/usr/bin/java">
                        <small>Java可执行文件的路径，留空使用系统默认</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="defaultJavaArgs">默认Java参数</label>
                        <textarea id="defaultJavaArgs" class="form-input" rows="3" placeholder="-XX:+UseG1GC -XX:+UnlockExperimentalVMOptions"></textarea>
                        <small>创建新服务器时的默认JVM参数</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="minHeapSize">最小堆内存</label>
                        <input type="text" id="minHeapSize" class="form-input" value="512M">
                        <small>JVM最小堆内存大小</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="maxHeapSize">最大堆内存</label>
                        <input type="text" id="maxHeapSize" class="form-input" value="2G">
                        <small>JVM最大堆内存大小</small>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>存储设置</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <label for="serversPath">服务器存储路径</label>
                        <input type="text" id="serversPath" class="form-input" placeholder="/opt/easilypanel/servers">
                        <small>服务器文件的存储目录</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="backupsPath">备份存储路径</label>
                        <input type="text" id="backupsPath" class="form-input" placeholder="/opt/easilypanel/backups">
                        <small>备份文件的存储目录</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="templatesPath">模板存储路径</label>
                        <input type="text" id="templatesPath" class="form-input" placeholder="/opt/easilypanel/templates">
                        <small>服务器模板的存储目录</small>
                    </div>
                    
                    <div class="setting-item">
                        <label for="maxDiskUsage">最大磁盘使用率（%）</label>
                        <input type="number" id="maxDiskUsage" class="form-input" min="50" max="95" value="85">
                        <small>磁盘使用率上限，超过将停止创建新服务器</small>
                    </div>
                </div>
            </div>
            
            <div class="settings-section">
                <h3>调试选项</h3>
                <div class="settings-grid">
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableDebugMode">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用调试模式</span>
                                <small>输出详细的调试信息到日志</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <div class="setting-toggle">
                            <label class="switch">
                                <input type="checkbox" id="enableAPILogging">
                                <span class="slider"></span>
                            </label>
                            <div class="toggle-info">
                                <span class="toggle-label">启用API日志</span>
                                <small>记录所有API请求和响应</small>
                            </div>
                        </div>
                    </div>
                    
                    <div class="setting-item">
                        <label for="logLevel">日志级别</label>
                        <select id="logLevel" class="form-input">
                            <option value="error">Error</option>
                            <option value="warn">Warning</option>
                            <option value="info">Info</option>
                            <option value="debug">Debug</option>
                        </select>
                        <small>系统日志的详细程度</small>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染操作栏
    renderActionBar() {
        return `
            <div class="action-bar" id="actionBar" style="display: none;">
                <div class="action-info">
                    <i class="mdi mdi-information"></i>
                    <span>您有未保存的更改</span>
                </div>
                <div class="action-buttons">
                    <button class="btn" id="discardBtn">
                        <i class="mdi mdi-close"></i>
                        <span>放弃更改</span>
                    </button>
                    <button class="btn primary" id="saveBtn">
                        <i class="mdi mdi-content-save"></i>
                        <span>保存设置</span>
                    </button>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 标签页切换
        const tabBtns = document.querySelectorAll('.tab-btn');
        tabBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const tab = e.currentTarget.getAttribute('data-tab');
                this.switchTab(tab);
            });
        });
        
        // 保存和放弃按钮
        const saveBtn = document.getElementById('saveBtn');
        const discardBtn = document.getElementById('discardBtn');
        
        if (saveBtn) {
            saveBtn.addEventListener('click', () => this.saveSettings());
        }
        if (discardBtn) {
            discardBtn.addEventListener('click', () => this.discardChanges());
        }
        
        // 监听表单变化
        this.setupFormChangeListeners();
    }
    
    // 设置表单变化监听器
    setupFormChangeListeners() {
        const inputs = document.querySelectorAll('input, select, textarea');
        inputs.forEach(input => {
            input.addEventListener('change', () => {
                this.markAsChanged();
            });
            
            if (input.type === 'text' || input.tagName === 'TEXTAREA') {
                input.addEventListener('input', () => {
                    this.markAsChanged();
                });
            }
        });
    }
    
    // 切换标签页
    switchTab(tab) {
        this.currentTab = tab;
        
        // 更新标签按钮状态
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-tab="${tab}"]`).classList.add('active');
        
        // 更新标签页内容
        document.querySelectorAll('.tab-pane').forEach(pane => {
            pane.classList.remove('active');
        });
        document.getElementById(`${tab}-tab`).classList.add('active');
        
        // 重新设置表单监听器
        setTimeout(() => {
            this.setupFormChangeListeners();
        }, 100);
    }
    
    // 标记为已更改
    markAsChanged() {
        this.hasUnsavedChanges = true;
        const actionBar = document.getElementById('actionBar');
        if (actionBar) {
            actionBar.style.display = 'flex';
        }
    }
    
    // 清除更改标记
    clearChanges() {
        this.hasUnsavedChanges = false;
        const actionBar = document.getElementById('actionBar');
        if (actionBar) {
            actionBar.style.display = 'none';
        }
    }
    
    // 加载设置
    async loadSettings() {
        try {
            const response = await fetch('/api/settings');
            const result = await response.json();

            if (response.ok && result.success) {
                this.settings = result.data || {};
                this.populateForm();
            } else {
                const errorMessage = result.message || result.error || '加载设置失败';
                this.showError(errorMessage);
            }
        } catch (error) {
            console.error('Failed to load settings:', error);
            this.showError('网络错误，请重试');
        }
    }
    
    // 填充表单
    populateForm() {
        Object.keys(this.settings).forEach(key => {
            const element = document.getElementById(key);
            if (element) {
                if (element.type === 'checkbox') {
                    element.checked = this.settings[key];
                } else {
                    element.value = this.settings[key];
                }
            }
        });
        
        this.clearChanges();
    }
    
    // 收集表单数据
    collectFormData() {
        const formData = {};
        const inputs = document.querySelectorAll('input, select, textarea');
        
        inputs.forEach(input => {
            if (input.id) {
                if (input.type === 'checkbox') {
                    formData[input.id] = input.checked;
                } else {
                    formData[input.id] = input.value;
                }
            }
        });
        
        return formData;
    }
    
    // 保存设置
    async saveSettings() {
        const uiManager = window.getUIManager();

        try {
            const formData = this.collectFormData();

            const response = await fetch('/api/settings', {
                method: 'PUT',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(formData)
            });

            const result = await response.json();

            if (response.ok && result.success) {
                this.settings = formData;
                this.clearChanges();
                uiManager?.showNotification('保存成功', '设置已保存', 'success');
            } else {
                const errorMessage = result.message || result.error || '保存设置失败';
                uiManager?.showNotification('保存失败', errorMessage, 'error');
            }
        } catch (error) {
            console.error('Save settings failed:', error);
            uiManager?.showNotification('保存失败', '网络错误，请重试', 'error');
        }
    }
    
    // 放弃更改
    discardChanges() {
        this.populateForm();
        this.clearChanges();
    }
    
    // 显示错误信息
    showError(message) {
        const uiManager = window.getUIManager();
        uiManager?.showNotification('错误', message, 'error');
    }
}

// 全局设置页面管理器实例
let settingsPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    settingsPageManager = new SettingsPageManager();
});

// 导出到全局作用域
window.SettingsPageManager = SettingsPageManager;
window.getSettingsPageManager = () => settingsPageManager;
