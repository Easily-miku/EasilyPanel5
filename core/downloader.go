package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"easilypanel5/config"
)

// FastMirrorAPI FastMirror API客户端
type FastMirrorAPI struct {
	baseURL string
	client  *http.Client
}

// NewFastMirrorAPI 创建FastMirror API客户端
func NewFastMirrorAPI() *FastMirrorAPI {
	return &FastMirrorAPI{
		baseURL: config.Get().Download.FastMirrorAPI,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FastMirrorCore FastMirror核心信息
type FastMirrorCore struct {
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	Recommend bool   `json:"recommend"`
}

// FastMirrorProject FastMirror项目信息
type FastMirrorProject struct {
	Name       string   `json:"name"`
	Tag        string   `json:"tag"`
	Homepage   string   `json:"homepage"`
	MCVersions []string `json:"mc_versions"`
}

// FastMirrorBuilds FastMirror构建信息
type FastMirrorBuilds struct {
	Builds []FastMirrorBuild `json:"builds"`
	Offset int               `json:"offset"`
	Limit  int               `json:"limit"`
	Count  int               `json:"count"`
}

// FastMirrorBuild FastMirror构建详情
type FastMirrorBuild struct {
	Name        string `json:"name"`
	MCVersion   string `json:"mc_version"`
	CoreVersion string `json:"core_version"`
	UpdateTime  string `json:"update_time"`
	SHA1        string `json:"sha1"`
}

// FastMirrorMetadata FastMirror元数据
type FastMirrorMetadata struct {
	Name        string `json:"name"`
	MCVersion   string `json:"mc_version"`
	CoreVersion string `json:"core_version"`
	UpdateTime  string `json:"update_time"`
	SHA1        string `json:"sha1"`
	Filename    string `json:"filename"`
	DownloadURL string `json:"download_url"`
}

// FastMirrorResponse FastMirror API响应
type FastMirrorResponse struct {
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
}

// GetCoresList 获取核心列表
func (api *FastMirrorAPI) GetCoresList() ([]FastMirrorCore, error) {
	url := api.baseURL
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch cores list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response FastMirrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	// 转换数据
	var cores []FastMirrorCore
	if data, err := json.Marshal(response.Data); err == nil {
		json.Unmarshal(data, &cores)
	}

	return cores, nil
}

// GetProjectInfo 获取项目信息
func (api *FastMirrorAPI) GetProjectInfo(name string) (*FastMirrorProject, error) {
	url := fmt.Sprintf("%s/%s", api.baseURL, name)
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response FastMirrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	var project FastMirrorProject
	if data, err := json.Marshal(response.Data); err == nil {
		json.Unmarshal(data, &project)
	}

	return &project, nil
}

// GetBuilds 获取构建列表
func (api *FastMirrorAPI) GetBuilds(name, mcVersion string, offset, count int) (*FastMirrorBuilds, error) {
	url := fmt.Sprintf("%s/%s/%s?offset=%d&count=%d", api.baseURL, name, mcVersion, offset, count)
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch builds: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response FastMirrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	var builds FastMirrorBuilds
	if data, err := json.Marshal(response.Data); err == nil {
		json.Unmarshal(data, &builds)
	}

	return &builds, nil
}

// GetMetadata 获取核心元数据
func (api *FastMirrorAPI) GetMetadata(name, mcVersion, coreVersion string) (*FastMirrorMetadata, error) {
	url := fmt.Sprintf("%s/%s/%s/%s", api.baseURL, name, mcVersion, coreVersion)
	resp, err := api.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	var response FastMirrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	if !response.Success {
		return nil, fmt.Errorf("API error: %s", response.Message)
	}

	var metadata FastMirrorMetadata
	if data, err := json.Marshal(response.Data); err == nil {
		json.Unmarshal(data, &metadata)
	}

	return &metadata, nil
}

// DownloadManager 下载管理器
type DownloadManager struct {
	tasks  map[string]*config.DownloadTask
	mutex  sync.RWMutex
	api    *FastMirrorAPI
}

var downloadManager *DownloadManager

// InitDownloadManager 初始化下载管理器
func InitDownloadManager() {
	downloadManager = &DownloadManager{
		tasks: make(map[string]*config.DownloadTask),
		api:   NewFastMirrorAPI(),
	}
}

// GetDownloadManager 获取下载管理器
func GetDownloadManager() *DownloadManager {
	return downloadManager
}

// StartDownload 开始下载
func (dm *DownloadManager) StartDownload(coreType, mcVersion, coreVersion string) (string, error) {
	// 获取核心元数据
	metadata, err := dm.api.GetMetadata(coreType, mcVersion, coreVersion)
	if err != nil {
		return "", fmt.Errorf("failed to get metadata: %v", err)
	}

	// 创建下载任务
	taskID := fmt.Sprintf("%s-%s-%s-%d", coreType, mcVersion, coreVersion, time.Now().Unix())
	task := &config.DownloadTask{
		ID: taskID,
		CoreInfo: config.CoreInfo{
			Name:        metadata.Name,
			Type:        coreType,
			MCVersion:   metadata.MCVersion,
			CoreVersion: metadata.CoreVersion,
			FileName:    metadata.Filename,
			DownloadURL: metadata.DownloadURL,
			SHA1:        metadata.SHA1,
			UpdateTime:  metadata.UpdateTime,
		},
		Status:    "pending",
		Progress:  0,
		StartTime: time.Now(),
	}

	dm.mutex.Lock()
	dm.tasks[taskID] = task
	dm.mutex.Unlock()

	// 启动下载协程
	go dm.downloadCore(task)

	return taskID, nil
}

// GetTask 获取下载任务
func (dm *DownloadManager) GetTask(taskID string) (*config.DownloadTask, bool) {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	
	task, exists := dm.tasks[taskID]
	return task, exists
}

// GetAllTasks 获取所有下载任务
func (dm *DownloadManager) GetAllTasks() map[string]*config.DownloadTask {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()
	
	result := make(map[string]*config.DownloadTask)
	for k, v := range dm.tasks {
		result[k] = v
	}
	return result
}

// downloadCore 下载核心文件
func (dm *DownloadManager) downloadCore(task *config.DownloadTask) {
	// 更新状态为下载中
	dm.updateTaskStatus(task.ID, "downloading", "")

	// 创建目标目录
	coresDir := config.Get().Download.CoresDir
	if err := os.MkdirAll(coresDir, 0755); err != nil {
		dm.updateTaskStatus(task.ID, "failed", fmt.Sprintf("Failed to create directory: %v", err))
		return
	}

	// 目标文件路径
	filePath := filepath.Join(coresDir, task.CoreInfo.FileName)

	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		dm.updateTaskStatus(task.ID, "completed", "")
		task.Progress = 100
		now := time.Now()
		task.CompleteTime = &now
		return
	}

	// 开始下载
	if err := dm.downloadFile(task, filePath); err != nil {
		dm.updateTaskStatus(task.ID, "failed", err.Error())
		return
	}

	// 下载完成
	dm.updateTaskStatus(task.ID, "completed", "")
	task.Progress = 100
	now := time.Now()
	task.CompleteTime = &now
}

// downloadFile 下载文件
func (dm *DownloadManager) downloadFile(task *config.DownloadTask, filePath string) error {
	// 创建HTTP请求
	resp, err := http.Get(task.CoreInfo.DownloadURL)
	if err != nil {
		return fmt.Errorf("failed to start download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// 获取文件大小
	task.Total = resp.ContentLength

	// 创建目标文件
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// 下载文件并更新进度
	buffer := make([]byte, 32*1024) // 32KB buffer
	var downloaded int64
	startTime := time.Now()

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("failed to write file: %v", writeErr)
			}
			
			downloaded += int64(n)
			task.Downloaded = downloaded
			
			// 计算进度和速度
			if task.Total > 0 {
				task.Progress = float64(downloaded) / float64(task.Total) * 100
			}
			
			elapsed := time.Since(startTime).Seconds()
			if elapsed > 0 {
				task.Speed = int64(float64(downloaded) / elapsed)
			}
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("download error: %v", err)
		}
	}

	return nil
}

// updateTaskStatus 更新任务状态
func (dm *DownloadManager) updateTaskStatus(taskID, status, errorMsg string) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	
	if task, exists := dm.tasks[taskID]; exists {
		task.Status = status
		task.Error = errorMsg
	}
}
