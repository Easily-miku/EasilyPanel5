package frp

import (
	"sync"
	"time"
)

var (
	globalManager FRPManager
	managerOnce   sync.Once
)

// GetManager 获取全局FRP管理器实例
func GetManager() FRPManager {
	managerOnce.Do(func() {
		config := getDefaultConfig()
		globalManager = NewManager(config)
	})
	return globalManager
}

// getDefaultConfig 获取默认FRP配置
func getDefaultConfig() *OpenFRPConfig {
	return &OpenFRPConfig{
		Enabled:            false, // 默认禁用，需要用户手动启用
		APIEndpoint:        "https://api.openfrp.net",
		DefaultNode:        "",
		AutoStart:          false,
		AutoRestart:        true,
		MaxTunnels:         10,
		MonitorInterval:    30 * time.Second,
		StatsRetention:     24 * time.Hour,
		DefaultBandwidth:   1024, // 1MB/s
		DefaultCompression: true,
		DefaultEncryption:  false,
	}
}

// InitManager 初始化FRP管理器
func InitManager(config *OpenFRPConfig) error {
	if config == nil {
		config = getDefaultConfig()
	}
	
	manager := NewManager(config)
	globalManager = manager
	
	return manager.Start()
}

// StartManager 启动FRP管理器
func StartManager() error {
	if globalManager == nil {
		return InitManager(nil)
	}
	return globalManager.Start()
}

// StopManager 停止FRP管理器
func StopManager() error {
	if globalManager == nil {
		return nil
	}
	return globalManager.Stop()
}

// IsManagerRunning 检查FRP管理器是否运行中
func IsManagerRunning() bool {
	if globalManager == nil {
		return false
	}
	return globalManager.IsRunning()
}

// 便捷函数，直接调用全局管理器的方法

// CreateTunnel 创建隧道
func CreateTunnel(req *TunnelRequest) (*TunnelConfig, error) {
	return GetManager().CreateTunnel(req)
}

// DeleteTunnel 删除隧道
func DeleteTunnel(tunnelID string) error {
	return GetManager().DeleteTunnel(tunnelID)
}

// StartTunnel 启动隧道
func StartTunnel(tunnelID string) error {
	return GetManager().StartTunnel(tunnelID)
}

// StopTunnel 停止隧道
func StopTunnel(tunnelID string) error {
	return GetManager().StopTunnel(tunnelID)
}

// RestartTunnel 重启隧道
func RestartTunnel(tunnelID string) error {
	return GetManager().RestartTunnel(tunnelID)
}

// GetTunnel 获取隧道
func GetTunnel(tunnelID string) (*TunnelConfig, error) {
	return GetManager().GetTunnel(tunnelID)
}

// GetAllTunnels 获取所有隧道
func GetAllTunnels() ([]*TunnelConfig, error) {
	return GetManager().GetAllTunnels()
}

// GetTunnelsByServer 获取指定服务器的隧道
func GetTunnelsByServer(serverID string) ([]*TunnelConfig, error) {
	return GetManager().GetTunnelsByServer(serverID)
}

// GetTunnelStats 获取隧道统计
func GetTunnelStats(tunnelID string) (*TunnelStats, error) {
	return GetManager().GetTunnelStats(tunnelID)
}

// GetAllStats 获取所有隧道统计
func GetAllStats() (map[string]*TunnelStats, error) {
	return GetManager().GetAllStats()
}

// GetNodes 获取节点列表
func GetNodes() ([]*OpenFRPNode, error) {
	return GetManager().GetNodes()
}

// GetUserInfo 获取用户信息
func GetUserInfo(token string) (*OpenFRPUser, error) {
	return GetManager().GetUserInfo(token)
}

// ValidateToken 验证令牌
func ValidateToken(token string) error {
	return GetManager().ValidateToken(token)
}

// CreateTunnelForServer 为服务器创建隧道的便捷方法
func CreateTunnelForServer(serverID string, serverPort int, token string, node string) (*TunnelConfig, error) {
	req := &TunnelRequest{
		Name:      "Minecraft-" + serverID,
		Type:      TunnelTypeTCP,
		LocalIP:   "127.0.0.1",
		LocalPort: serverPort,
		ServerID:  serverID,
		Token:     token,
		Node:      node,
	}
	
	return CreateTunnel(req)
}

// GetServerTunnelStatus 获取服务器隧道状态
func GetServerTunnelStatus(serverID string) (map[string]interface{}, error) {
	tunnels, err := GetTunnelsByServer(serverID)
	if err != nil {
		return nil, err
	}
	
	status := map[string]interface{}{
		"server_id":     serverID,
		"tunnel_count":  len(tunnels),
		"running_count": 0,
		"tunnels":       make([]map[string]interface{}, 0),
	}
	
	runningCount := 0
	for _, tunnel := range tunnels {
		if tunnel.Status == TunnelStatusRunning {
			runningCount++
		}
		
		tunnelInfo := map[string]interface{}{
			"id":          tunnel.ID,
			"name":        tunnel.Name,
			"type":        tunnel.Type,
			"status":      tunnel.Status,
			"local_port":  tunnel.LocalPort,
			"remote_port": tunnel.RemotePort,
			"node":        tunnel.Node,
		}
		
		status["tunnels"] = append(status["tunnels"].([]map[string]interface{}), tunnelInfo)
	}
	
	status["running_count"] = runningCount
	return status, nil
}
