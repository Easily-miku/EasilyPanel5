package download

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadManager 下载管理器
type DownloadManager struct {
	fastMirror *FastMirrorClient
	downloader *Downloader
	dataDir    string
}

// NewDownloadManager 创建新的下载管理器
func NewDownloadManager(dataDir string) *DownloadManager {
	return &DownloadManager{
		fastMirror: NewFastMirrorClient(),
		downloader: NewDownloader(),
		dataDir:    dataDir,
	}
}

// GetDownloadDir 获取下载目录
func (dm *DownloadManager) GetDownloadDir() string {
	return filepath.Join(dm.dataDir, "downloads")
}

// ListAvailableServers 列出可用的服务端
func (dm *DownloadManager) ListAvailableServers() ([]ServerInfo, error) {
	return dm.fastMirror.GetServerList()
}

// SearchServers 搜索服务端
func (dm *DownloadManager) SearchServers(keyword string) ([]ServerInfo, error) {
	return dm.fastMirror.SearchServers(keyword)
}

// GetServerInfo 获取服务端详细信息
func (dm *DownloadManager) GetServerInfo(name string) (*ProjectInfo, error) {
	return dm.fastMirror.GetProjectInfo(name)
}

// ListVersions 列出服务端支持的MC版本
func (dm *DownloadManager) ListVersions(serverName string) ([]string, error) {
	info, err := dm.fastMirror.GetProjectInfo(serverName)
	if err != nil {
		return nil, err
	}
	return info.MCVersions, nil
}

// ListBuilds 列出构建版本
func (dm *DownloadManager) ListBuilds(serverName, mcVersion string, limit int) ([]BuildInfo, error) {
	if limit <= 0 {
		limit = 10 // 默认显示10个
	}
	
	builds, err := dm.fastMirror.GetBuilds(serverName, mcVersion, 0, limit)
	if err != nil {
		return nil, err
	}
	
	return builds.Builds, nil
}

// GetLatestBuild 获取最新构建
func (dm *DownloadManager) GetLatestBuild(serverName, mcVersion string) (*BuildInfo, error) {
	return dm.fastMirror.GetLatestBuild(serverName, mcVersion)
}

// DownloadServer 下载服务端
func (dm *DownloadManager) DownloadServer(serverName, mcVersion, coreVersion string, showProgress bool) (string, error) {
	// 获取核心信息
	coreInfo, err := dm.fastMirror.GetCoreInfo(serverName, mcVersion, coreVersion)
	if err != nil {
		return "", fmt.Errorf("获取核心信息失败: %w", err)
	}
	
	// 确定下载路径
	downloadDir := dm.GetDownloadDir()
	fileName := coreInfo.Filename
	if fileName == "" {
		fileName = fmt.Sprintf("%s-%s-%s.jar", serverName, mcVersion, coreVersion)
	}
	
	filePath := filepath.Join(downloadDir, fileName)
	
	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		// 文件存在，验证校验和
		if err := dm.downloader.VerifyFile(filePath, coreInfo.SHA1); err == nil {
			fmt.Printf("文件已存在且校验通过: %s\n", filePath)
			return filePath, nil
		} else {
			fmt.Printf("文件校验失败，重新下载: %v\n", err)
			os.Remove(filePath) // 删除损坏的文件
		}
	}
	
	// 创建下载目录
	if err := os.MkdirAll(downloadDir, 0755); err != nil {
		return "", fmt.Errorf("创建下载目录失败: %w", err)
	}
	
	fmt.Printf("开始下载: %s\n", fileName)
	fmt.Printf("下载地址: %s\n", coreInfo.DownloadURL)
	
	// 设置进度回调
	var progress *DownloadProgress
	var callback ProgressCallback
	
	if showProgress {
		progress = NewDownloadProgress()
		callback = func(downloaded, total int64, percent float64) {
			progress.Update(downloaded, total)
			fmt.Printf("\r下载进度: %s", progress.String())
		}
	}
	
	// 开始下载
	startTime := time.Now()
	err = dm.downloader.DownloadWithRetry(coreInfo.DownloadURL, filePath, 3, callback)
	if err != nil {
		return "", fmt.Errorf("下载失败: %w", err)
	}
	
	if showProgress {
		fmt.Println() // 换行
	}
	
	// 验证文件
	if coreInfo.SHA1 != "" {
		fmt.Print("正在验证文件...")
		if err := dm.downloader.VerifyFile(filePath, coreInfo.SHA1); err != nil {
			os.Remove(filePath) // 删除损坏的文件
			return "", fmt.Errorf("文件校验失败: %w", err)
		}
		fmt.Println(" 校验通过")
	}
	
	duration := time.Since(startTime)
	fileInfo, _ := os.Stat(filePath)
	fmt.Printf("下载完成: %s (耗时: %v, 大小: %s)\n", 
		fileName, duration.Round(time.Second), FormatBytes(fileInfo.Size()))
	
	return filePath, nil
}

// DownloadLatest 下载最新版本
func (dm *DownloadManager) DownloadLatest(serverName, mcVersion string, showProgress bool) (string, error) {
	// 获取最新构建
	latest, err := dm.GetLatestBuild(serverName, mcVersion)
	if err != nil {
		return "", err
	}
	
	fmt.Printf("找到最新版本: %s %s %s\n", latest.Name, latest.MCVersion, latest.CoreVersion)
	
	return dm.DownloadServer(serverName, mcVersion, latest.CoreVersion, showProgress)
}

// ListDownloadedFiles 列出已下载的文件
func (dm *DownloadManager) ListDownloadedFiles() ([]string, error) {
	downloadDir := dm.GetDownloadDir()
	
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		return []string{}, nil
	}
	
	entries, err := os.ReadDir(downloadDir)
	if err != nil {
		return nil, fmt.Errorf("读取下载目录失败: %w", err)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".jar") {
			files = append(files, entry.Name())
		}
	}
	
	return files, nil
}

// GetDownloadedFilePath 获取已下载文件的完整路径
func (dm *DownloadManager) GetDownloadedFilePath(filename string) string {
	return filepath.Join(dm.GetDownloadDir(), filename)
}

// DeleteDownloadedFile 删除已下载的文件
func (dm *DownloadManager) DeleteDownloadedFile(filename string) error {
	filePath := dm.GetDownloadedFilePath(filename)
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filename)
	}
	
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %w", err)
	}
	
	return nil
}

// GetRecommendedServers 获取推荐的服务端
func (dm *DownloadManager) GetRecommendedServers() ([]ServerInfo, error) {
	return dm.fastMirror.GetRecommendedServers()
}

// GetServersByTag 根据标签获取服务端
func (dm *DownloadManager) GetServersByTag(tag string) ([]ServerInfo, error) {
	return dm.fastMirror.GetServersByTag(tag)
}

// PrintServerList 打印服务端列表
func (dm *DownloadManager) PrintServerList(servers []ServerInfo) {
	if len(servers) == 0 {
		fmt.Println("未找到任何服务端")
		return
	}
	
	fmt.Printf("找到 %d 个服务端:\n\n", len(servers))
	fmt.Println("名称           | 标签    | 推荐")
	fmt.Println("---------------|---------|----")
	
	for _, server := range servers {
		recommend := "否"
		if server.Recommend {
			recommend = "是"
		}
		
		fmt.Printf("%-14s | %-7s | %s\n", server.Name, server.Tag, recommend)
	}
}

// PrintBuildList 打印构建列表
func (dm *DownloadManager) PrintBuildList(builds []BuildInfo) {
	if len(builds) == 0 {
		fmt.Println("未找到任何构建版本")
		return
	}
	
	fmt.Printf("找到 %d 个构建版本:\n\n", len(builds))
	fmt.Println("核心版本       | MC版本  | 更新时间")
	fmt.Println("---------------|---------|--------------------")
	
	for _, build := range builds {
		// 解析时间
		updateTime := build.UpdateTime
		if t, err := time.Parse("2006-01-02T15:04:05", build.UpdateTime); err == nil {
			updateTime = t.Format("2006-01-02 15:04:05")
		}
		
		fmt.Printf("%-14s | %-7s | %s\n", build.CoreVersion, build.MCVersion, updateTime)
	}
}

// PrintDownloadedFiles 打印已下载的文件
func (dm *DownloadManager) PrintDownloadedFiles() error {
	files, err := dm.ListDownloadedFiles()
	if err != nil {
		return err
	}
	
	if len(files) == 0 {
		fmt.Println("未找到已下载的文件")
		return nil
	}
	
	fmt.Printf("已下载的文件 (%d 个):\n\n", len(files))
	
	downloadDir := dm.GetDownloadDir()
	for i, file := range files {
		filePath := filepath.Join(downloadDir, file)
		if info, err := os.Stat(filePath); err == nil {
			fmt.Printf("%d. %s (%s)\n", i+1, file, FormatBytes(info.Size()))
		} else {
			fmt.Printf("%d. %s\n", i+1, file)
		}
	}
	
	return nil
}


