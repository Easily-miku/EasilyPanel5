package download

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ProgressCallback 下载进度回调函数
type ProgressCallback func(downloaded, total int64, percent float64)

// Downloader 文件下载器
type Downloader struct {
	httpClient *http.Client
	userAgent  string
}

// NewDownloader 创建新的下载器
func NewDownloader() *Downloader {
	return &Downloader{
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // 5分钟超时
		},
		userAgent: "EasilyPanel5/1.0.0",
	}
}

// SetTimeout 设置下载超时时间
func (d *Downloader) SetTimeout(timeout time.Duration) {
	d.httpClient.Timeout = timeout
}

// SetUserAgent 设置User-Agent
func (d *Downloader) SetUserAgent(userAgent string) {
	d.userAgent = userAgent
}

// DownloadFile 下载文件
func (d *Downloader) DownloadFile(url, destPath string, callback ProgressCallback) error {
	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	// 创建HTTP请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	
	// 设置User-Agent
	req.Header.Set("User-Agent", d.userAgent)
	
	// 发送请求
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败: HTTP %d %s", resp.StatusCode, resp.Status)
	}
	
	// 获取文件大小
	contentLength := resp.Header.Get("Content-Length")
	var totalSize int64
	if contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			totalSize = size
		}
	}
	
	// 创建临时文件
	tempPath := destPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer file.Close()
	
	// 创建进度读取器
	var reader io.Reader = resp.Body
	if callback != nil && totalSize > 0 {
		reader = &progressReader{
			reader:   resp.Body,
			total:    totalSize,
			callback: callback,
		}
	}
	
	// 复制数据
	_, err = io.Copy(file, reader)
	if err != nil {
		os.Remove(tempPath) // 清理临时文件
		return fmt.Errorf("下载文件失败: %w", err)
	}
	
	// 关闭文件
	file.Close()
	
	// 重命名临时文件为目标文件
	if err := os.Rename(tempPath, destPath); err != nil {
		os.Remove(tempPath) // 清理临时文件
		return fmt.Errorf("重命名文件失败: %w", err)
	}
	
	return nil
}

// DownloadWithRetry 带重试的下载
func (d *Downloader) DownloadWithRetry(url, destPath string, retries int, callback ProgressCallback) error {
	var lastErr error
	
	for i := 0; i <= retries; i++ {
		if i > 0 {
			fmt.Printf("下载失败，正在重试 (%d/%d)...\n", i, retries)
			time.Sleep(time.Duration(i) * time.Second) // 递增延迟
		}
		
		err := d.DownloadFile(url, destPath, callback)
		if err == nil {
			return nil
		}
		
		lastErr = err
	}
	
	return fmt.Errorf("下载失败，已重试 %d 次: %w", retries, lastErr)
}

// VerifyFile 验证文件SHA1校验和
func (d *Downloader) VerifyFile(filePath, expectedSHA1 string) error {
	if expectedSHA1 == "" {
		return nil // 没有提供校验和，跳过验证
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()
	
	hash := sha1.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("计算文件哈希失败: %w", err)
	}
	
	actualSHA1 := fmt.Sprintf("%x", hash.Sum(nil))
	if !strings.EqualFold(actualSHA1, expectedSHA1) {
		return fmt.Errorf("文件校验失败: 期望 %s, 实际 %s", expectedSHA1, actualSHA1)
	}
	
	return nil
}

// GetFileSize 获取远程文件大小
func (d *Downloader) GetFileSize(url string) (int64, error) {
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return 0, fmt.Errorf("创建请求失败: %w", err)
	}
	
	req.Header.Set("User-Agent", d.userAgent)
	
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("请求失败: HTTP %d %s", resp.StatusCode, resp.Status)
	}
	
	contentLength := resp.Header.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("无法获取文件大小")
	}
	
	size, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("解析文件大小失败: %w", err)
	}
	
	return size, nil
}

// progressReader 带进度的读取器
type progressReader struct {
	reader     io.Reader
	total      int64
	downloaded int64
	callback   ProgressCallback
	lastUpdate time.Time
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	if n > 0 {
		pr.downloaded += int64(n)
		
		// 限制回调频率，避免过于频繁的更新
		now := time.Now()
		if now.Sub(pr.lastUpdate) >= 100*time.Millisecond || pr.downloaded == pr.total {
			percent := float64(pr.downloaded) / float64(pr.total) * 100
			pr.callback(pr.downloaded, pr.total, percent)
			pr.lastUpdate = now
		}
	}
	return n, err
}

// FormatBytes 格式化字节数为人类可读格式
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatSpeed 格式化下载速度
func FormatSpeed(bytesPerSecond float64) string {
	return FormatBytes(int64(bytesPerSecond)) + "/s"
}

// DownloadProgress 下载进度信息
type DownloadProgress struct {
	Downloaded int64
	Total      int64
	Percent    float64
	Speed      float64 // 字节/秒
	ETA        time.Duration
	StartTime  time.Time
}

// NewDownloadProgress 创建下载进度跟踪器
func NewDownloadProgress() *DownloadProgress {
	return &DownloadProgress{
		StartTime: time.Now(),
	}
}

// Update 更新进度信息
func (dp *DownloadProgress) Update(downloaded, total int64) {
	dp.Downloaded = downloaded
	dp.Total = total
	
	if total > 0 {
		dp.Percent = float64(downloaded) / float64(total) * 100
	}
	
	elapsed := time.Since(dp.StartTime)
	if elapsed.Seconds() > 0 {
		dp.Speed = float64(downloaded) / elapsed.Seconds()
		
		if dp.Speed > 0 && total > downloaded {
			remaining := total - downloaded
			dp.ETA = time.Duration(float64(remaining)/dp.Speed) * time.Second
		}
	}
}

// String 返回进度的字符串表示
func (dp *DownloadProgress) String() string {
	if dp.Total > 0 {
		return fmt.Sprintf("%.1f%% (%s/%s) %s ETA: %v",
			dp.Percent,
			FormatBytes(dp.Downloaded),
			FormatBytes(dp.Total),
			FormatSpeed(dp.Speed),
			dp.ETA.Round(time.Second))
	}
	return fmt.Sprintf("%s %s",
		FormatBytes(dp.Downloaded),
		FormatSpeed(dp.Speed))
}
