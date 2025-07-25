package frp

import "time"

// APIResponse OpenFRP API统一响应格式
type APIResponse struct {
	Flag   bool        `json:"flag"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data"`
	Detail interface{} `json:"detail"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID            int    `json:"id"`
	Username      string `json:"username"`
	Email         string `json:"email"`
	Token         string `json:"token"`
	Realname      bool   `json:"realname"`
	RegTime       string `json:"regTime"`  // 修改为string类型，因为API返回的是时间字符串
	Group         string `json:"group"`
	FriendlyGroup string `json:"friendlyGroup"`
	InLimit       int    `json:"inLimit"`
	OutLimit      int    `json:"outLimit"`
	Used          int    `json:"used"`
	Proxies       int    `json:"proxies"`
	Traffic       int    `json:"traffic"`
}

// ProxyInfo 隧道信息
type ProxyInfo struct {
	ID                     int    `json:"id"`
	UID                    int    `json:"uid"`
	ProxyName              string `json:"proxyName"`
	ProxyType              string `json:"proxyType"`
	LocalIP                string `json:"localIp"`
	LocalPort              int    `json:"localPort"`
	RemotePort             int    `json:"remotePort"`
	Domain                 string `json:"domain"`
	NID                    int    `json:"nid"`
	FriendlyNode           string `json:"friendlyNode"`
	ConnectAddress         string `json:"connectAddress"`
	Status                 bool   `json:"status"`
	Online                 bool   `json:"online"`
	UseEncryption          bool   `json:"useEncryption"`
	UseCompression         bool   `json:"useCompression"`
	ProxyProtocolVersion   bool   `json:"proxyProtocolVersion"`
	AutoTLS                string `json:"autoTls"`
	ForceHTTPS             bool   `json:"forceHttps"`
	Custom                 string `json:"custom"`
	LastUpdate             int64  `json:"lastUpdate"`
	LastLogin              *int64 `json:"lastLogin"`
}

// ProxyListResponse 隧道列表响应
type ProxyListResponse struct {
	Total int         `json:"total"`
	List  []ProxyInfo `json:"list"`
}

// NodeInfo 节点信息
type NodeInfo struct {
	ID                      int                    `json:"id"`
	Name                    string                 `json:"name"`
	Hostname                string                 `json:"hostname"`
	Port                    int                    `json:"port"`
	Status                  int                    `json:"status"`
	Classify                int                    `json:"classify"`
	Group                   []string               `json:"group"`
	Bandwidth               int                    `json:"bandwidth"`
	BandwidthMagnification  float64                `json:"bandwidthMagnification"`
	MaxOnlineMagnification  float64                `json:"maxOnlineMagnification"`
	Comments                string                 `json:"comments"`
	Description             string                 `json:"description"`
	NeedRealname            bool                   `json:"needRealname"`
	EnableDefaultTLS        bool                   `json:"enableDefaultTls"`
	AllowEC                 bool                   `json:"allowEc"`
	UnitcostEC              float64                `json:"unitcostEc"`
	AllowPort               string                 `json:"allowPort"`
	FullyLoaded             bool                   `json:"fullyLoaded"`
	ProtocolSupport         map[string]bool        `json:"protocolSupport"`
}

// NodeListResponse 节点列表响应
type NodeListResponse struct {
	Total int        `json:"total"`
	List  []NodeInfo `json:"list"`
}

// CreateProxyRequest 创建隧道请求
type CreateProxyRequest struct {
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	LocalAddr              string `json:"local_addr"`
	LocalPort              string `json:"local_port"`
	RemotePort             int    `json:"remote_port,omitempty"`
	NodeID                 int    `json:"node_id"`
	DomainBind             string `json:"domain_bind,omitempty"`
	DataEncrypt            bool   `json:"dataEncrypt"`
	DataGzip               bool   `json:"dataGzip"`
	ProxyProtocolVersion   bool   `json:"proxyProtocolVersion"`
	AutoTLS                string `json:"autoTls,omitempty"`
	ForceHTTPS             bool   `json:"forceHttps"`
	Custom                 string `json:"custom,omitempty"`
}

// EditProxyRequest 编辑隧道请求
type EditProxyRequest struct {
	ProxyID                int    `json:"proxy_id"`
	Name                   string `json:"name"`
	Type                   string `json:"type"`
	LocalAddr              string `json:"local_addr"`
	LocalPort              string `json:"local_port"`
	RemotePort             int    `json:"remote_port,omitempty"`
	NodeID                 int    `json:"node_id"`
	DomainBind             string `json:"domain_bind,omitempty"`
	DataEncrypt            bool   `json:"dataEncrypt"`
	DataGzip               bool   `json:"dataGzip"`
	ProxyProtocolVersion   bool   `json:"proxyProtocolVersion"`
	AutoTLS                string `json:"autoTls,omitempty"`
	ForceHTTPS             bool   `json:"forceHttps"`
	Custom                 string `json:"custom,omitempty"`
}

// TokenProxyNode Token API节点信息
type TokenProxyNode struct {
	Node    string            `json:"node"`
	Proxies []TokenProxyInfo  `json:"proxies"`
}

// TokenProxyInfo Token API隧道信息
type TokenProxyInfo struct {
	Name   string `json:"name"`
	ID     int    `json:"id"`
	Type   string `json:"type"`
	Remote string `json:"remote"`
	Local  string `json:"local"`
}

// TokenProxyResponse Token API响应
type TokenProxyResponse struct {
	Status  int              `json:"status"`
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    []TokenProxyNode `json:"data"`
}

// ProxyTemplate 隧道模板
type ProxyTemplate struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"`
	LocalPort   int               `json:"local_port"`
	Config      map[string]string `json:"config"`
}

// ProxyStatus 隧道状态
type ProxyStatus struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Online      bool      `json:"online"`
	LastSeen    time.Time `json:"last_seen"`
	ConnectAddr string    `json:"connect_addr"`
	Traffic     struct {
		In  int64 `json:"in"`
		Out int64 `json:"out"`
	} `json:"traffic"`
}

// FRPCConfig frpc配置文件结构
type FRPCConfig struct {
	Common   FRPCCommonConfig            `ini:"common"`
	Proxies  map[string]FRPCProxyConfig  `ini:",extends"`
}

// FRPCCommonConfig frpc通用配置
type FRPCCommonConfig struct {
	ServerAddr      string `ini:"server_addr"`
	ServerPort      int    `ini:"server_port"`
	Token           string `ini:"token"`
	User            string `ini:"user"`
	LogLevel        string `ini:"log_level"`
	LogFile         string `ini:"log_file"`
	LogMaxDays      int    `ini:"log_max_days"`
	AdminAddr       string `ini:"admin_addr"`
	AdminPort       int    `ini:"admin_port"`
	AdminUser       string `ini:"admin_user"`
	AdminPwd        string `ini:"admin_pwd"`
	PoolCount       int    `ini:"pool_count"`
	TCPMux          bool   `ini:"tcp_mux"`
	HeartbeatInterval int  `ini:"heartbeat_interval"`
	HeartbeatTimeout  int  `ini:"heartbeat_timeout"`
}

// FRPCProxyConfig frpc隧道配置
type FRPCProxyConfig struct {
	Type               string `ini:"type"`
	LocalIP            string `ini:"local_ip"`
	LocalPort          int    `ini:"local_port"`
	RemotePort         int    `ini:"remote_port,omitempty"`
	CustomDomains      string `ini:"custom_domains,omitempty"`
	SubDomain          string `ini:"subdomain,omitempty"`
	UseEncryption      bool   `ini:"use_encryption"`
	UseCompression     bool   `ini:"use_compression"`
	ProxyProtocolVersion string `ini:"proxy_protocol_version,omitempty"`
	HealthCheckType    string `ini:"health_check_type,omitempty"`
	HealthCheckURL     string `ini:"health_check_url,omitempty"`
	HealthCheckIntervalS int  `ini:"health_check_interval_s,omitempty"`
}

// ProxyTypeInfo 隧道类型信息
type ProxyTypeInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	RequirePort bool     `json:"require_port"`
	SupportDomain bool   `json:"support_domain"`
	DefaultPort int      `json:"default_port"`
	Examples    []string `json:"examples"`
}

// GetProxyTypes 获取支持的隧道类型
func GetProxyTypes() []ProxyTypeInfo {
	return []ProxyTypeInfo{
		{
			Name:        "tcp",
			DisplayName: "TCP",
			Description: "TCP协议隧道，适用于大多数应用",
			RequirePort: true,
			SupportDomain: false,
			DefaultPort: 25565,
			Examples:    []string{"Minecraft服务器", "SSH", "数据库"},
		},
		{
			Name:        "udp",
			DisplayName: "UDP",
			Description: "UDP协议隧道，适用于游戏和实时应用",
			RequirePort: true,
			SupportDomain: false,
			DefaultPort: 25565,
			Examples:    []string{"Minecraft基岩版", "游戏服务器", "DNS"},
		},
		{
			Name:        "http",
			DisplayName: "HTTP",
			Description: "HTTP协议隧道，适用于网站",
			RequirePort: false,
			SupportDomain: true,
			DefaultPort: 80,
			Examples:    []string{"网站", "Web应用", "API服务"},
		},
		{
			Name:        "https",
			DisplayName: "HTTPS",
			Description: "HTTPS协议隧道，适用于安全网站",
			RequirePort: false,
			SupportDomain: true,
			DefaultPort: 443,
			Examples:    []string{"安全网站", "HTTPS API", "Web应用"},
		},
	}
}
