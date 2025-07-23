package frp

import (
	"fmt"
	"time"
)

// TunnelConfig 隧道配置
type TunnelConfig struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`         // tcp, udp, http, https
	LocalIP      string            `json:"local_ip"`
	LocalPort    int               `json:"local_port"`
	RemotePort   int               `json:"remote_port"`  // 仅TCP/UDP
	CustomDomain string            `json:"custom_domain"` // 仅HTTP/HTTPS
	ServerID     string            `json:"server_id"`    // 关联的Minecraft服务器ID
	Status       string            `json:"status"`       // stopped, starting, running, error
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	
	// OpenFRP特定配置
	Token        string            `json:"token"`        // OpenFRP令牌
	Node         string            `json:"node"`         // 节点名称
	Bandwidth    int64             `json:"bandwidth"`    // 带宽限制(KB/s)
	Compression  bool              `json:"compression"`  // 启用压缩
	Encryption   bool              `json:"encryption"`   // 启用加密
	
	// 统计信息
	Stats        *TunnelStats      `json:"stats,omitempty"`
}

// TunnelStats 隧道统计信息
type TunnelStats struct {
	BytesIn      int64     `json:"bytes_in"`      // 入流量
	BytesOut     int64     `json:"bytes_out"`     // 出流量
	Connections  int       `json:"connections"`   // 当前连接数
	LastActivity time.Time `json:"last_activity"` // 最后活动时间
	Uptime       time.Duration `json:"uptime"`    // 运行时间
	UpdatedAt    time.Time `json:"updated_at"`
}

// OpenFRPNode OpenFRP节点信息
type OpenFRPNode struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Location    string  `json:"location"`
	Load        float64 `json:"load"`        // 负载百分比
	Online      bool    `json:"online"`
	MaxBandwidth int64  `json:"max_bandwidth"` // 最大带宽
	Description string  `json:"description"`
}

// OpenFRPUser OpenFRP用户信息
type OpenFRPUser struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Token       string    `json:"token"`
	Bandwidth   int64     `json:"bandwidth"`   // 可用带宽
	Traffic     int64     `json:"traffic"`     // 已用流量
	Tunnels     int       `json:"tunnels"`     // 隧道数量
	ExpiresAt   time.Time `json:"expires_at"`  // 到期时间
}

// TunnelRequest 创建隧道请求
type TunnelRequest struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	LocalIP      string `json:"local_ip"`
	LocalPort    int    `json:"local_port"`
	RemotePort   int    `json:"remote_port,omitempty"`
	CustomDomain string `json:"custom_domain,omitempty"`
	ServerID     string `json:"server_id,omitempty"`
	Token        string `json:"token"`
	Node         string `json:"node"`
	Bandwidth    int64  `json:"bandwidth,omitempty"`
	Compression  bool   `json:"compression,omitempty"`
	Encryption   bool   `json:"encryption,omitempty"`
}

// TunnelStatus 隧道状态常量
const (
	TunnelStatusStopped  = "stopped"
	TunnelStatusStarting = "starting"
	TunnelStatusRunning  = "running"
	TunnelStatusError    = "error"
)

// TunnelType 隧道类型常量
const (
	TunnelTypeTCP   = "tcp"
	TunnelTypeUDP   = "udp"
	TunnelTypeHTTP  = "http"
	TunnelTypeHTTPS = "https"
)

// OpenFRPConfig OpenFRP配置
type OpenFRPConfig struct {
	Enabled     bool   `json:"enabled"`
	APIEndpoint string `json:"api_endpoint"`
	DefaultNode string `json:"default_node"`
	AutoStart   bool   `json:"auto_start"`     // 自动启动隧道
	AutoRestart bool   `json:"auto_restart"`   // 自动重启隧道
	MaxTunnels  int    `json:"max_tunnels"`    // 最大隧道数
	
	// 监控配置
	MonitorInterval time.Duration `json:"monitor_interval"` // 监控间隔
	StatsRetention  time.Duration `json:"stats_retention"`  // 统计数据保留时间
	
	// 默认设置
	DefaultBandwidth int64 `json:"default_bandwidth"` // 默认带宽限制
	DefaultCompression bool `json:"default_compression"` // 默认启用压缩
	DefaultEncryption  bool `json:"default_encryption"`  // 默认启用加密
}

// FRPManager FRP管理器接口
type FRPManager interface {
	// 隧道管理
	CreateTunnel(req *TunnelRequest) (*TunnelConfig, error)
	DeleteTunnel(tunnelID string) error
	StartTunnel(tunnelID string) error
	StopTunnel(tunnelID string) error
	RestartTunnel(tunnelID string) error
	
	// 隧道查询
	GetTunnel(tunnelID string) (*TunnelConfig, error)
	GetAllTunnels() ([]*TunnelConfig, error)
	GetTunnelsByServer(serverID string) ([]*TunnelConfig, error)
	
	// 统计信息
	GetTunnelStats(tunnelID string) (*TunnelStats, error)
	GetAllStats() (map[string]*TunnelStats, error)
	
	// OpenFRP API
	GetNodes() ([]*OpenFRPNode, error)
	GetUserInfo(token string) (*OpenFRPUser, error)
	ValidateToken(token string) error
	
	// 管理器控制
	Start() error
	Stop() error
	IsRunning() bool
}

// TunnelEvent 隧道事件
type TunnelEvent struct {
	Type     string      `json:"type"`
	TunnelID string      `json:"tunnel_id"`
	Data     interface{} `json:"data"`
	Error    string      `json:"error,omitempty"`
	Time     time.Time   `json:"time"`
}

// 事件类型常量
const (
	EventTunnelCreated    = "tunnel_created"
	EventTunnelDeleted    = "tunnel_deleted"
	EventTunnelStarted    = "tunnel_started"
	EventTunnelStopped    = "tunnel_stopped"
	EventTunnelError      = "tunnel_error"
	EventTunnelStatsUpdate = "tunnel_stats_update"
	EventNodeStatusChange = "node_status_change"
)

// OpenFRPError OpenFRP错误类型
type OpenFRPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *OpenFRPError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("OpenFRP Error %d: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("OpenFRP Error %d: %s", e.Code, e.Message)
}

// 常见错误代码
const (
	ErrCodeInvalidToken     = 1001
	ErrCodeInsufficientQuota = 1002
	ErrCodeNodeUnavailable  = 1003
	ErrCodePortInUse        = 1004
	ErrCodeTunnelNotFound   = 1005
	ErrCodeAPIError         = 1006
)
