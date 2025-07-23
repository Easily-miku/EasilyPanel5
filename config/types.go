package config

import "time"

// AppConfig 应用程序配置
type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Java     JavaConfig     `json:"java"`
	Download DownloadConfig `json:"download"`
	Logging  LoggingConfig  `json:"logging"`
	Daemon   DaemonConfig   `json:"daemon"`
	FRP      FRPConfig      `json:"frp"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int    `json:"port"`
	Host         string `json:"host"`
	StaticDir    string `json:"static_dir"`
	MaxServers   int    `json:"max_servers"`
	DefaultMemory int   `json:"default_memory"` // MB
}

// JavaConfig Java环境配置
type JavaConfig struct {
	AutoDetect   bool     `json:"auto_detect"`
	JavaPath     string   `json:"java_path"`
	DefaultArgs  []string `json:"default_args"`
	MinVersion   int      `json:"min_version"`
	MaxMemory    int      `json:"max_memory"` // MB
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	CoresDir     string        `json:"cores_dir"`
	Timeout      time.Duration `json:"timeout"`
	MaxRetries   int           `json:"max_retries"`
	ChunkSize    int64         `json:"chunk_size"`
	FastMirrorAPI string       `json:"fastmirror_api"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level        string `json:"level"`
	LogsDir      string `json:"logs_dir"`
	MaxFileSize  int64  `json:"max_file_size"`  // bytes
	MaxFiles     int    `json:"max_files"`
	MaxAge       int    `json:"max_age"`        // days
	EnableColors bool   `json:"enable_colors"`
}

// DaemonConfig 守护进程配置
type DaemonConfig struct {
	EnableAutoRestart   bool          `json:"enable_auto_restart"`   // 启用自动重启
	MaxRestartAttempts  int           `json:"max_restart_attempts"`  // 最大重启尝试次数
	RestartDelay        time.Duration `json:"restart_delay"`         // 重启延迟
	CrashDetectionDelay time.Duration `json:"crash_detection_delay"` // 崩溃检测延迟
	ResourceMonitoring  bool          `json:"resource_monitoring"`   // 启用资源监控
	MonitorInterval     time.Duration `json:"monitor_interval"`      // 监控间隔
	LogRotation         bool          `json:"log_rotation"`          // 启用日志轮转
	MaxLogSize          int64         `json:"max_log_size"`          // 最大日志文件大小
	MaxLogFiles         int           `json:"max_log_files"`         // 最大日志文件数量
}

// FRPConfig FRP配置
type FRPConfig struct {
	Enabled            bool          `json:"enabled"`             // 启用FRP
	APIEndpoint        string        `json:"api_endpoint"`        // API端点
	DefaultNode        string        `json:"default_node"`        // 默认节点
	AutoStart          bool          `json:"auto_start"`          // 自动启动隧道
	AutoRestart        bool          `json:"auto_restart"`        // 自动重启隧道
	MaxTunnels         int           `json:"max_tunnels"`         // 最大隧道数
	MonitorInterval    time.Duration `json:"monitor_interval"`    // 监控间隔
	StatsRetention     time.Duration `json:"stats_retention"`     // 统计数据保留时间
	DefaultBandwidth   int64         `json:"default_bandwidth"`   // 默认带宽限制
	DefaultCompression bool          `json:"default_compression"` // 默认启用压缩
	DefaultEncryption  bool          `json:"default_encryption"`  // 默认启用加密
}

// MinecraftServer Minecraft服务器配置
type MinecraftServer struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	CoreType    string            `json:"core_type"`    // paper, spigot, fabric等
	MCVersion   string            `json:"mc_version"`   // 1.20.1等
	CoreVersion string            `json:"core_version"` // build版本
	JavaPath    string            `json:"java_path"`
	Memory      int               `json:"memory"`       // MB
	Port        int               `json:"port"`
	Status      string            `json:"status"`       // stopped, running, crashed
	PID         int               `json:"pid"`
	WorkDir     string            `json:"work_dir"`
	JarFile     string            `json:"jar_file"`
	StartTime   *time.Time        `json:"start_time,omitempty"`
	StopTime    *time.Time        `json:"stop_time,omitempty"`
	JavaArgs    []string          `json:"java_args"`
	Properties  map[string]string `json:"properties"`
	AutoStart   bool              `json:"auto_start"`
	AutoRestart bool              `json:"auto_restart"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`

	// 守护进程相关字段
	RestartAttempts    int       `json:"restart_attempts"`     // 当前重启尝试次数
	LastCrashTime      *time.Time `json:"last_crash_time"`     // 最后崩溃时间
	CrashCount         int       `json:"crash_count"`          // 崩溃次数
	ResourceUsage      *ResourceUsage `json:"resource_usage"`  // 资源使用情况
	DaemonEnabled      bool      `json:"daemon_enabled"`       // 是否启用守护
	MaxRestartAttempts int       `json:"max_restart_attempts"` // 最大重启尝试次数

	// 实例管理增强字段
	Group              string            `json:"group"`                // 服务器分组
	Tags               []string          `json:"tags"`                 // 标签
	Description        string            `json:"description"`          // 描述
	TemplateID         string            `json:"template_id"`          // 模板ID
	BackupEnabled      bool              `json:"backup_enabled"`       // 启用备份
	BackupSchedule     string            `json:"backup_schedule"`      // 备份计划(cron格式)
	LastBackupTime     *time.Time        `json:"last_backup_time"`     // 最后备份时间
	MonitoringEnabled  bool              `json:"monitoring_enabled"`   // 启用监控
	AlertThresholds    *AlertThresholds  `json:"alert_thresholds"`     // 告警阈值
	CustomConfig       map[string]interface{} `json:"custom_config"`   // 自定义配置
}

// ResourceUsage 资源使用情况
type ResourceUsage struct {
	CPUPercent    float64   `json:"cpu_percent"`    // CPU使用率
	MemoryUsed    int64     `json:"memory_used"`    // 内存使用量（字节）
	MemoryPercent float64   `json:"memory_percent"` // 内存使用率
	DiskRead      int64     `json:"disk_read"`      // 磁盘读取量
	DiskWrite     int64     `json:"disk_write"`     // 磁盘写入量
	NetworkIn     int64     `json:"network_in"`     // 网络入流量
	NetworkOut    int64     `json:"network_out"`    // 网络出流量
	UpdatedAt     time.Time `json:"updated_at"`     // 更新时间
}

// AlertThresholds 告警阈值
type AlertThresholds struct {
	CPUPercent    float64 `json:"cpu_percent"`    // CPU使用率阈值
	MemoryPercent float64 `json:"memory_percent"` // 内存使用率阈值
	DiskPercent   float64 `json:"disk_percent"`   // 磁盘使用率阈值
	PlayerCount   int     `json:"player_count"`   // 玩家数量阈值
}

// ServerTemplate 服务器模板
type ServerTemplate struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CoreType    string            `json:"core_type"`
	MCVersion   string            `json:"mc_version"`
	Memory      int               `json:"memory"`
	JavaArgs    []string          `json:"java_args"`
	Properties  map[string]string `json:"properties"`
	Plugins     []string          `json:"plugins"`     // 插件列表
	Mods        []string          `json:"mods"`        // 模组列表
	Config      map[string]interface{} `json:"config"` // 其他配置
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	IsDefault   bool              `json:"is_default"`  // 是否为默认模板
}

// ServerGroup 服务器分组
type ServerGroup struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`       // 分组颜色
	Icon        string    `json:"icon"`        // 分组图标
	ServerCount int       `json:"server_count"` // 服务器数量
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BatchOperation 批量操作
type BatchOperation struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`      // start, stop, restart, delete
	ServerIDs []string               `json:"server_ids"`
	Status    string                 `json:"status"`    // pending, running, completed, failed
	Progress  int                    `json:"progress"`  // 进度百分比
	Results   map[string]interface{} `json:"results"`   // 操作结果
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	Error     string                 `json:"error,omitempty"`
}

// ServerStatus 服务器状态常量
const (
	StatusStopped = "stopped"
	StatusRunning = "running"
	StatusCrashed = "crashed"
	StatusStarting = "starting"
	StatusStopping = "stopping"
)

// CoreInfo 核心信息
type CoreInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	MCVersion   string `json:"mc_version"`
	CoreVersion string `json:"core_version"`
	FileName    string `json:"filename"`
	DownloadURL string `json:"download_url"`
	SHA1        string `json:"sha1"`
	UpdateTime  string `json:"update_time"`
	Recommend   bool   `json:"recommend"`
}

// DownloadTask 下载任务
type DownloadTask struct {
	ID          string    `json:"id"`
	CoreInfo    CoreInfo  `json:"core_info"`
	Status      string    `json:"status"`      // pending, downloading, completed, failed
	Progress    float64   `json:"progress"`    // 0-100
	Speed       int64     `json:"speed"`       // bytes/sec
	Downloaded  int64     `json:"downloaded"`  // bytes
	Total       int64     `json:"total"`       // bytes
	Error       string    `json:"error,omitempty"`
	StartTime   time.Time `json:"start_time"`
	CompleteTime *time.Time `json:"complete_time,omitempty"`
}

// JavaInfo Java环境信息
type JavaInfo struct {
	Path        string `json:"path"`
	Version     string `json:"version"`
	Vendor      string `json:"vendor"`
	Architecture string `json:"architecture"`
	IsValid     bool   `json:"is_valid"`
	Error       string `json:"error,omitempty"`
}
