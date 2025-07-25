package instance

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProcessManager 进程管理器
type ProcessManager struct {
	dataDir string
	manager *Manager
}

// NewProcessManager 创建新的进程管理器
func NewProcessManager(dataDir string) *ProcessManager {
	return &ProcessManager{
		dataDir: dataDir,
		manager: NewManager(dataDir),
	}
}

// StartInstance 启动实例
func (pm *ProcessManager) StartInstance(name string) error {
	// 获取实例
	instance, err := pm.manager.GetInstance(name)
	if err != nil {
		return err
	}
	
	// 检查实例状态
	if instance.IsRunning() {
		return fmt.Errorf("实例 '%s' 已在运行中", name)
	}
	
	// 更新状态为启动中
	instance.UpdateStatus(StatusStarting)
	if err := pm.manager.UpdateInstance(instance); err != nil {
		return fmt.Errorf("更新实例状态失败: %w", err)
	}
	
	// 获取启动命令
	command, args, err := instance.GetStartCommand()
	if err != nil {
		instance.UpdateStatus(StatusError)
		pm.manager.UpdateInstance(instance)
		return fmt.Errorf("获取启动命令失败: %w", err)
	}
	
	// 设置工作目录
	workDir := instance.GetWorkDir(pm.dataDir)
	
	// 创建命令
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir
	
	// 设置环境变量
	cmd.Env = os.Environ()
	
	// 创建日志文件
	logFile := filepath.Join(workDir, "server.log")
	logWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		instance.UpdateStatus(StatusError)
		pm.manager.UpdateInstance(instance)
		return fmt.Errorf("创建日志文件失败: %w", err)
	}
	
	// 设置输出重定向
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	
	// 启动进程
	if err := cmd.Start(); err != nil {
		logWriter.Close()
		instance.UpdateStatus(StatusError)
		pm.manager.UpdateInstance(instance)
		return fmt.Errorf("启动进程失败: %w", err)
	}
	
	// 更新实例信息
	instance.SetPID(cmd.Process.Pid)
	instance.UpdateStatus(StatusRunning)
	if err := pm.manager.UpdateInstance(instance); err != nil {
		// 如果更新失败，尝试停止进程
		cmd.Process.Kill()
		logWriter.Close()
		return fmt.Errorf("更新实例信息失败: %w", err)
	}
	
	// 启动监控协程
	go pm.monitorProcess(instance, cmd, logWriter)
	
	fmt.Printf("实例 '%s' 启动成功 (PID: %d)\n", name, cmd.Process.Pid)
	return nil
}

// StopInstance 停止实例
func (pm *ProcessManager) StopInstance(name string) error {
	// 获取实例
	instance, err := pm.manager.GetInstance(name)
	if err != nil {
		return err
	}
	
	// 检查实例状态
	if !instance.IsRunning() {
		return fmt.Errorf("实例 '%s' 未在运行", name)
	}
	
	// 更新状态为停止中
	instance.UpdateStatus(StatusStopping)
	if err := pm.manager.UpdateInstance(instance); err != nil {
		return fmt.Errorf("更新实例状态失败: %w", err)
	}
	
	// 查找进程
	if instance.PID <= 0 {
		instance.UpdateStatus(StatusStopped)
		pm.manager.UpdateInstance(instance)
		return fmt.Errorf("实例 '%s' 的PID无效", name)
	}
	
	// 尝试优雅停止
	if err := pm.gracefulStop(instance.PID); err != nil {
		// 如果优雅停止失败，强制停止
		if err := pm.forceStop(instance.PID); err != nil {
			instance.UpdateStatus(StatusError)
			pm.manager.UpdateInstance(instance)
			return fmt.Errorf("停止进程失败: %w", err)
		}
	}
	
	// 更新实例状态
	instance.UpdateStatus(StatusStopped)
	if err := pm.manager.UpdateInstance(instance); err != nil {
		return fmt.Errorf("更新实例状态失败: %w", err)
	}
	
	fmt.Printf("实例 '%s' 已停止\n", name)
	return nil
}

// RestartInstance 重启实例
func (pm *ProcessManager) RestartInstance(name string) error {
	// 先停止
	if err := pm.StopInstance(name); err != nil {
		// 如果停止失败但实例确实没在运行，继续启动
		instance, getErr := pm.manager.GetInstance(name)
		if getErr != nil || instance.IsRunning() {
			return fmt.Errorf("停止实例失败: %w", err)
		}
	}
	
	// 等待一段时间确保进程完全停止
	time.Sleep(2 * time.Second)
	
	// 再启动
	return pm.StartInstance(name)
}

// gracefulStop 优雅停止进程
func (pm *ProcessManager) gracefulStop(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	
	// 发送SIGTERM信号
	if runtime.GOOS == "windows" {
		// Windows下使用taskkill
		cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	} else {
		// Unix系统使用SIGTERM
		if err := process.Signal(syscall.SIGTERM); err != nil {
			return err
		}
		
		// 等待进程退出
		done := make(chan error, 1)
		go func() {
			_, err := process.Wait()
			done <- err
		}()
		
		select {
		case <-time.After(10 * time.Second):
			// 超时，返回错误以便强制停止
			return fmt.Errorf("优雅停止超时")
		case err := <-done:
			return err
		}
	}
}

// forceStop 强制停止进程
func (pm *ProcessManager) forceStop(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	
	if runtime.GOOS == "windows" {
		// Windows下使用taskkill /F
		cmd := exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid))
		return cmd.Run()
	} else {
		// Unix系统使用SIGKILL
		return process.Signal(syscall.SIGKILL)
	}
}

// monitorProcess 监控进程状态
func (pm *ProcessManager) monitorProcess(instance *Instance, cmd *exec.Cmd, logWriter io.WriteCloser) {
	defer logWriter.Close()
	
	// 等待进程结束
	err := cmd.Wait()
	
	// 重新加载实例以获取最新状态
	currentInstance, loadErr := pm.manager.GetInstance(instance.Name)
	if loadErr != nil {
		fmt.Printf("警告: 重新加载实例 '%s' 失败: %v\n", instance.Name, loadErr)
		return
	}
	
	// 如果实例状态不是停止中，说明是异常退出
	if currentInstance.Status != StatusStopping {
		if err != nil {
			fmt.Printf("实例 '%s' 异常退出: %v\n", instance.Name, err)
			currentInstance.UpdateStatus(StatusError)
		} else {
			fmt.Printf("实例 '%s' 正常退出\n", instance.Name)
			currentInstance.UpdateStatus(StatusStopped)
		}
		
		// 如果启用了自动重启且不是手动停止
		if currentInstance.AutoRestart && currentInstance.Status != StatusStopping {
			fmt.Printf("实例 '%s' 启用了自动重启，5秒后重新启动...\n", instance.Name)
			time.Sleep(5 * time.Second)
			
			if restartErr := pm.StartInstance(instance.Name); restartErr != nil {
				fmt.Printf("自动重启实例 '%s' 失败: %v\n", instance.Name, restartErr)
			}
			return
		}
	} else {
		// 正常停止
		currentInstance.UpdateStatus(StatusStopped)
	}
	
	// 保存状态
	pm.manager.UpdateInstance(currentInstance)
}

// IsProcessRunning 检查进程是否正在运行
func (pm *ProcessManager) IsProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	
	if runtime.GOOS == "windows" {
		// Windows下检查进程状态
		cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid))
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), strconv.Itoa(pid))
	} else {
		// Unix系统发送信号0检查进程是否存在
		err := process.Signal(syscall.Signal(0))
		return err == nil
	}
}

// SendCommand 向实例发送命令
func (pm *ProcessManager) SendCommand(name, command string) error {
	instance, err := pm.manager.GetInstance(name)
	if err != nil {
		return err
	}
	
	if !instance.IsRunning() {
		return fmt.Errorf("实例 '%s' 未在运行", name)
	}
	
	// 这里需要实现向进程的stdin发送命令
	// 由于我们当前的实现没有保持stdin的引用，这个功能需要重构
	// 暂时返回未实现错误
	return fmt.Errorf("发送命令功能暂未实现")
}

// GetInstanceLogs 获取实例日志
func (pm *ProcessManager) GetInstanceLogs(name string, lines int) ([]string, error) {
	instance, err := pm.manager.GetInstance(name)
	if err != nil {
		return nil, err
	}
	
	logFile := filepath.Join(instance.GetWorkDir(pm.dataDir), "server.log")
	
	file, err := os.Open(logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	defer file.Close()
	
	var logLines []string
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		logLines = append(logLines, scanner.Text())
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取日志文件失败: %w", err)
	}
	
	// 返回最后N行
	if lines > 0 && len(logLines) > lines {
		return logLines[len(logLines)-lines:], nil
	}
	
	return logLines, nil
}
