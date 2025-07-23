package server

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"easilypanel5/config"
	"easilypanel5/utils"
)

// ProcessDaemon 进程守护管理器
type ProcessDaemon struct {
	processes       map[string]*DaemonProcess
	mutex           sync.RWMutex
	monitorTicker   *time.Ticker
	monitorStop     chan bool
	isRunning       bool
	resourceMonitor *ResourceMonitor
}

// DaemonProcess 守护进程信息
type DaemonProcess struct {
	ServerID        string
	Process         *ServerProcess
	Server          *config.MinecraftServer
	RestartAttempts int
	LastCrashTime   *time.Time
	IsMonitoring    bool
	StopRequested   bool
}

// ResourceMonitor 资源监控器
type ResourceMonitor struct {
	processes map[string]*ProcessStats
	mutex     sync.RWMutex
}

// ProcessStats 进程统计信息
type ProcessStats struct {
	PID           int
	CPUPercent    float64
	MemoryUsed    int64
	MemoryPercent float64
	LastUpdate    time.Time
}

var (
	globalDaemon *ProcessDaemon
	daemonOnce   sync.Once
)

// GetDaemon 获取全局守护管理器实例
func GetDaemon() *ProcessDaemon {
	daemonOnce.Do(func() {
		globalDaemon = &ProcessDaemon{
			processes:       make(map[string]*DaemonProcess),
			monitorStop:     make(chan bool),
			resourceMonitor: &ResourceMonitor{
				processes: make(map[string]*ProcessStats),
			},
		}
	})
	return globalDaemon
}

// Start 启动守护管理器
func (d *ProcessDaemon) Start() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.isRunning {
		return fmt.Errorf("daemon is already running")
	}

	log.Println("启动进程守护管理器")

	// 启动监控协程
	cfg := config.Get()
	if cfg.Daemon.ResourceMonitoring {
		d.monitorTicker = time.NewTicker(cfg.Daemon.MonitorInterval)
		go d.monitorLoop()
	}

	d.isRunning = true

	// 恢复之前的服务器状态
	d.recoverServers()

	return nil
}

// Stop 停止守护管理器
func (d *ProcessDaemon) Stop() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.isRunning {
		return nil
	}

	log.Println("停止进程守护管理器")

	// 停止监控
	if d.monitorTicker != nil {
		d.monitorTicker.Stop()
		d.monitorStop <- true
	}

	// 停止所有进程
	for serverID, dp := range d.processes {
		dp.StopRequested = true
		if dp.Process != nil {
			log.Printf("停止服务器进程: %s", serverID)
			// 这里应该调用优雅停止
		}
	}

	d.isRunning = false
	return nil
}

// RegisterProcess 注册进程到守护管理器
func (d *ProcessDaemon) RegisterProcess(serverID string, process *ServerProcess, server *config.MinecraftServer) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	log.Printf("注册进程到守护管理器: %s (PID: %d)", serverID, process.Server.PID)

	dp := &DaemonProcess{
		ServerID:        serverID,
		Process:         process,
		Server:          server,
		RestartAttempts: 0,
		IsMonitoring:    true,
		StopRequested:   false,
	}

	d.processes[serverID] = dp

	// 启动进程监控
	go d.monitorProcess(dp)
}

// UnregisterProcess 从守护管理器注销进程
func (d *ProcessDaemon) UnregisterProcess(serverID string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if dp, exists := d.processes[serverID]; exists {
		log.Printf("从守护管理器注销进程: %s", serverID)
		dp.IsMonitoring = false
		dp.StopRequested = true
		delete(d.processes, serverID)

		// 清理资源监控数据
		d.resourceMonitor.mutex.Lock()
		delete(d.resourceMonitor.processes, serverID)
		d.resourceMonitor.mutex.Unlock()
	}
}

// monitorProcess 监控单个进程
func (d *ProcessDaemon) monitorProcess(dp *DaemonProcess) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("进程监控协程异常退出: %v", r)
		}
	}()

	for dp.IsMonitoring && !dp.StopRequested {
		// 检查进程是否还在运行
		if dp.Process != nil && dp.Process.Cmd != nil && dp.Process.Cmd.Process != nil {
			if !isProcessRunning(dp.Process.Cmd.Process.Pid) {
				log.Printf("检测到进程 %s 已退出", dp.ServerID)
				d.handleProcessExit(dp)
				break
			}
		}

		time.Sleep(time.Second)
	}
}

// handleProcessExit 处理进程退出
func (d *ProcessDaemon) handleProcessExit(dp *DaemonProcess) {
	if dp.StopRequested {
		log.Printf("进程 %s 正常停止", dp.ServerID)
		return
	}

	log.Printf("进程 %s 意外退出，准备重启", dp.ServerID)

	// 记录崩溃信息
	now := time.Now()
	dp.LastCrashTime = &now
	dp.RestartAttempts++

	// 更新服务器状态
	dp.Server.Status = config.StatusCrashed
	dp.Server.LastCrashTime = &now
	dp.Server.CrashCount++
	dp.Server.RestartAttempts = dp.RestartAttempts

	// 发送崩溃事件
	utils.EmitEvent("server_crashed", dp.ServerID, map[string]interface{}{
		"crash_time":       now,
		"restart_attempts": dp.RestartAttempts,
	})

	// 检查是否应该重启
	cfg := config.Get()
	maxAttempts := dp.Server.MaxRestartAttempts
	if maxAttempts == 0 {
		maxAttempts = cfg.Daemon.MaxRestartAttempts
	}

	if cfg.Daemon.EnableAutoRestart && dp.RestartAttempts < maxAttempts {
		log.Printf("将在 %v 后重启服务器 %s (尝试 %d/%d)", 
			cfg.Daemon.RestartDelay, dp.ServerID, dp.RestartAttempts, maxAttempts)

		// 延迟重启
		time.AfterFunc(cfg.Daemon.RestartDelay, func() {
			d.restartProcess(dp)
		})
	} else {
		log.Printf("服务器 %s 达到最大重启次数，停止自动重启", dp.ServerID)
		utils.EmitEvent("server_restart_failed", dp.ServerID, map[string]interface{}{
			"max_attempts": maxAttempts,
			"final_crash":  now,
		})
	}
}

// restartProcess 重启进程
func (d *ProcessDaemon) restartProcess(dp *DaemonProcess) {
	if dp.StopRequested {
		return
	}

	log.Printf("重启服务器进程: %s", dp.ServerID)

	// 这里应该调用服务器启动逻辑
	// 由于需要与现有的server包集成，这里先发送重启事件
	utils.EmitEvent("server_restart_requested", dp.ServerID, map[string]interface{}{
		"restart_attempts": dp.RestartAttempts,
		"auto_restart":     true,
	})
}

// monitorLoop 资源监控循环
func (d *ProcessDaemon) monitorLoop() {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("资源监控协程异常退出: %v", r)
		}
	}()

	for {
		select {
		case <-d.monitorTicker.C:
			d.updateResourceStats()
		case <-d.monitorStop:
			return
		}
	}
}

// updateResourceStats 更新资源统计
func (d *ProcessDaemon) updateResourceStats() {
	d.mutex.RLock()
	processes := make(map[string]*DaemonProcess)
	for k, v := range d.processes {
		processes[k] = v
	}
	d.mutex.RUnlock()

	for serverID, dp := range processes {
		if dp.Process != nil && dp.Process.Cmd != nil && dp.Process.Cmd.Process != nil {
			stats := d.collectProcessStats(dp.Process.Cmd.Process.Pid)
			if stats != nil {
				d.resourceMonitor.mutex.Lock()
				d.resourceMonitor.processes[serverID] = stats
				d.resourceMonitor.mutex.Unlock()

				// 更新服务器配置中的资源使用情况
				dp.Server.ResourceUsage = &config.ResourceUsage{
					CPUPercent:    stats.CPUPercent,
					MemoryUsed:    stats.MemoryUsed,
					MemoryPercent: stats.MemoryPercent,
					UpdatedAt:     stats.LastUpdate,
				}

				// 发送资源使用事件
				utils.EmitEvent("resource_usage", serverID, dp.Server.ResourceUsage)
			}
		}
	}
}

// collectProcessStats 收集进程统计信息
func (d *ProcessDaemon) collectProcessStats(pid int) *ProcessStats {
	// 这里应该实现实际的进程资源收集逻辑
	// 由于跨平台兼容性，这里先返回模拟数据
	return &ProcessStats{
		PID:           pid,
		CPUPercent:    0.0, // 需要实现实际的CPU使用率计算
		MemoryUsed:    0,   // 需要实现实际的内存使用量计算
		MemoryPercent: 0.0, // 需要实现实际的内存使用率计算
		LastUpdate:    time.Now(),
	}
}

// recoverServers 恢复服务器状态
func (d *ProcessDaemon) recoverServers() {
	// 这里应该从配置中恢复之前运行的服务器
	// 并检查它们的状态
	log.Println("恢复服务器状态...")
}

// GetProcessStats 获取进程统计信息
func (d *ProcessDaemon) GetProcessStats(serverID string) *ProcessStats {
	d.resourceMonitor.mutex.RLock()
	defer d.resourceMonitor.mutex.RUnlock()

	if stats, exists := d.resourceMonitor.processes[serverID]; exists {
		return stats
	}
	return nil
}

// GetAllProcessStats 获取所有进程统计信息
func (d *ProcessDaemon) GetAllProcessStats() map[string]*ProcessStats {
	d.resourceMonitor.mutex.RLock()
	defer d.resourceMonitor.mutex.RUnlock()

	result := make(map[string]*ProcessStats)
	for k, v := range d.resourceMonitor.processes {
		result[k] = v
	}
	return result
}

// isProcessRunning 检查进程是否还在运行（跨平台）
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}

	if runtime.GOOS == "windows" {
		// Windows平台
		process, err := os.FindProcess(pid)
		if err != nil {
			return false
		}

		// 在Windows上，FindProcess总是成功，需要发送信号0来检查
		err = process.Signal(syscall.Signal(0))
		return err == nil
	} else {
		// Unix/Linux平台
		process, err := os.FindProcess(pid)
		if err != nil {
			return false
		}

		// 发送信号0来检查进程是否存在
		err = process.Signal(syscall.Signal(0))
		if err != nil {
			// 如果是权限错误，说明进程存在但没有权限
			if err.Error() == "operation not permitted" {
				return true
			}
			return false
		}
		return true
	}
}
