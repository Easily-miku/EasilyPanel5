// 监控面板页面管理器
class MonitoringPageManager {
    constructor() {
        this.systemStats = null;
        this.serverStats = new Map();
        this.refreshInterval = null;
        this.refreshRate = 5000; // 5秒刷新一次
        this.charts = new Map();
        this.currentTimeRange = '1h'; // 1h, 6h, 24h, 7d
    }
    
    // 初始化页面
    async init() {
        this.renderPage();
        this.setupEventListeners();
        await this.loadData();
        this.startAutoRefresh();
    }
    
    // 销毁页面
    destroy() {
        this.stopAutoRefresh();
        this.charts.clear();
    }
    
    // 渲染页面
    renderPage() {
        const monitoringPage = document.getElementById('monitoring-page');
        if (!monitoringPage) return;
        
        monitoringPage.innerHTML = `
            <div class="page-header">
                <h2>监控面板</h2>
                <p>实时监控系统和服务器状态</p>
            </div>
            
            <div class="monitoring-container">
                ${this.renderControlPanel()}
                ${this.renderSystemOverview()}
                ${this.renderServersList()}
                ${this.renderChartsSection()}
            </div>
        `;
    }
    
    // 渲染控制面板
    renderControlPanel() {
        return `
            <div class="control-panel">
                <div class="control-left">
                    <div class="refresh-controls">
                        <button class="btn" id="refreshBtn">
                            <i class="mdi mdi-refresh"></i>
                            <span>刷新</span>
                        </button>
                        
                        <div class="auto-refresh">
                            <label class="switch">
                                <input type="checkbox" id="autoRefreshToggle" checked>
                                <span class="slider"></span>
                            </label>
                            <span>自动刷新</span>
                        </div>
                        
                        <select id="refreshRate" class="refresh-rate-select">
                            <option value="1000">1秒</option>
                            <option value="5000" selected>5秒</option>
                            <option value="10000">10秒</option>
                            <option value="30000">30秒</option>
                            <option value="60000">1分钟</option>
                        </select>
                    </div>
                </div>
                
                <div class="control-right">
                    <div class="time-range-selector">
                        <button class="time-btn ${this.currentTimeRange === '1h' ? 'active' : ''}" data-range="1h">1小时</button>
                        <button class="time-btn ${this.currentTimeRange === '6h' ? 'active' : ''}" data-range="6h">6小时</button>
                        <button class="time-btn ${this.currentTimeRange === '24h' ? 'active' : ''}" data-range="24h">24小时</button>
                        <button class="time-btn ${this.currentTimeRange === '7d' ? 'active' : ''}" data-range="7d">7天</button>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染系统概览
    renderSystemOverview() {
        return `
            <div class="system-overview">
                <h3>系统概览</h3>
                <div class="overview-grid" id="systemOverview">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载系统信息...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染服务器列表
    renderServersList() {
        return `
            <div class="servers-monitoring">
                <h3>服务器状态</h3>
                <div class="servers-grid" id="serversMonitoring">
                    <div class="loading">
                        <i class="mdi mdi-loading mdi-spin"></i>
                        <span>加载服务器状态...</span>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 渲染图表区域
    renderChartsSection() {
        return `
            <div class="charts-section">
                <h3>性能图表</h3>
                <div class="charts-grid">
                    <div class="chart-container">
                        <div class="chart-header">
                            <h4>CPU使用率</h4>
                            <span class="chart-value" id="cpuValue">--%</span>
                        </div>
                        <div class="chart-content">
                            <canvas id="cpuChart" width="400" height="200"></canvas>
                        </div>
                    </div>
                    
                    <div class="chart-container">
                        <div class="chart-header">
                            <h4>内存使用率</h4>
                            <span class="chart-value" id="memoryValue">--%</span>
                        </div>
                        <div class="chart-content">
                            <canvas id="memoryChart" width="400" height="200"></canvas>
                        </div>
                    </div>
                    
                    <div class="chart-container">
                        <div class="chart-header">
                            <h4>磁盘使用率</h4>
                            <span class="chart-value" id="diskValue">--%</span>
                        </div>
                        <div class="chart-content">
                            <canvas id="diskChart" width="400" height="200"></canvas>
                        </div>
                    </div>
                    
                    <div class="chart-container">
                        <div class="chart-header">
                            <h4>网络流量</h4>
                            <span class="chart-value" id="networkValue">-- MB/s</span>
                        </div>
                        <div class="chart-content">
                            <canvas id="networkChart" width="400" height="200"></canvas>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }
    
    // 设置事件监听器
    setupEventListeners() {
        // 刷新按钮
        const refreshBtn = document.getElementById('refreshBtn');
        if (refreshBtn) {
            refreshBtn.addEventListener('click', () => this.loadData());
        }
        
        // 自动刷新开关
        const autoRefreshToggle = document.getElementById('autoRefreshToggle');
        if (autoRefreshToggle) {
            autoRefreshToggle.addEventListener('change', (e) => {
                if (e.target.checked) {
                    this.startAutoRefresh();
                } else {
                    this.stopAutoRefresh();
                }
            });
        }
        
        // 刷新频率选择
        const refreshRateSelect = document.getElementById('refreshRate');
        if (refreshRateSelect) {
            refreshRateSelect.addEventListener('change', (e) => {
                this.refreshRate = parseInt(e.target.value);
                if (this.refreshInterval) {
                    this.stopAutoRefresh();
                    this.startAutoRefresh();
                }
            });
        }
        
        // 时间范围选择
        const timeBtns = document.querySelectorAll('.time-btn');
        timeBtns.forEach(btn => {
            btn.addEventListener('click', (e) => {
                const range = e.currentTarget.getAttribute('data-range');
                this.setTimeRange(range);
            });
        });
    }
    
    // 设置时间范围
    setTimeRange(range) {
        this.currentTimeRange = range;
        
        // 更新按钮状态
        document.querySelectorAll('.time-btn').forEach(btn => {
            btn.classList.remove('active');
        });
        document.querySelector(`[data-range="${range}"]`).classList.add('active');
        
        // 重新加载图表数据
        this.loadChartsData();
    }
    
    // 开始自动刷新
    startAutoRefresh() {
        this.stopAutoRefresh();
        this.refreshInterval = setInterval(() => {
            this.loadData();
        }, this.refreshRate);
    }
    
    // 停止自动刷新
    stopAutoRefresh() {
        if (this.refreshInterval) {
            clearInterval(this.refreshInterval);
            this.refreshInterval = null;
        }
    }
    
    // 加载数据
    async loadData() {
        await Promise.all([
            this.loadSystemStats(),
            this.loadServerStats(),
            this.loadChartsData()
        ]);
    }
    
    // 加载系统统计
    async loadSystemStats() {
        try {
            const response = await fetch('/api/monitoring/system');
            const result = await response.json();

            if (response.ok && result.success) {
                this.systemStats = result.data;
                this.renderSystemStats();
            } else {
                const errorMessage = result.message || result.error || '加载系统信息失败';
                this.showSystemError(errorMessage);
            }
        } catch (error) {
            console.error('Failed to load system stats:', error);
            this.showSystemError('网络错误，请重试');
        }
    }
    
    // 加载服务器统计
    async loadServerStats() {
        try {
            const response = await fetch('/api/monitoring/servers');
            const result = await response.json();

            if (response.ok && result.success) {
                const servers = result.data || [];
                this.serverStats.clear();
                servers.forEach(server => {
                    this.serverStats.set(server.server_id, server);
                });
                this.renderServerStats();
            } else {
                const errorMessage = result.message || result.error || '加载服务器状态失败';
                this.showServersError(errorMessage);
            }
        } catch (error) {
            console.error('Failed to load server stats:', error);
            this.showServersError('网络错误，请重试');
        }
    }
    
    // 加载图表数据
    async loadChartsData() {
        try {
            const response = await fetch(`/api/monitoring/historical?type=system&range=${this.currentTimeRange}`);
            const result = await response.json();

            if (response.ok && result.success) {
                const chartsData = result.data || [];
                this.updateCharts(chartsData);
            } else {
                // 如果历史数据API不存在，使用当前系统数据
                this.updateCharts({});
            }
        } catch (error) {
            console.error('Failed to load charts data:', error);
            // 使用当前系统数据作为后备
            this.updateCharts({});
        }
    }
    
    // 渲染系统统计
    renderSystemStats() {
        const systemOverview = document.getElementById('systemOverview');
        if (!systemOverview || !this.systemStats) return;

        // 计算使用率
        const cpuUsage = this.systemStats.cpu?.usage || 0;
        const memoryUsage = this.systemStats.memory ?
            (this.systemStats.memory.used / this.systemStats.memory.total * 100) : 0;
        const diskUsage = this.systemStats.disk ?
            (this.systemStats.disk.used / this.systemStats.disk.total * 100) : 0;

        systemOverview.innerHTML = `
            <div class="stat-card cpu">
                <div class="stat-icon">
                    <i class="mdi mdi-cpu-64-bit"></i>
                </div>
                <div class="stat-content">
                    <div class="stat-label">CPU使用率</div>
                    <div class="stat-value">${cpuUsage.toFixed(1)}%</div>
                    <div class="stat-detail">${this.systemStats.cpu?.cores || 0} 核心</div>
                </div>
                <div class="stat-progress">
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${cpuUsage}%"></div>
                    </div>
                </div>
            </div>

            <div class="stat-card memory">
                <div class="stat-icon">
                    <i class="mdi mdi-memory"></i>
                </div>
                <div class="stat-content">
                    <div class="stat-label">内存使用率</div>
                    <div class="stat-value">${memoryUsage.toFixed(1)}%</div>
                    <div class="stat-detail">${this.formatBytes(this.systemStats.memory?.used || 0)} / ${this.formatBytes(this.systemStats.memory?.total || 0)}</div>
                </div>
                <div class="stat-progress">
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${memoryUsage}%"></div>
                    </div>
                </div>
            </div>

            <div class="stat-card disk">
                <div class="stat-icon">
                    <i class="mdi mdi-harddisk"></i>
                </div>
                <div class="stat-content">
                    <div class="stat-label">磁盘使用率</div>
                    <div class="stat-value">${diskUsage.toFixed(1)}%</div>
                    <div class="stat-detail">${this.formatBytes(this.systemStats.disk?.used || 0)} / ${this.formatBytes(this.systemStats.disk?.total || 0)}</div>
                </div>
                <div class="stat-progress">
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: ${diskUsage}%"></div>
                    </div>
                </div>
            </div>

            <div class="stat-card network">
                <div class="stat-icon">
                    <i class="mdi mdi-network"></i>
                </div>
                <div class="stat-content">
                    <div class="stat-label">网络流量</div>
                    <div class="stat-value">${this.formatBytes((this.systemStats.network?.bytes_recv || 0) + (this.systemStats.network?.bytes_sent || 0))}</div>
                    <div class="stat-detail">↓ ${this.formatBytes(this.systemStats.network?.bytes_recv || 0)} ↑ ${this.formatBytes(this.systemStats.network?.bytes_sent || 0)}</div>
                </div>
            </div>

            <div class="stat-card uptime">
                <div class="stat-icon">
                    <i class="mdi mdi-clock"></i>
                </div>
                <div class="stat-content">
                    <div class="stat-label">系统运行时间</div>
                    <div class="stat-value">${this.formatUptime(this.systemStats.uptime || 0)}</div>
                    <div class="stat-detail">CPU核心: ${this.systemStats.cpu?.cores || 0}</div>
                </div>
            </div>

            <div class="stat-card processes">
                <div class="stat-icon">
                    <i class="mdi mdi-application"></i>
                </div>
                <div class="stat-content">
                    <div class="stat-label">服务器状态</div>
                    <div class="stat-value">${this.serverStats.size}</div>
                    <div class="stat-detail">运行中的服务器数量</div>
                </div>
            </div>
        `;
    }
    
    // 渲染服务器统计
    renderServerStats() {
        const serversMonitoring = document.getElementById('serversMonitoring');
        if (!serversMonitoring) return;
        
        const servers = Array.from(this.serverStats.values());
        
        if (servers.length === 0) {
            serversMonitoring.innerHTML = `
                <div class="empty-state">
                    <i class="mdi mdi-server-off"></i>
                    <h4>没有服务器</h4>
                    <p>当前没有运行的服务器</p>
                </div>
            `;
            return;
        }
        
        serversMonitoring.innerHTML = servers.map(server => `
            <div class="server-monitor-card">
                <div class="server-header">
                    <div class="server-info">
                        <h4>${server.name}</h4>
                        <span class="server-status ${server.status}">${this.getStatusText(server.status)}</span>
                    </div>
                    <div class="server-actions">
                        <button class="btn-icon" onclick="window.getMonitoringPageManager().showServerDetails('${server.id}')" title="详情">
                            <i class="mdi mdi-information"></i>
                        </button>
                    </div>
                </div>
                
                <div class="server-stats">
                    <div class="server-stat">
                        <span class="stat-label">CPU</span>
                        <span class="stat-value">${server.cpu_usage?.toFixed(1) || 0}%</span>
                        <div class="mini-progress">
                            <div class="mini-progress-fill" style="width: ${server.cpu_usage || 0}%"></div>
                        </div>
                    </div>
                    
                    <div class="server-stat">
                        <span class="stat-label">内存</span>
                        <span class="stat-value">${server.memory_usage?.toFixed(1) || 0}%</span>
                        <div class="mini-progress">
                            <div class="mini-progress-fill" style="width: ${server.memory_usage || 0}%"></div>
                        </div>
                    </div>
                    
                    <div class="server-stat">
                        <span class="stat-label">在线玩家</span>
                        <span class="stat-value">${server.online_players || 0}/${server.max_players || 0}</span>
                    </div>
                    
                    <div class="server-stat">
                        <span class="stat-label">运行时间</span>
                        <span class="stat-value">${this.formatUptime(server.uptime || 0)}</span>
                    </div>
                </div>
            </div>
        `).join('');
    }
    
    // 更新图表
    updateCharts(chartsData) {
        // 如果有系统统计数据，使用它来更新图表值
        if (this.systemStats) {
            const cpuUsage = this.systemStats.cpu?.usage || 0;
            const memoryUsage = this.systemStats.memory ?
                (this.systemStats.memory.used / this.systemStats.memory.total * 100) : 0;
            const diskUsage = this.systemStats.disk ?
                (this.systemStats.disk.used / this.systemStats.disk.total * 100) : 0;
            const networkTotal = (this.systemStats.network?.bytes_recv || 0) + (this.systemStats.network?.bytes_sent || 0);

            const cpuValue = document.getElementById('cpuValue');
            const memoryValue = document.getElementById('memoryValue');
            const diskValue = document.getElementById('diskValue');
            const networkValue = document.getElementById('networkValue');

            if (cpuValue) cpuValue.textContent = `${cpuUsage.toFixed(1)}%`;
            if (memoryValue) memoryValue.textContent = `${memoryUsage.toFixed(1)}%`;
            if (diskValue) diskValue.textContent = `${diskUsage.toFixed(1)}%`;
            if (networkValue) networkValue.textContent = `${this.formatBytes(networkTotal)}`;

            // 绘制简单的图表
            this.drawSimpleChart('cpuChart', cpuUsage, '% CPU');
            this.drawSimpleChart('memoryChart', memoryUsage, '% Memory');
            this.drawSimpleChart('diskChart', diskUsage, '% Disk');
            this.drawNetworkChart('networkChart', this.systemStats.network);
        }
    }

    // 绘制简单的图表
    drawSimpleChart(canvasId, value, label) {
        const canvas = document.getElementById(canvasId);
        if (!canvas) return;

        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;

        // 清除画布
        ctx.clearRect(0, 0, width, height);

        // 设置样式
        ctx.fillStyle = '#f0f0f0';
        ctx.fillRect(0, 0, width, height);

        // 绘制进度条
        const barHeight = 20;
        const barY = height / 2 - barHeight / 2;
        const barWidth = width - 40;
        const barX = 20;

        // 背景
        ctx.fillStyle = '#e0e0e0';
        ctx.fillRect(barX, barY, barWidth, barHeight);

        // 进度
        const progressWidth = (barWidth * value) / 100;
        ctx.fillStyle = value > 80 ? '#ff4444' : value > 60 ? '#ffaa00' : '#44aa44';
        ctx.fillRect(barX, barY, progressWidth, barHeight);

        // 文字
        ctx.fillStyle = '#333';
        ctx.font = '14px Arial';
        ctx.textAlign = 'center';
        ctx.fillText(`${value.toFixed(1)}${label}`, width / 2, barY - 10);
    }

    // 绘制网络图表
    drawNetworkChart(canvasId, networkData) {
        const canvas = document.getElementById(canvasId);
        if (!canvas || !networkData) return;

        const ctx = canvas.getContext('2d');
        const width = canvas.width;
        const height = canvas.height;

        // 清除画布
        ctx.clearRect(0, 0, width, height);

        // 设置样式
        ctx.fillStyle = '#f0f0f0';
        ctx.fillRect(0, 0, width, height);

        // 绘制上传和下载
        const recv = networkData.bytes_recv || 0;
        const sent = networkData.bytes_sent || 0;

        ctx.fillStyle = '#333';
        ctx.font = '12px Arial';
        ctx.textAlign = 'left';
        ctx.fillText(`↓ ${this.formatBytes(recv)}`, 10, height / 2 - 10);
        ctx.fillText(`↑ ${this.formatBytes(sent)}`, 10, height / 2 + 20);
    }
    
    // 格式化字节数
    formatBytes(bytes) {
        if (bytes === 0) return '0 B';
        
        const k = 1024;
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        
        return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
    }
    
    // 格式化运行时间
    formatUptime(seconds) {
        const days = Math.floor(seconds / 86400);
        const hours = Math.floor((seconds % 86400) / 3600);
        const minutes = Math.floor((seconds % 3600) / 60);
        
        if (days > 0) {
            return `${days}天 ${hours}小时`;
        } else if (hours > 0) {
            return `${hours}小时 ${minutes}分钟`;
        } else {
            return `${minutes}分钟`;
        }
    }
    
    // 获取状态文本
    getStatusText(status) {
        const statusTexts = {
            'running': '运行中',
            'stopped': '已停止',
            'starting': '启动中',
            'stopping': '停止中',
            'error': '错误'
        };
        return statusTexts[status] || status;
    }
    
    // 显示系统错误
    showSystemError(message) {
        const systemOverview = document.getElementById('systemOverview');
        if (systemOverview) {
            systemOverview.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h4>加载失败</h4>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getMonitoringPageManager().loadSystemStats()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 显示服务器错误
    showServersError(message) {
        const serversMonitoring = document.getElementById('serversMonitoring');
        if (serversMonitoring) {
            serversMonitoring.innerHTML = `
                <div class="error-state">
                    <i class="mdi mdi-alert-circle"></i>
                    <h4>加载失败</h4>
                    <p>${message}</p>
                    <button class="btn primary" onclick="window.getMonitoringPageManager().loadServerStats()">
                        <i class="mdi mdi-refresh"></i>
                        <span>重试</span>
                    </button>
                </div>
            `;
        }
    }
    
    // 显示服务器详情
    showServerDetails(serverId) {
        const server = this.serverStats.get(serverId);
        if (!server) return;
        
        const uiManager = window.getUIManager();
        
        const content = `
            <div class="server-details">
                <div class="detail-header">
                    <h3>${server.name}</h3>
                    <span class="status-badge ${server.status}">${this.getStatusText(server.status)}</span>
                </div>
                
                <div class="detail-grid">
                    <div class="detail-section">
                        <h4>基本信息</h4>
                        <div class="detail-items">
                            <div class="detail-item">
                                <span class="label">版本</span>
                                <span class="value">${server.version || 'Unknown'}</span>
                            </div>
                            <div class="detail-item">
                                <span class="label">端口</span>
                                <span class="value">${server.port || 25565}</span>
                            </div>
                            <div class="detail-item">
                                <span class="label">运行时间</span>
                                <span class="value">${this.formatUptime(server.uptime || 0)}</span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="detail-section">
                        <h4>性能指标</h4>
                        <div class="detail-items">
                            <div class="detail-item">
                                <span class="label">CPU使用率</span>
                                <span class="value">${server.cpu_usage?.toFixed(1) || 0}%</span>
                            </div>
                            <div class="detail-item">
                                <span class="label">内存使用率</span>
                                <span class="value">${server.memory_usage?.toFixed(1) || 0}%</span>
                            </div>
                            <div class="detail-item">
                                <span class="label">内存使用量</span>
                                <span class="value">${this.formatBytes(server.memory_used || 0)} / ${this.formatBytes(server.memory_total || 0)}</span>
                            </div>
                        </div>
                    </div>
                    
                    <div class="detail-section">
                        <h4>玩家信息</h4>
                        <div class="detail-items">
                            <div class="detail-item">
                                <span class="label">在线玩家</span>
                                <span class="value">${server.online_players || 0}/${server.max_players || 0}</span>
                            </div>
                            <div class="detail-item">
                                <span class="label">今日峰值</span>
                                <span class="value">${server.peak_players || 0}</span>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        
        uiManager?.showModal('服务器详情', content, {
            width: '600px',
            footer: `
                <button class="btn" onclick="window.getUIManager().closeModal()">
                    <i class="mdi mdi-close"></i>
                    <span>关闭</span>
                </button>
            `
        });
    }
}

// 全局监控页面管理器实例
let monitoringPageManager;

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', () => {
    monitoringPageManager = new MonitoringPageManager();
});

// 导出到全局作用域
window.MonitoringPageManager = MonitoringPageManager;
window.getMonitoringPageManager = () => monitoringPageManager;
