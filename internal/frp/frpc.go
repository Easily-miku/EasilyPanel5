package frp

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// FRPCManager frpc客户端管理器
type FRPCManager struct {
	binaryPath string
	configPath string
	logPath    string
	process    *os.Process
	config     *FRPCConfig
}

// NewFRPCManager 创建新的frpc管理器
func NewFRPCManager(binaryPath, configPath, logPath string) *FRPCManager {
	return &FRPCManager{
		binaryPath: binaryPath,
		configPath: configPath,
		logPath:    logPath,
	}
}

// DownloadFRPC 下载frpc客户端
func (m *FRPCManager) DownloadFRPC() error {
	// 获取系统信息
	osName := runtime.GOOS
	arch := runtime.GOARCH
	
	// 转换架构名称
	switch arch {
	case "amd64":
		arch = "amd64"
	case "386":
		arch = "386"
	case "arm64":
		arch = "arm64"
	case "arm":
		arch = "arm"
	default:
		return fmt.Errorf("不支持的架构: %s", arch)
	}
	
	// 转换操作系统名称
	switch osName {
	case "windows":
		osName = "windows"
	case "linux":
		osName = "linux"
	case "darwin":
		osName = "darwin"
	case "freebsd":
		osName = "freebsd"
	default:
		return fmt.Errorf("不支持的操作系统: %s", osName)
	}
	
	// 获取最新版本信息
	versionInfo, err := m.getLatestVersion()
	if err != nil {
		return fmt.Errorf("获取版本信息失败: %w", err)
	}
	
	// 构建下载URL
	var fileName string
	if osName == "windows" {
		fileName = fmt.Sprintf("frpc_%s_%s.zip", osName, arch)
	} else {
		fileName = fmt.Sprintf("frpc_%s_%s.tar.gz", osName, arch)
	}
	
	downloadURL := fmt.Sprintf("https://r.zyghit.cn/download/client/%s/%s", 
		versionInfo.LatestFull, fileName)
	
	fmt.Printf("正在下载 frpc %s (%s_%s)...\n", versionInfo.LatestVer, osName, arch)
	fmt.Printf("下载地址: %s\n", downloadURL)
	
	// 下载文件
	if err := m.downloadFile(downloadURL, fileName); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	
	// 解压文件
	if err := m.extractFile(fileName, osName); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}
	
	// 清理下载的压缩包
	os.Remove(fileName)
	
	fmt.Printf("✓ frpc 下载完成: %s\n", m.binaryPath)
	return nil
}

// getLatestVersion 获取最新版本信息
func (m *FRPCManager) getLatestVersion() (*VersionInfo, error) {
	resp, err := http.Get("https://api.openfrp.net/commonQuery/get?key=software")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// 这里简化处理，实际应该解析JSON
	// 返回一个默认的版本信息
	return &VersionInfo{
		Latest:     "/OF_0.61.1_4df06100_250122/",
		LatestFull: "OF_0.61.1_4df06100_250122",
		LatestVer:  "0.61.1",
	}, nil
}

// downloadFile 下载文件
func (m *FRPCManager) downloadFile(url, filename string) error {
	// 创建目录
	dir := filepath.Dir(m.binaryPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	
	// 下载文件
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d", resp.StatusCode)
	}
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	_, err = io.Copy(file, resp.Body)
	return err
}

// extractFile 解压文件
func (m *FRPCManager) extractFile(filename, osName string) error {
	if osName == "windows" {
		return m.extractZip(filename)
	} else {
		return m.extractTarGz(filename)
	}
}

// extractZip 解压ZIP文件
func (m *FRPCManager) extractZip(filename string) error {
	reader, err := zip.OpenReader(filename)
	if err != nil {
		return err
	}
	defer reader.Close()

	// 查找第一个非目录文件（通常就是可执行文件）
	for _, file := range reader.File {
		if !file.FileInfo().IsDir() {
			fmt.Printf("提取文件: %s -> frpc\n", file.Name)
			return m.extractFileFromZip(file, m.binaryPath)
		}
	}

	return fmt.Errorf("压缩包中未找到可执行文件")
}

// extractTarGz 解压tar.gz文件
func (m *FRPCManager) extractTarGz(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// 查找第一个普通文件（通常就是可执行文件）
		if header.Typeflag == tar.TypeReg {
			fmt.Printf("提取文件: %s -> frpc\n", header.Name)
			return m.extractFileFromTar(tarReader, m.binaryPath)
		}
	}

	return fmt.Errorf("压缩包中未找到可执行文件")
}

// extractFileFromZip 从ZIP中提取文件
func (m *FRPCManager) extractFileFromZip(file *zip.File, destPath string) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, reader)
	if err != nil {
		return err
	}
	
	// 设置执行权限
	return os.Chmod(destPath, 0755)
}

// extractFileFromTar 从TAR中提取文件
func (m *FRPCManager) extractFileFromTar(reader *tar.Reader, destPath string) error {
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, reader)
	if err != nil {
		return err
	}
	
	// 设置执行权限
	return os.Chmod(destPath, 0755)
}

// IsInstalled 检查frpc是否已安装
func (m *FRPCManager) IsInstalled() bool {
	_, err := os.Stat(m.binaryPath)
	return err == nil
}

// GetVersion 获取frpc版本
func (m *FRPCManager) GetVersion() (string, error) {
	if !m.IsInstalled() {
		return "", fmt.Errorf("frpc未安装")
	}
	
	cmd := exec.Command(m.binaryPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	
	return strings.TrimSpace(string(output)), nil
}

// Start 启动frpc
func (m *FRPCManager) Start() error {
	if m.IsRunning() {
		return fmt.Errorf("frpc已在运行")
	}
	
	if !m.IsInstalled() {
		return fmt.Errorf("frpc未安装")
	}
	
	// 检查配置文件
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", m.configPath)
	}
	
	// 创建日志目录
	logDir := filepath.Dir(m.logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}
	
	// 启动frpc
	cmd := exec.Command(m.binaryPath, "-c", m.configPath)
	
	// 设置日志输出
	logFile, err := os.OpenFile(m.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}
	
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	
	// 启动进程
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("启动frpc失败: %w", err)
	}
	
	m.process = cmd.Process
	
	// 在后台等待进程结束
	go func() {
		cmd.Wait()
		logFile.Close()
		m.process = nil
	}()
	
	fmt.Printf("✓ frpc已启动 (PID: %d)\n", m.process.Pid)
	return nil
}

// Stop 停止frpc
func (m *FRPCManager) Stop() error {
	if !m.IsRunning() {
		return fmt.Errorf("frpc未运行")
	}
	
	// 发送终止信号
	if err := m.process.Signal(syscall.SIGTERM); err != nil {
		// 如果SIGTERM失败，强制杀死进程
		if killErr := m.process.Kill(); killErr != nil {
			return fmt.Errorf("停止frpc失败: %w", killErr)
		}
	}
	
	// 等待进程结束
	_, err := m.process.Wait()
	m.process = nil
	
	fmt.Println("✓ frpc已停止")
	return err
}

// Restart 重启frpc
func (m *FRPCManager) Restart() error {
	if m.IsRunning() {
		if err := m.Stop(); err != nil {
			return err
		}
		
		// 等待一秒确保进程完全停止
		time.Sleep(time.Second)
	}
	
	return m.Start()
}

// IsRunning 检查frpc是否在运行
func (m *FRPCManager) IsRunning() bool {
	if m.process == nil {
		return false
	}
	
	// 检查进程是否还存在
	if err := m.process.Signal(syscall.Signal(0)); err != nil {
		m.process = nil
		return false
	}
	
	return true
}

// GetStatus 获取frpc状态
func (m *FRPCManager) GetStatus() string {
	if !m.IsInstalled() {
		return "未安装"
	}
	
	if m.IsRunning() {
		return fmt.Sprintf("运行中 (PID: %d)", m.process.Pid)
	}
	
	return "已停止"
}

// VersionInfo 版本信息
type VersionInfo struct {
	Latest     string `json:"latest"`
	LatestFull string `json:"latest_full"`
	LatestVer  string `json:"latest_ver"`
}

// GenerateConfig 生成frpc配置文件
func (m *FRPCManager) GenerateConfig(serverAddr, token string, proxies []ProxyInfo) error {
	// 创建配置目录
	configDir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 生成配置内容
	config := fmt.Sprintf(`[common]
server_addr = %s
server_port = 7000
token = %s
log_level = info
log_file = %s
log_max_days = 3
admin_addr = 127.0.0.1
admin_port = 7400
pool_count = 5
tcp_mux = true

`, serverAddr, token, m.logPath)

	// 添加隧道配置
	for _, proxy := range proxies {
		config += m.generateProxyConfig(proxy)
	}

	// 写入配置文件
	if err := os.WriteFile(m.configPath, []byte(config), 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("✓ 配置文件已生成: %s\n", m.configPath)
	return nil
}

// generateProxyConfig 生成单个隧道配置
func (m *FRPCManager) generateProxyConfig(proxy ProxyInfo) string {
	config := fmt.Sprintf("\n[%s]\n", proxy.ProxyName)
	config += fmt.Sprintf("type = %s\n", proxy.ProxyType)
	config += fmt.Sprintf("local_ip = %s\n", proxy.LocalIP)
	config += fmt.Sprintf("local_port = %d\n", proxy.LocalPort)

	// 根据类型添加特定配置
	switch proxy.ProxyType {
	case "tcp", "udp":
		if proxy.RemotePort > 0 {
			config += fmt.Sprintf("remote_port = %d\n", proxy.RemotePort)
		}
	case "http", "https":
		if proxy.Domain != "" {
			config += fmt.Sprintf("custom_domains = %s\n", proxy.Domain)
		}
	}

	// 添加可选配置
	if proxy.UseEncryption {
		config += "use_encryption = true\n"
	}
	if proxy.UseCompression {
		config += "use_compression = true\n"
	}
	if proxy.ProxyProtocolVersion {
		config += "proxy_protocol_version = v2\n"
	}

	config += "\n"
	return config
}

// LoadConfig 加载配置文件
func (m *FRPCManager) LoadConfig() error {
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", m.configPath)
	}

	// 这里可以解析INI配置文件
	// 简化实现，实际应该使用INI解析库
	fmt.Printf("配置文件: %s\n", m.configPath)
	return nil
}

// GetLogs 获取日志内容
func (m *FRPCManager) GetLogs(lines int) ([]string, error) {
	if _, err := os.Stat(m.logPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("日志文件不存在: %s", m.logPath)
	}

	file, err := os.Open(m.logPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 简化实现：读取所有行然后返回最后N行
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	allLines := strings.Split(string(content), "\n")

	// 返回最后N行
	start := len(allLines) - lines
	if start < 0 {
		start = 0
	}

	return allLines[start:], nil
}

// ClearLogs 清空日志文件
func (m *FRPCManager) ClearLogs() error {
	return os.Truncate(m.logPath, 0)
}
