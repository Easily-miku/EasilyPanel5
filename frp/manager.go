package frp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"easilypanel5/utils"
)

// Manager FRP管理器实现
type Manager struct {
	config      *OpenFRPConfig
	client      *OpenFRPClient
	tunnels     map[string]*TunnelConfig
	mutex       sync.RWMutex
	isRunning   bool
	stopChan    chan bool
	monitorTicker *time.Ticker
	configFile  string
}

// NewManager 创建新的FRP管理器
func NewManager(config *OpenFRPConfig) *Manager {
	return &Manager{
		config:     config,
		client:     NewOpenFRPClient(config.APIEndpoint),
		tunnels:    make(map[string]*TunnelConfig),
		stopChan:   make(chan bool),
		configFile: "data/frp_tunnels.json",
	}
}

// Start 启动FRP管理器
func (m *Manager) Start() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.isRunning {
		return fmt.Errorf("FRP manager is already running")
	}

	log.Println("启动FRP管理器")

	// 加载隧道配置
	if err := m.loadTunnels(); err != nil {
		log.Printf("Failed to load tunnels: %v", err)
	}

	// 启动监控
	if m.config.MonitorInterval > 0 {
		m.monitorTicker = time.NewTicker(m.config.MonitorInterval)
		go m.monitorLoop()
	}

	// 自动启动隧道
	if m.config.AutoStart {
		go m.autoStartTunnels()
	}

	m.isRunning = true
	return nil
}

// Stop 停止FRP管理器
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if !m.isRunning {
		return nil
	}

	log.Println("停止FRP管理器")

	// 停止监控
	if m.monitorTicker != nil {
		m.monitorTicker.Stop()
		m.stopChan <- true
	}

	// 停止所有隧道
	for _, tunnel := range m.tunnels {
		if tunnel.Status == TunnelStatusRunning {
			m.stopTunnelInternal(tunnel.ID)
		}
	}

	m.isRunning = false
	return nil
}

// IsRunning 检查管理器是否运行中
func (m *Manager) IsRunning() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.isRunning
}

// CreateTunnel 创建隧道
func (m *Manager) CreateTunnel(req *TunnelRequest) (*TunnelConfig, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查隧道数量限制
	if len(m.tunnels) >= m.config.MaxTunnels {
		return nil, fmt.Errorf("maximum tunnel limit reached: %d", m.config.MaxTunnels)
	}

	// 验证令牌
	if err := m.client.ValidateToken(req.Token); err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	// 设置默认值
	if req.Bandwidth == 0 {
		req.Bandwidth = m.config.DefaultBandwidth
	}
	if req.LocalIP == "" {
		req.LocalIP = "127.0.0.1"
	}

	// 生成隧道ID
	tunnelID := fmt.Sprintf("tunnel_%d", time.Now().Unix())

	// 创建隧道配置
	tunnel := &TunnelConfig{
		ID:           tunnelID,
		Name:         req.Name,
		Type:         req.Type,
		LocalIP:      req.LocalIP,
		LocalPort:    req.LocalPort,
		RemotePort:   req.RemotePort,
		CustomDomain: req.CustomDomain,
		ServerID:     req.ServerID,
		Status:       TunnelStatusStopped,
		Token:        req.Token,
		Node:         req.Node,
		Bandwidth:    req.Bandwidth,
		Compression:  req.Compression,
		Encryption:   req.Encryption,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// 保存隧道
	m.tunnels[tunnelID] = tunnel
	if err := m.saveTunnels(); err != nil {
		delete(m.tunnels, tunnelID)
		return nil, fmt.Errorf("failed to save tunnel: %v", err)
	}

	// 发送事件
	utils.EmitEvent(EventTunnelCreated, tunnel.ServerID, tunnel)

	return tunnel, nil
}

// DeleteTunnel 删除隧道
func (m *Manager) DeleteTunnel(tunnelID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	tunnel, exists := m.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	// 如果隧道正在运行，先停止
	if tunnel.Status == TunnelStatusRunning {
		if err := m.stopTunnelInternal(tunnelID); err != nil {
			return fmt.Errorf("failed to stop tunnel before deletion: %v", err)
		}
	}

	// 删除隧道
	delete(m.tunnels, tunnelID)
	if err := m.saveTunnels(); err != nil {
		return fmt.Errorf("failed to save tunnels: %v", err)
	}

	// 发送事件
	utils.EmitEvent(EventTunnelDeleted, tunnel.ServerID, map[string]interface{}{
		"tunnel_id": tunnelID,
		"name":      tunnel.Name,
	})

	return nil
}

// StartTunnel 启动隧道
func (m *Manager) StartTunnel(tunnelID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.startTunnelInternal(tunnelID)
}

// startTunnelInternal 内部启动隧道方法
func (m *Manager) startTunnelInternal(tunnelID string) error {
	tunnel, exists := m.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	if tunnel.Status == TunnelStatusRunning {
		return fmt.Errorf("tunnel is already running")
	}

	log.Printf("启动隧道: %s (%s)", tunnel.Name, tunnelID)

	// 更新状态
	tunnel.Status = TunnelStatusStarting
	tunnel.UpdatedAt = time.Now()

	// 通过API启动隧道
	if err := m.client.StartTunnel(tunnel.Token, tunnelID); err != nil {
		tunnel.Status = TunnelStatusError
		tunnel.UpdatedAt = time.Now()
		
		utils.EmitEvent(EventTunnelError, tunnel.ServerID, map[string]interface{}{
			"tunnel_id": tunnelID,
			"error":     err.Error(),
		})
		
		return fmt.Errorf("failed to start tunnel via API: %v", err)
	}

	// 更新状态
	tunnel.Status = TunnelStatusRunning
	tunnel.UpdatedAt = time.Now()

	// 保存配置
	m.saveTunnels()

	// 发送事件
	utils.EmitEvent(EventTunnelStarted, tunnel.ServerID, tunnel)

	return nil
}

// StopTunnel 停止隧道
func (m *Manager) StopTunnel(tunnelID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.stopTunnelInternal(tunnelID)
}

// stopTunnelInternal 内部停止隧道方法
func (m *Manager) stopTunnelInternal(tunnelID string) error {
	tunnel, exists := m.tunnels[tunnelID]
	if !exists {
		return fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	if tunnel.Status != TunnelStatusRunning {
		return fmt.Errorf("tunnel is not running")
	}

	log.Printf("停止隧道: %s (%s)", tunnel.Name, tunnelID)

	// 通过API停止隧道
	if err := m.client.StopTunnel(tunnel.Token, tunnelID); err != nil {
		log.Printf("Failed to stop tunnel via API: %v", err)
		// 继续执行，更新本地状态
	}

	// 更新状态
	tunnel.Status = TunnelStatusStopped
	tunnel.UpdatedAt = time.Now()

	// 保存配置
	m.saveTunnels()

	// 发送事件
	utils.EmitEvent(EventTunnelStopped, tunnel.ServerID, tunnel)

	return nil
}

// RestartTunnel 重启隧道
func (m *Manager) RestartTunnel(tunnelID string) error {
	if err := m.StopTunnel(tunnelID); err != nil {
		return err
	}

	// 等待一段时间
	time.Sleep(2 * time.Second)

	return m.StartTunnel(tunnelID)
}

// GetTunnel 获取隧道
func (m *Manager) GetTunnel(tunnelID string) (*TunnelConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tunnel, exists := m.tunnels[tunnelID]
	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	return tunnel, nil
}

// GetAllTunnels 获取所有隧道
func (m *Manager) GetAllTunnels() ([]*TunnelConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	tunnels := make([]*TunnelConfig, 0, len(m.tunnels))
	for _, tunnel := range m.tunnels {
		tunnels = append(tunnels, tunnel)
	}

	return tunnels, nil
}

// GetTunnelsByServer 获取指定服务器的隧道
func (m *Manager) GetTunnelsByServer(serverID string) ([]*TunnelConfig, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var tunnels []*TunnelConfig
	for _, tunnel := range m.tunnels {
		if tunnel.ServerID == serverID {
			tunnels = append(tunnels, tunnel)
		}
	}

	return tunnels, nil
}

// GetTunnelStats 获取隧道统计
func (m *Manager) GetTunnelStats(tunnelID string) (*TunnelStats, error) {
	m.mutex.RLock()
	tunnel, exists := m.tunnels[tunnelID]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("tunnel not found: %s", tunnelID)
	}

	// 从API获取最新统计
	stats, err := m.client.GetTunnelStats(tunnel.Token, tunnelID)
	if err != nil {
		// 如果API调用失败，返回本地缓存的统计
		if tunnel.Stats != nil {
			return tunnel.Stats, nil
		}
		return nil, fmt.Errorf("failed to get tunnel stats: %v", err)
	}

	// 更新本地缓存
	m.mutex.Lock()
	tunnel.Stats = stats
	m.mutex.Unlock()

	return stats, nil
}

// GetAllStats 获取所有隧道统计
func (m *Manager) GetAllStats() (map[string]*TunnelStats, error) {
	m.mutex.RLock()
	tunnels := make(map[string]*TunnelConfig)
	for k, v := range m.tunnels {
		tunnels[k] = v
	}
	m.mutex.RUnlock()

	stats := make(map[string]*TunnelStats)
	for tunnelID, tunnel := range tunnels {
		if tunnel.Status == TunnelStatusRunning {
			if tunnelStats, err := m.GetTunnelStats(tunnelID); err == nil {
				stats[tunnelID] = tunnelStats
			}
		}
	}

	return stats, nil
}

// GetNodes 获取节点列表
func (m *Manager) GetNodes() ([]*OpenFRPNode, error) {
	return m.client.GetNodes()
}

// GetUserInfo 获取用户信息
func (m *Manager) GetUserInfo(token string) (*OpenFRPUser, error) {
	return m.client.GetUserInfo(token)
}

// ValidateToken 验证令牌
func (m *Manager) ValidateToken(token string) error {
	return m.client.ValidateToken(token)
}

// loadTunnels 加载隧道配置
func (m *Manager) loadTunnels() error {
	if _, err := os.Stat(m.configFile); os.IsNotExist(err) {
		return nil // 文件不存在，跳过加载
	}

	data, err := os.ReadFile(m.configFile)
	if err != nil {
		return fmt.Errorf("failed to read tunnels file: %v", err)
	}

	var tunnels map[string]*TunnelConfig
	if err := json.Unmarshal(data, &tunnels); err != nil {
		return fmt.Errorf("failed to parse tunnels file: %v", err)
	}

	m.tunnels = tunnels
	log.Printf("加载了 %d 个隧道配置", len(tunnels))
	return nil
}

// saveTunnels 保存隧道配置
func (m *Manager) saveTunnels() error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(m.configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(m.tunnels, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tunnels: %v", err)
	}

	if err := os.WriteFile(m.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write tunnels file: %v", err)
	}

	return nil
}

// monitorLoop 监控循环
func (m *Manager) monitorLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("FRP monitor loop panic: %v", r)
		}
	}()

	for {
		select {
		case <-m.monitorTicker.C:
			m.updateTunnelStats()
		case <-m.stopChan:
			return
		}
	}
}

// updateTunnelStats 更新隧道统计
func (m *Manager) updateTunnelStats() {
	m.mutex.RLock()
	runningTunnels := make([]*TunnelConfig, 0)
	for _, tunnel := range m.tunnels {
		if tunnel.Status == TunnelStatusRunning {
			runningTunnels = append(runningTunnels, tunnel)
		}
	}
	m.mutex.RUnlock()

	for _, tunnel := range runningTunnels {
		stats, err := m.client.GetTunnelStats(tunnel.Token, tunnel.ID)
		if err != nil {
			log.Printf("Failed to get stats for tunnel %s: %v", tunnel.ID, err)
			continue
		}

		m.mutex.Lock()
		tunnel.Stats = stats
		m.mutex.Unlock()

		// 发送统计更新事件
		utils.EmitEvent(EventTunnelStatsUpdate, tunnel.ServerID, map[string]interface{}{
			"tunnel_id": tunnel.ID,
			"stats":     stats,
		})
	}
}

// autoStartTunnels 自动启动隧道
func (m *Manager) autoStartTunnels() {
	time.Sleep(5 * time.Second) // 等待系统稳定

	m.mutex.RLock()
	tunnels := make([]*TunnelConfig, 0)
	for _, tunnel := range m.tunnels {
		if tunnel.Status == TunnelStatusStopped {
			tunnels = append(tunnels, tunnel)
		}
	}
	m.mutex.RUnlock()

	for _, tunnel := range tunnels {
		log.Printf("自动启动隧道: %s", tunnel.Name)
		if err := m.StartTunnel(tunnel.ID); err != nil {
			log.Printf("Failed to auto-start tunnel %s: %v", tunnel.Name, err)
		}
		time.Sleep(2 * time.Second) // 避免同时启动太多隧道
	}
}
