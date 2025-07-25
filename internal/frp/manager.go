package frp

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Manager FRP管理器，整合API客户端和frpc管理
type Manager struct {
	client     *OpenFRPClient
	frpc       *FRPCManager
	dataDir    string
	authorized bool
}

// NewManager 创建新的FRP管理器
func NewManager(dataDir string) *Manager {
	// 设置默认路径
	binaryPath := filepath.Join(dataDir, "bin", "frpc")
	if os.Getenv("GOOS") == "windows" {
		binaryPath += ".exe"
	}
	
	configPath := filepath.Join(dataDir, "configs", "frpc.ini")
	logPath := filepath.Join(dataDir, "logs", "frpc.log")
	
	return &Manager{
		client:  NewOpenFRPClient("", ""),
		frpc:    NewFRPCManager(binaryPath, configPath, logPath),
		dataDir: dataDir,
	}
}

// SetAuthorization 设置认证令牌
func (m *Manager) SetAuthorization(auth string) {
	m.client.SetAuthorization(auth)
	m.authorized = auth != ""
}

// IsAuthorized 检查是否已认证
func (m *Manager) IsAuthorized() bool {
	return m.authorized
}

// TestConnection 测试连接
func (m *Manager) TestConnection() error {
	if !m.authorized {
		return fmt.Errorf("未设置认证令牌")
	}
	
	return m.client.TestConnection()
}

// GetUserInfo 获取用户信息
func (m *Manager) GetUserInfo() (*UserInfo, error) {
	if !m.authorized {
		return nil, fmt.Errorf("未设置认证令牌")
	}
	
	return m.client.GetUserInfo()
}

// GetProxies 获取隧道列表
func (m *Manager) GetProxies() ([]ProxyInfo, error) {
	if !m.authorized {
		return nil, fmt.Errorf("未设置认证令牌")
	}
	
	resp, err := m.client.GetProxies()
	if err != nil {
		return nil, err
	}
	
	return resp.List, nil
}

// GetNodes 获取节点列表
func (m *Manager) GetNodes() ([]NodeInfo, error) {
	if !m.authorized {
		return nil, fmt.Errorf("未设置认证令牌")
	}
	
	resp, err := m.client.GetNodes()
	if err != nil {
		return nil, err
	}
	
	return resp.List, nil
}

// GetAvailableNodes 获取可用节点列表
func (m *Manager) GetAvailableNodes() ([]NodeInfo, error) {
	nodes, err := m.GetNodes()
	if err != nil {
		return nil, err
	}
	
	var available []NodeInfo
	for _, node := range nodes {
		// 过滤可用节点
		if node.Status == 200 && !node.FullyLoaded {
			available = append(available, node)
		}
	}
	
	return available, nil
}

// CreateProxy 创建隧道
func (m *Manager) CreateProxy(req *CreateProxyRequest) error {
	if !m.authorized {
		return fmt.Errorf("未设置认证令牌")
	}
	
	return m.client.CreateProxy(req)
}

// EditProxy 编辑隧道
func (m *Manager) EditProxy(req *EditProxyRequest) error {
	if !m.authorized {
		return fmt.Errorf("未设置认证令牌")
	}
	
	return m.client.EditProxy(req)
}

// DeleteProxy 删除隧道
func (m *Manager) DeleteProxy(proxyID int) error {
	if !m.authorized {
		return fmt.Errorf("未设置认证令牌")
	}
	
	return m.client.DeleteProxy(proxyID)
}

// SetupFRPC 设置frpc客户端
func (m *Manager) SetupFRPC() error {
	// 检查是否已安装
	if !m.frpc.IsInstalled() {
		fmt.Println("frpc未安装，正在下载...")
		if err := m.frpc.DownloadFRPC(); err != nil {
			return fmt.Errorf("下载frpc失败: %w", err)
		}
	}
	
	// 检查版本
	version, err := m.frpc.GetVersion()
	if err != nil {
		return fmt.Errorf("获取frpc版本失败: %w", err)
	}
	
	fmt.Printf("frpc版本: %s\n", version)
	return nil
}

// GenerateConfig 生成frpc配置
func (m *Manager) GenerateConfig(serverAddr, token string) error {
	// 获取隧道列表
	proxies, err := m.GetProxies()
	if err != nil {
		return fmt.Errorf("获取隧道列表失败: %w", err)
	}
	
	// 过滤启用的隧道
	var enabledProxies []ProxyInfo
	for _, proxy := range proxies {
		if proxy.Status {
			enabledProxies = append(enabledProxies, proxy)
		}
	}
	
	return m.frpc.GenerateConfig(serverAddr, token, enabledProxies)
}

// StartFRPC 启动frpc
func (m *Manager) StartFRPC() error {
	return m.frpc.Start()
}

// StopFRPC 停止frpc
func (m *Manager) StopFRPC() error {
	return m.frpc.Stop()
}

// RestartFRPC 重启frpc
func (m *Manager) RestartFRPC() error {
	return m.frpc.Restart()
}

// GetFRPCStatus 获取frpc状态
func (m *Manager) GetFRPCStatus() string {
	return m.frpc.GetStatus()
}

// IsFRPCRunning 检查frpc是否运行
func (m *Manager) IsFRPCRunning() bool {
	return m.frpc.IsRunning()
}

// GetFRPCLogs 获取frpc日志
func (m *Manager) GetFRPCLogs(lines int) ([]string, error) {
	return m.frpc.GetLogs(lines)
}

// ClearFRPCLogs 清空frpc日志
func (m *Manager) ClearFRPCLogs() error {
	return m.frpc.ClearLogs()
}

// PrintProxyList 打印隧道列表
func (m *Manager) PrintProxyList(proxies []ProxyInfo) {
	if len(proxies) == 0 {
		fmt.Println("未找到任何隧道")
		return
	}
	
	fmt.Printf("找到 %d 个隧道:\n\n", len(proxies))
	
	// 按类型分组
	groups := make(map[string][]ProxyInfo)
	for _, proxy := range proxies {
		groups[proxy.ProxyType] = append(groups[proxy.ProxyType], proxy)
	}
	
	for proxyType, typeProxies := range groups {
		fmt.Printf("=== %s 隧道 ===\n", strings.ToUpper(proxyType))
		for _, proxy := range typeProxies {
			status := "❌ 离线"
			if proxy.Online {
				status = "✅ 在线"
			} else if proxy.Status {
				status = "⏸️ 已启用"
			}
			
			fmt.Printf("ID: %d | %s | %s\n", proxy.ID, proxy.ProxyName, status)
			fmt.Printf("  节点: %s\n", proxy.FriendlyNode)
			fmt.Printf("  本地: %s:%d\n", proxy.LocalIP, proxy.LocalPort)
			
			if proxy.ProxyType == "tcp" || proxy.ProxyType == "udp" {
				fmt.Printf("  远程: %s\n", proxy.ConnectAddress)
			} else if proxy.ProxyType == "http" || proxy.ProxyType == "https" {
				if proxy.Domain != "" {
					fmt.Printf("  域名: %s\n", proxy.Domain)
				}
				fmt.Printf("  地址: %s\n", proxy.ConnectAddress)
			}
			
			if proxy.LastLogin != nil {
				lastLogin := time.Unix(*proxy.LastLogin/1000, 0)
				fmt.Printf("  最后连接: %s\n", lastLogin.Format("2006-01-02 15:04:05"))
			}
			
			fmt.Println()
		}
	}
}

// PrintNodeList 打印节点列表
func (m *Manager) PrintNodeList(nodes []NodeInfo) {
	if len(nodes) == 0 {
		fmt.Println("未找到任何节点")
		return
	}
	
	fmt.Printf("找到 %d 个节点:\n\n", len(nodes))
	
	// 按地区分组
	regions := map[int]string{
		1: "🇨🇳 中国大陆",
		2: "🇭🇰 港澳台",
		3: "🌍 海外地区",
	}
	
	for classify, regionName := range regions {
		var regionNodes []NodeInfo
		for _, node := range nodes {
			if node.Classify == classify {
				regionNodes = append(regionNodes, node)
			}
		}
		
		if len(regionNodes) == 0 {
			continue
		}
		
		fmt.Printf("=== %s ===\n", regionName)
		for _, node := range regionNodes {
			status := "❌ 离线"
			if node.Status == 200 {
				if node.FullyLoaded {
					status = "⚠️ 满载"
				} else {
					status = "✅ 在线"
				}
			}
			
			fmt.Printf("ID: %d | %s | %s\n", node.ID, node.Name, status)
			if node.Comments != "" {
				fmt.Printf("  标签: %s\n", node.Comments)
			}
			if node.Description != "" {
				fmt.Printf("  描述: %s\n", node.Description)
			}
			
			// 显示支持的协议
			var protocols []string
			for proto, supported := range node.ProtocolSupport {
				if supported {
					protocols = append(protocols, strings.ToUpper(proto))
				}
			}
			if len(protocols) > 0 {
				fmt.Printf("  协议: %s\n", strings.Join(protocols, ", "))
			}
			
			fmt.Println()
		}
	}
}

// CreateProxyFromTemplate 从模板创建隧道
func (m *Manager) CreateProxyFromTemplate(template ProxyTemplate, name string, localPort int, nodeID int) error {
	req := &CreateProxyRequest{
		Name:        name,
		Type:        template.Type,
		LocalAddr:   "127.0.0.1",
		LocalPort:   strconv.Itoa(localPort),
		NodeID:      nodeID,
		DataEncrypt: false,
		DataGzip:    false,
	}
	
	// 根据类型设置默认值
	switch template.Type {
	case "tcp", "udp":
		// TCP/UDP需要远程端口，这里使用0让系统自动分配
		req.RemotePort = 0
	case "http", "https":
		// HTTP/HTTPS可以设置自定义域名
		if domain, ok := template.Config["domain"]; ok {
			req.DomainBind = domain
		}
		if forceHTTPS, ok := template.Config["force_https"]; ok && forceHTTPS == "true" {
			req.ForceHTTPS = true
		}
	}
	
	return m.CreateProxy(req)
}

// GetProxyTemplates 获取隧道模板
func (m *Manager) GetProxyTemplates() []ProxyTemplate {
	return []ProxyTemplate{
		{
			Name:        "Minecraft Java版",
			Description: "Minecraft Java版服务器 (TCP 25565)",
			Type:        "tcp",
			LocalPort:   25565,
			Config:      map[string]string{},
		},
		{
			Name:        "Minecraft 基岩版",
			Description: "Minecraft 基岩版服务器 (UDP 19132)",
			Type:        "udp",
			LocalPort:   19132,
			Config:      map[string]string{},
		},
		{
			Name:        "Web服务器",
			Description: "HTTP网站服务器 (HTTP 80)",
			Type:        "http",
			LocalPort:   80,
			Config:      map[string]string{},
		},
		{
			Name:        "HTTPS网站",
			Description: "HTTPS安全网站 (HTTPS 443)",
			Type:        "https",
			LocalPort:   443,
			Config:      map[string]string{"force_https": "true"},
		},
		{
			Name:        "SSH服务器",
			Description: "SSH远程连接 (TCP 22)",
			Type:        "tcp",
			LocalPort:   22,
			Config:      map[string]string{},
		},
		{
			Name:        "FTP服务器",
			Description: "FTP文件传输 (TCP 21)",
			Type:        "tcp",
			LocalPort:   21,
			Config:      map[string]string{},
		},
	}
}
