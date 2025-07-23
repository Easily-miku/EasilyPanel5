package api

import (
	"fmt"
	"net/http"
	"runtime"
	"time"
)

// SystemStats 系统统计信息
type SystemStats struct {
	CPU        CPUStats    `json:"cpu"`
	Memory     MemoryStats `json:"memory"`
	Disk       DiskStats   `json:"disk"`
	Network    NetworkStats `json:"network"`
	Uptime     int64       `json:"uptime"`
	Timestamp  time.Time   `json:"timestamp"`
	GoRoutines int         `json:"goroutines"`
	GoVersion  string      `json:"go_version"`
}

// CPUStats CPU统计信息
type CPUStats struct {
	Usage     float64 `json:"usage"`      // CPU使用率百分比
	Cores     int     `json:"cores"`      // CPU核心数
	LoadAvg1  float64 `json:"load_avg_1"` // 1分钟平均负载
	LoadAvg5  float64 `json:"load_avg_5"` // 5分钟平均负载
	LoadAvg15 float64 `json:"load_avg_15"`// 15分钟平均负载
}

// MemoryStats 内存统计信息
type MemoryStats struct {
	Total     uint64  `json:"total"`      // 总内存 (bytes)
	Used      uint64  `json:"used"`       // 已使用内存 (bytes)
	Free      uint64  `json:"free"`       // 空闲内存 (bytes)
	Available uint64  `json:"available"`  // 可用内存 (bytes)
	Usage     float64 `json:"usage"`      // 内存使用率百分比
	Swap      SwapStats `json:"swap"`     // 交换分区信息
}

// SwapStats 交换分区统计信息
type SwapStats struct {
	Total uint64  `json:"total"` // 总交换分区 (bytes)
	Used  uint64  `json:"used"`  // 已使用交换分区 (bytes)
	Free  uint64  `json:"free"`  // 空闲交换分区 (bytes)
	Usage float64 `json:"usage"` // 交换分区使用率百分比
}

// DiskStats 磁盘统计信息
type DiskStats struct {
	Total uint64  `json:"total"` // 总磁盘空间 (bytes)
	Used  uint64  `json:"used"`  // 已使用磁盘空间 (bytes)
	Free  uint64  `json:"free"`  // 空闲磁盘空间 (bytes)
	Usage float64 `json:"usage"` // 磁盘使用率百分比
}

// NetworkStats 网络统计信息
type NetworkStats struct {
	BytesSent     uint64 `json:"bytes_sent"`     // 发送字节数
	BytesRecv     uint64 `json:"bytes_recv"`     // 接收字节数
	PacketsSent   uint64 `json:"packets_sent"`   // 发送包数
	PacketsRecv   uint64 `json:"packets_recv"`   // 接收包数
	ErrorsIn      uint64 `json:"errors_in"`      // 接收错误数
	ErrorsOut     uint64 `json:"errors_out"`     // 发送错误数
	DroppedIn     uint64 `json:"dropped_in"`     // 接收丢包数
	DroppedOut    uint64 `json:"dropped_out"`    // 发送丢包数
}

// ServerMonitoringStats 服务器监控统计
type ServerMonitoringStats struct {
	ServerID    string    `json:"server_id"`
	Name        string    `json:"name"`
	Status      string    `json:"status"`
	PID         int       `json:"pid"`
	CPU         float64   `json:"cpu"`         // CPU使用率
	Memory      uint64    `json:"memory"`      // 内存使用量 (bytes)
	Uptime      int64     `json:"uptime"`      // 运行时间 (seconds)
	Players     int       `json:"players"`     // 在线玩家数
	MaxPlayers  int       `json:"max_players"` // 最大玩家数
	LastUpdate  time.Time `json:"last_update"`
}

// handleSystemStats 处理系统统计请求
func handleSystemStats(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	stats, err := getSystemStats()
	if err != nil {
		WriteStandardError(w, "GET_STATS_FAILED", fmt.Sprintf("Failed to get system stats: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, stats)
}

// handleServerStats 处理服务器统计请求
func handleServerStats(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// 获取时间范围参数
	timeRange := r.URL.Query().Get("range")
	if timeRange == "" {
		timeRange = "1h"
	}

	stats, err := getServerStats(timeRange)
	if err != nil {
		WriteStandardError(w, "GET_SERVER_STATS_FAILED", fmt.Sprintf("Failed to get server stats: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, stats)
}

// handleHistoricalStats 处理历史统计请求
func handleHistoricalStats(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// 获取参数
	statsType := r.URL.Query().Get("type") // system, server
	timeRange := r.URL.Query().Get("range") // 1h, 6h, 24h, 7d
	serverID := r.URL.Query().Get("server_id")

	if statsType == "" {
		statsType = "system"
	}
	if timeRange == "" {
		timeRange = "1h"
	}

	var stats interface{}
	var err error

	switch statsType {
	case "system":
		stats, err = getHistoricalSystemStats(timeRange)
	case "server":
		if serverID == "" {
			WriteStandardError(w, "MISSING_SERVER_ID", "Server ID is required for server stats", http.StatusBadRequest)
			return
		}
		stats, err = getHistoricalServerStats(serverID, timeRange)
	default:
		WriteStandardError(w, "INVALID_STATS_TYPE", "Invalid stats type", http.StatusBadRequest)
		return
	}

	if err != nil {
		WriteStandardError(w, "GET_HISTORICAL_STATS_FAILED", fmt.Sprintf("Failed to get historical stats: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, stats)
}

// getSystemStats 获取系统统计信息
func getSystemStats() (*SystemStats, error) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	stats := &SystemStats{
		CPU: CPUStats{
			Usage: getCPUUsage(),
			Cores: runtime.NumCPU(),
		},
		Memory: MemoryStats{
			Total: getMemoryTotal(),
			Used:  memStats.Sys,
			Free:  getMemoryFree(),
		},
		Disk: DiskStats{
			Total: getDiskTotal(),
			Used:  getDiskUsed(),
			Free:  getDiskFree(),
		},
		Network: NetworkStats{
			BytesSent:   getNetworkBytesSent(),
			BytesRecv:   getNetworkBytesRecv(),
			PacketsSent: getNetworkPacketsSent(),
			PacketsRecv: getNetworkPacketsRecv(),
		},
		Uptime:     getSystemUptime(),
		Timestamp:  time.Now(),
		GoRoutines: runtime.NumGoroutine(),
		GoVersion:  runtime.Version(),
	}

	// 计算使用率
	if stats.Memory.Total > 0 {
		stats.Memory.Usage = float64(stats.Memory.Used) / float64(stats.Memory.Total) * 100
	}
	if stats.Disk.Total > 0 {
		stats.Disk.Usage = float64(stats.Disk.Used) / float64(stats.Disk.Total) * 100
	}

	return stats, nil
}

// getServerStats 获取服务器统计信息
func getServerStats(timeRange string) ([]ServerMonitoringStats, error) {
	// 简化实现：返回模拟数据
	// 实际应用中应该从服务器管理模块获取真实数据
	var stats []ServerMonitoringStats

	// 模拟一些服务器数据
	mockServers := []struct {
		ID         string
		Name       string
		Status     string
		MaxPlayers int
	}{
		{"server1", "生存服务器", "running", 20},
		{"server2", "创造服务器", "stopped", 30},
		{"server3", "小游戏服务器", "running", 50},
	}

	for _, srv := range mockServers {
		serverStats := ServerMonitoringStats{
			ServerID:   srv.ID,
			Name:       srv.Name,
			Status:     srv.Status,
			PID:        1234,
			CPU:        getServerCPUUsage(srv.ID),
			Memory:     getServerMemoryUsage(srv.ID),
			Uptime:     getServerUptime(srv.ID),
			Players:    getServerPlayerCount(srv.ID),
			MaxPlayers: srv.MaxPlayers,
			LastUpdate: time.Now(),
		}
		stats = append(stats, serverStats)
	}

	return stats, nil
}

// getHistoricalSystemStats 获取历史系统统计
func getHistoricalSystemStats(timeRange string) ([]SystemStats, error) {
	// 简化实现：返回模拟数据
	// 实际应用中应该从数据库或时间序列数据库中获取
	var stats []SystemStats
	
	duration := parseTimeRange(timeRange)
	points := getDataPoints(timeRange)
	interval := duration / time.Duration(points)
	
	now := time.Now()
	for i := 0; i < points; i++ {
		timestamp := now.Add(-duration + time.Duration(i)*interval)
		stat, _ := getSystemStats()
		stat.Timestamp = timestamp
		// 添加一些随机变化使数据更真实
		stat.CPU.Usage = stat.CPU.Usage + float64(i%10-5)
		if stat.CPU.Usage < 0 {
			stat.CPU.Usage = 0
		}
		if stat.CPU.Usage > 100 {
			stat.CPU.Usage = 100
		}
		stats = append(stats, *stat)
	}
	
	return stats, nil
}

// getHistoricalServerStats 获取历史服务器统计
func getHistoricalServerStats(serverID, timeRange string) ([]ServerMonitoringStats, error) {
	// 简化实现：返回模拟数据
	var stats []ServerMonitoringStats
	
	duration := parseTimeRange(timeRange)
	points := getDataPoints(timeRange)
	interval := duration / time.Duration(points)
	
	now := time.Now()
	for i := 0; i < points; i++ {
		timestamp := now.Add(-duration + time.Duration(i)*interval)
		stat := ServerMonitoringStats{
			ServerID:   serverID,
			CPU:        float64(i%20 + 10), // 模拟CPU使用率
			Memory:     uint64(i%100 + 50) * 1024 * 1024, // 模拟内存使用
			Players:    i % 10, // 模拟玩家数
			LastUpdate: timestamp,
		}
		stats = append(stats, stat)
	}
	
	return stats, nil
}

// 辅助函数（简化实现）
func getCPUUsage() float64 { return 25.5 } // 模拟CPU使用率
func getMemoryTotal() uint64 { return 8 * 1024 * 1024 * 1024 } // 8GB
func getMemoryFree() uint64 { return 4 * 1024 * 1024 * 1024 } // 4GB
func getDiskTotal() uint64 { return 500 * 1024 * 1024 * 1024 } // 500GB
func getDiskUsed() uint64 { return 200 * 1024 * 1024 * 1024 } // 200GB
func getDiskFree() uint64 { return 300 * 1024 * 1024 * 1024 } // 300GB
func getNetworkBytesSent() uint64 { return 1024 * 1024 * 100 } // 100MB
func getNetworkBytesRecv() uint64 { return 1024 * 1024 * 200 } // 200MB
func getNetworkPacketsSent() uint64 { return 10000 }
func getNetworkPacketsRecv() uint64 { return 15000 }
func getSystemUptime() int64 { return 86400 * 7 } // 7天
func getServerCPUUsage(serverID string) float64 { return 15.2 }
func getServerMemoryUsage(serverID string) uint64 { return 512 * 1024 * 1024 } // 512MB
func getServerUptime(serverID string) int64 { return 3600 * 24 } // 1天
func getServerPlayerCount(serverID string) int { return 5 }

func parseTimeRange(timeRange string) time.Duration {
	switch timeRange {
	case "1h": return time.Hour
	case "6h": return 6 * time.Hour
	case "24h": return 24 * time.Hour
	case "7d": return 7 * 24 * time.Hour
	default: return time.Hour
	}
}

func getDataPoints(timeRange string) int {
	switch timeRange {
	case "1h": return 60   // 每分钟一个点
	case "6h": return 72   // 每5分钟一个点
	case "24h": return 96  // 每15分钟一个点
	case "7d": return 168  // 每小时一个点
	default: return 60
	}
}
