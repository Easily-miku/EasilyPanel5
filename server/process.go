package server

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"easilypanel5/config"
	"easilypanel5/utils"
)

// ServerProcess 服务器进程信息
type ServerProcess struct {
	Server      *config.MinecraftServer
	Cmd         *exec.Cmd
	Stdin       io.WriteCloser
	Stdout      io.ReadCloser
	Stderr      io.ReadCloser
	LogFile     *os.File
	lastLogTime []time.Time // 用于日志限流
	mutex       sync.RWMutex
}

var (
	runningProcesses = make(map[string]*ServerProcess)
	processMutex     sync.RWMutex
)

// startServerProcess 启动服务器进程
func startServerProcess(server *config.MinecraftServer) (*ServerProcess, error) {
	processMutex.Lock()
	defer processMutex.Unlock()

	// 检查是否已经在运行
	if _, exists := runningProcesses[server.ID]; exists {
		return nil, fmt.Errorf("server process already running")
	}

	// 构建启动命令
	args := buildJavaArgs(server)
	cmd := exec.Command(server.JavaPath, args...)
	cmd.Dir = server.WorkDir

	// 设置环境变量
	cmd.Env = os.Environ()

	// 创建管道
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// 创建日志文件
	logDir := filepath.Join(server.WorkDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, "latest.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to create log file: %v", err)
	}

	// 启动进程
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		logFile.Close()
		return nil, fmt.Errorf("failed to start process: %v", err)
	}

	// 更新服务器信息
	server.PID = cmd.Process.Pid
	server.Status = config.StatusRunning
	config.GetServers().Update(server)

	// 创建进程对象
	process := &ServerProcess{
		Server:  server,
		Cmd:     cmd,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
		LogFile: logFile,
	}

	runningProcesses[server.ID] = process

	// 启动日志处理协程
	go process.handleLogs()
	go process.monitorProcess()

	return process, nil
}

// stopServerProcess 停止服务器进程
func stopServerProcess(server *config.MinecraftServer) error {
	processMutex.Lock()
	defer processMutex.Unlock()

	process, exists := runningProcesses[server.ID]
	if !exists {
		return fmt.Errorf("server process not found")
	}

	// 发送stop命令
	if err := process.sendCommand("stop"); err != nil {
		// 如果发送stop命令失败，强制终止进程
		return process.forceKill()
	}

	// 等待进程优雅退出
	done := make(chan error, 1)
	go func() {
		done <- process.Cmd.Wait()
	}()

	select {
	case <-done:
		// 进程已退出
	case <-time.After(30 * time.Second):
		// 超时，强制终止
		process.forceKill()
	}

	// 清理资源
	process.cleanup()
	delete(runningProcesses, server.ID)

	// 更新服务器状态
	server.Status = config.StatusStopped
	server.PID = 0
	now := time.Now()
	server.StopTime = &now
	config.GetServers().Update(server)

	return nil
}

// sendCommandToProcess 向进程发送命令
func sendCommandToProcess(server *config.MinecraftServer, command string) error {
	processMutex.RLock()
	process, exists := runningProcesses[server.ID]
	processMutex.RUnlock()

	if !exists {
		return fmt.Errorf("server process not found")
	}

	return process.sendCommand(command)
}


// buildJavaArgs 构建Java启动参数
func buildJavaArgs(server *config.MinecraftServer) []string {
	args := []string{}

	// 内存参数
	if server.Memory > 0 {
		args = append(args, fmt.Sprintf("-Xms%dM", server.Memory/2))
		args = append(args, fmt.Sprintf("-Xmx%dM", server.Memory))
	}

	// 默认JVM参数
	cfg := config.Get()
	args = append(args, cfg.Java.DefaultArgs...)

	// 自定义参数
	if len(server.JavaArgs) > 0 {
		args = append(args, server.JavaArgs...)
	}

	// JAR文件参数
	args = append(args, "-jar", server.JarFile)

	// Minecraft服务器参数
	args = append(args, "--nogui")

	return args
}

// handleLogs 处理日志输出
func (p *ServerProcess) handleLogs() {
	// 处理stdout
	go func() {
		scanner := bufio.NewScanner(p.Stdout)
		for scanner.Scan() {
			line := scanner.Text()
			p.writeLog(line, "stdout")
			p.broadcastLog(line, "INFO")
		}
	}()

	// 处理stderr
	go func() {
		scanner := bufio.NewScanner(p.Stderr)
		for scanner.Scan() {
			line := scanner.Text()
			p.writeLog(line, "stderr")
			p.broadcastLog(line, "ERROR")
		}
	}()
}

// writeLog 写入日志文件
func (p *ServerProcess) writeLog(line, source string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.LogFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("[%s] [%s] %s\n", timestamp, source, line)

		// 检查是否需要轮转日志
		if p.shouldRotateLog() {
			p.rotateLog()
		}

		p.LogFile.WriteString(logLine)
		p.LogFile.Sync()
	}
}

// broadcastLog 广播日志到WebSocket客户端（带限流）
func (p *ServerProcess) broadcastLog(line, level string) {
	// 简单的限流：检查最近1秒内的日志数量
	now := time.Now()
	p.lastLogTime = append(p.lastLogTime, now)

	// 清理1秒前的记录
	cutoff := now.Add(-time.Second)
	for len(p.lastLogTime) > 0 && p.lastLogTime[0].Before(cutoff) {
		p.lastLogTime = p.lastLogTime[1:]
	}

	// 如果1秒内日志超过100条，跳过广播
	if len(p.lastLogTime) > 100 {
		return
	}

	// 解析日志获取更多信息
	parsed := utils.ParseMinecraftLog(line)

	utils.EmitEvent("log_message", p.Server.ID, map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
		"level":     parsed["level"],
		"message":   parsed["message"],
		"raw":       line,
		"thread":    parsed["thread"],
		"logger":    parsed["logger"],
	})
}

// monitorProcess 监控进程状态
func (p *ServerProcess) monitorProcess() {
	err := p.Cmd.Wait()

	processMutex.Lock()
	defer processMutex.Unlock()

	// 检查进程是否还在运行列表中
	if _, exists := runningProcesses[p.Server.ID]; !exists {
		return // 进程已被正常停止
	}

	// 进程意外退出
	p.cleanup()
	delete(runningProcesses, p.Server.ID)

	// 更新服务器状态
	p.Server.Status = config.StatusCrashed
	p.Server.PID = 0
	now := time.Now()
	p.Server.StopTime = &now
	config.GetServers().Update(p.Server)

	// 广播状态变化
	utils.EmitEvent("server_status", p.Server.ID, map[string]interface{}{
		"status": config.StatusCrashed,
		"pid":    0,
		"error":  fmt.Sprintf("Process exited with error: %v", err),
	})

	// 如果启用了自动重启
	if p.Server.AutoRestart {
		time.Sleep(5 * time.Second) // 等待5秒后重启
		if _, err := startServerProcess(p.Server); err != nil {
			// 重启失败，记录错误
			utils.EmitEvent("log_message", p.Server.ID, map[string]interface{}{
				"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
				"level":     "ERROR",
				"message":   fmt.Sprintf("Auto-restart failed: %v", err),
				"raw":       fmt.Sprintf("Auto-restart failed: %v", err),
			})
		}
	}
}

// sendCommand 发送命令到进程
func (p *ServerProcess) sendCommand(command string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.Stdin == nil {
		return fmt.Errorf("stdin not available")
	}

	_, err := p.Stdin.Write([]byte(command + "\n"))
	return err
}

// forceKill 强制终止进程
func (p *ServerProcess) forceKill() error {
	if p.Cmd != nil && p.Cmd.Process != nil {
		return p.Cmd.Process.Kill()
	}
	return nil
}

// cleanup 清理资源
func (p *ServerProcess) cleanup() {
	if p.Stdin != nil {
		p.Stdin.Close()
	}
	if p.Stdout != nil {
		p.Stdout.Close()
	}
	if p.Stderr != nil {
		p.Stderr.Close()
	}
	if p.LogFile != nil {
		p.LogFile.Close()
	}
}

// readLogLines 读取日志文件的最后几行（高效实现）
func readLogLines(filename string, lines int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 获取文件大小
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	fileSize := stat.Size()

	if fileSize == 0 {
		return []string{}, nil
	}

	// 从文件末尾开始读取
	const bufferSize = 4096
	var result []string
	var buffer []byte
	var offset int64 = fileSize
	var lineCount int

	for offset > 0 && lineCount < lines {
		// 计算读取大小
		readSize := int64(bufferSize)
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize

		// 读取数据块
		chunk := make([]byte, readSize)
		_, err := file.ReadAt(chunk, offset)
		if err != nil {
			return nil, err
		}

		// 将数据块添加到缓冲区前面
		buffer = append(chunk, buffer...)

		// 分割行
		splitLines := strings.Split(string(buffer), "\n")

		// 如果不是从文件开头开始读取，第一行可能是不完整的
		if offset > 0 && len(splitLines) > 0 {
			// 保留第一行作为下次读取的一部分
			buffer = []byte(splitLines[0])
			splitLines = splitLines[1:]
		} else {
			buffer = []byte{}
		}

		// 从后往前添加完整的行
		for i := len(splitLines) - 1; i >= 0 && lineCount < lines; i-- {
			if strings.TrimSpace(splitLines[i]) != "" {
				result = append([]string{splitLines[i]}, result...)
				lineCount++
			}
		}
	}

	// 限制返回的行数
	if len(result) > lines {
		result = result[len(result)-lines:]
	}

	return result, nil
}

// parseLogLine 解析日志行，提取时间戳和级别
func parseLogLine(line string) map[string]interface{} {
	// 使用utils包中的解析函数
	return utils.ParseMinecraftLog(line)
}

// shouldRotateLog 检查是否需要轮转日志
func (p *ServerProcess) shouldRotateLog() bool {
	if p.LogFile == nil {
		return false
	}

	// 获取配置
	cfg := config.Get()
	if !cfg.Daemon.LogRotation {
		return false
	}

	// 检查文件大小
	stat, err := p.LogFile.Stat()
	if err != nil {
		return false
	}

	return stat.Size() > cfg.Daemon.MaxLogSize
}

// rotateLog 轮转日志文件
func (p *ServerProcess) rotateLog() {
	if p.LogFile == nil {
		return
	}

	cfg := config.Get()
	logDir := filepath.Join(p.Server.WorkDir, "logs")

	// 关闭当前日志文件
	p.LogFile.Close()

	// 重命名当前日志文件
	currentLogPath := filepath.Join(logDir, "latest.log")
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	archivedLogPath := filepath.Join(logDir, fmt.Sprintf("server-%s.log", timestamp))

	if err := os.Rename(currentLogPath, archivedLogPath); err != nil {
		fmt.Printf("Failed to rotate log file: %v\n", err)
	}

	// 创建新的日志文件
	newLogFile, err := os.OpenFile(
		currentLogPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		fmt.Printf("Failed to create new log file: %v\n", err)
		return
	}

	p.LogFile = newLogFile

	// 清理旧日志文件
	go p.cleanupOldLogs(logDir, cfg.Daemon.MaxLogFiles)
}

// cleanupOldLogs 清理旧的日志文件
func (p *ServerProcess) cleanupOldLogs(logDir string, maxFiles int) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return
	}

	// 过滤出日志文件并按修改时间排序
	var logFiles []os.FileInfo
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "server-") && strings.HasSuffix(entry.Name(), ".log") {
			info, err := entry.Info()
			if err == nil {
				logFiles = append(logFiles, info)
			}
		}
	}

	// 按修改时间排序（最新的在前）
	sort.Slice(logFiles, func(i, j int) bool {
		return logFiles[i].ModTime().After(logFiles[j].ModTime())
	})

	// 删除超出限制的文件
	for i := maxFiles; i < len(logFiles); i++ {
		filePath := filepath.Join(logDir, logFiles[i].Name())
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Failed to remove old log file %s: %v\n", filePath, err)
		}
	}
}
