package config

import "time"

// AppConfig 应用程序配置
type AppConfig struct {
	Server   ServerConfig   `json:"server"`
	Java     JavaConfig     `json:"java"`
	Download DownloadConfig `json:"download"`
	Logging  LoggingConfig  `json:"logging"`
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
