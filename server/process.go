package server

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"easilypanel5/config"
	"easilypanel5/utils"
)

// ServerProcess 服务器进程信息
type ServerProcess struct {
	Server  *config.MinecraftServer
	Cmd     *exec.Cmd
	Stdin   io.WriteCloser
	Stdout  io.ReadCloser
	Stderr  io.ReadCloser
	LogFile *os.File
	mutex   sync.RWMutex
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
		p.LogFile.WriteString(logLine)
		p.LogFile.Sync()
	}
}

// broadcastLog 广播日志到WebSocket客户端
func (p *ServerProcess) broadcastLog(line, level string) {
	utils.EmitEvent("log_message", p.Server.ID, map[string]interface{}{
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
		"level":     level,
		"message":   line,
		"raw":       line,
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

// readLogLines 读取日志文件的最后几行
func readLogLines(filename string, lines int) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 简单实现：读取所有行，然后返回最后几行
	var allLines []string
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// 返回最后几行
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}

	return allLines[start:], nil
}
