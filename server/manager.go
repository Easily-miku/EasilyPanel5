package server

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"easilypanel5/config"
	"easilypanel5/core"
)

// GetJavaInfo 检测Java环境 - 导出给API使用
func GetJavaInfo() ([]config.JavaInfo, error) {
	return detectJavaEnvironment()
}

// CheckJavaPath 验证Java路径 - 导出给API使用
func CheckJavaPath(javaPath string) (*config.JavaInfo, error) {
	return validateJavaPath(javaPath)
}

// GetCoresList 获取核心列表
func GetCoresList() ([]core.FastMirrorCore, error) {
	api := core.NewFastMirrorAPI()
	return api.GetCoresList()
}

// GetCoreVersions 获取核心版本列表
func GetCoreVersions(coreType string) (interface{}, error) {
	api := core.NewFastMirrorAPI()
	
	// 首先获取项目信息
	project, err := api.GetProjectInfo(coreType)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"name":        project.Name,
		"tag":         project.Tag,
		"homepage":    project.Homepage,
		"mc_versions": project.MCVersions,
	}, nil
}

// StartDownload 开始下载核心
func StartDownload(coreType, mcVersion, coreVersion string) (string, error) {
	if core.GetDownloadManager() == nil {
		core.InitDownloadManager()
	}
	
	return core.GetDownloadManager().StartDownload(coreType, mcVersion, coreVersion)
}

// CreateServer 创建新服务器
func CreateServer(serverConfig *config.MinecraftServer) (string, error) {
	// 生成服务器ID
	serverID := fmt.Sprintf("server_%d", time.Now().Unix())
	serverConfig.ID = serverID
	
	// 设置默认值
	if serverConfig.Memory == 0 {
		serverConfig.Memory = config.Get().Server.DefaultMemory
	}
	
	if serverConfig.Port == 0 {
		serverConfig.Port = 25565
	}
	
	if serverConfig.WorkDir == "" {
		serverConfig.WorkDir = filepath.Join("data", "servers", serverID)
	}
	
	// 设置Java路径
	if serverConfig.JavaPath == "" {
		cfg := config.Get()
		if cfg.Java.JavaPath != "" {
			serverConfig.JavaPath = cfg.Java.JavaPath
		} else {
			// 自动检测Java
			javaInfos, err := detectJavaEnvironment()
			if err != nil {
				return "", fmt.Errorf("failed to detect Java: %v", err)
			}
			if len(javaInfos) > 0 {
				serverConfig.JavaPath = javaInfos[0].Path
			}
		}
	}
	
	// 设置状态和时间
	serverConfig.Status = config.StatusStopped
	serverConfig.CreatedAt = time.Now()
	serverConfig.UpdatedAt = time.Now()
	
	// 创建工作目录
	if err := os.MkdirAll(serverConfig.WorkDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create work directory: %v", err)
	}
	
	// 保存服务器配置
	if err := config.GetServers().Add(serverConfig); err != nil {
		return "", fmt.Errorf("failed to save server config: %v", err)
	}
	
	return serverID, nil
}

// StartServer 启动服务器
func StartServer(serverID string) error {
	server, exists := config.GetServers().Get(serverID)
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}
	
	if server.Status == config.StatusRunning {
		return fmt.Errorf("server is already running")
	}
	
	// 检查核心文件是否存在
	if server.JarFile == "" {
		return fmt.Errorf("no jar file specified")
	}
	
	jarPath := filepath.Join(server.WorkDir, server.JarFile)
	if _, err := os.Stat(jarPath); os.IsNotExist(err) {
		return fmt.Errorf("jar file not found: %s", jarPath)
	}
	
	// 验证Java路径
	if _, err := validateJavaPath(server.JavaPath); err != nil {
		return fmt.Errorf("invalid Java path: %v", err)
	}
	
	// 更新服务器状态
	server.Status = config.StatusStarting
	now := time.Now()
	server.StartTime = &now
	config.GetServers().Update(server)
	
	// 启动服务器进程
	if err := startServerProcess(server); err != nil {
		server.Status = config.StatusCrashed
		config.GetServers().Update(server)
		return fmt.Errorf("failed to start server process: %v", err)
	}
	
	return nil
}

// StopServer 停止服务器
func StopServer(serverID string) error {
	server, exists := config.GetServers().Get(serverID)
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}
	
	if server.Status != config.StatusRunning {
		return fmt.Errorf("server is not running")
	}
	
	// 更新服务器状态
	server.Status = config.StatusStopping
	config.GetServers().Update(server)
	
	// 停止服务器进程
	if err := stopServerProcess(server); err != nil {
		return fmt.Errorf("failed to stop server process: %v", err)
	}
	
	return nil
}

// RestartServer 重启服务器
func RestartServer(serverID string) error {
	if err := StopServer(serverID); err != nil {
		return err
	}
	
	// 等待服务器完全停止
	time.Sleep(3 * time.Second)
	
	return StartServer(serverID)
}

// SendCommand 发送命令到服务器
func SendCommand(serverID, command string) error {
	server, exists := config.GetServers().Get(serverID)
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}
	
	if server.Status != config.StatusRunning {
		return fmt.Errorf("server is not running")
	}
	
	return sendCommandToProcess(server, command)
}

// DeleteServer 删除服务器
func DeleteServer(serverID string) error {
	server, exists := config.GetServers().Get(serverID)
	if !exists {
		return fmt.Errorf("server not found: %s", serverID)
	}
	
	// 如果服务器正在运行，先停止
	if server.Status == config.StatusRunning {
		if err := StopServer(serverID); err != nil {
			return fmt.Errorf("failed to stop server before deletion: %v", err)
		}
		
		// 等待服务器停止
		time.Sleep(3 * time.Second)
	}
	
	// 删除工作目录（可选）
	// 注意：这里可以添加用户确认机制
	
	// 从配置中删除服务器
	return config.GetServers().Delete(serverID)
}

// GetServerStatus 获取服务器状态
func GetServerStatus(serverID string) (interface{}, bool) {
	server, exists := config.GetServers().Get(serverID)
	if !exists {
		return nil, false
	}
	
	// 检查进程状态
	if server.Status == config.StatusRunning && server.PID > 0 {
		if !isProcessRunning(server.PID) {
			// 进程已停止，更新状态
			server.Status = config.StatusCrashed
			now := time.Now()
			server.StopTime = &now
			config.GetServers().Update(server)
		}
	}
	
	return map[string]interface{}{
		"id":          server.ID,
		"name":        server.Name,
		"status":      server.Status,
		"pid":         server.PID,
		"port":        server.Port,
		"memory":      server.Memory,
		"core_type":   server.CoreType,
		"mc_version":  server.MCVersion,
		"start_time":  server.StartTime,
		"stop_time":   server.StopTime,
		"auto_start":  server.AutoStart,
		"auto_restart": server.AutoRestart,
	}, true
}

// GetServerLogs 获取服务器日志
func GetServerLogs(serverID string, lines int) ([]string, error) {
	server, exists := config.GetServers().Get(serverID)
	if !exists {
		return nil, fmt.Errorf("server not found: %s", serverID)
	}
	
	logFile := filepath.Join(server.WorkDir, "logs", "latest.log")
	return readLogLines(logFile, lines)
}

// 辅助函数：检查字符串是否包含任何指定的子字符串
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if strings.Contains(s, substr) {
			return true
		}
	}
	return false
}
