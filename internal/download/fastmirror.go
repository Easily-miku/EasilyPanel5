package download

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	FastMirrorBaseURL = "https://download.fastmirror.net"
	FastMirrorAPIV3   = FastMirrorBaseURL + "/api/v3"
)

// FastMirrorResponse FastMirror API统一响应格式
type FastMirrorResponse struct {
	Data    interface{} `json:"data"`
	Code    string      `json:"code"`
	Success bool        `json:"success"`
	Message string      `json:"message"`
}

// ServerInfo 服务端信息
type ServerInfo struct {
	Name      string `json:"name"`
	Tag       string `json:"tag"`
	Recommend bool   `json:"recommend"`
}

// ProjectInfo 项目详细信息
type ProjectInfo struct {
	Name       string   `json:"name"`
	Tag        string   `json:"tag"`
	Homepage   string   `json:"homepage"`
	MCVersions []string `json:"mc_versions"`
}

// BuildInfo 构建信息
type BuildInfo struct {
	Name        string `json:"name"`
	MCVersion   string `json:"mc_version"`
	CoreVersion string `json:"core_version"`
	UpdateTime  string `json:"update_time"`
	SHA1        string `json:"sha1"`
}

// BuildsResponse 构建列表响应
type BuildsResponse struct {
	Builds []BuildInfo `json:"builds"`
	Offset int         `json:"offset"`
	Limit  int         `json:"limit"`
	Count  int         `json:"count"`
}

// CoreInfo 核心详细信息
type CoreInfo struct {
	Name        string `json:"name"`
	MCVersion   string `json:"mc_version"`
	CoreVersion string `json:"core_version"`
	UpdateTime  string `json:"update_time"`
	SHA1        string `json:"sha1"`
	Filename    string `json:"filename"`
	DownloadURL string `json:"download_url"`
}

// FastMirrorClient FastMirror API客户端
type FastMirrorClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewFastMirrorClient 创建新的FastMirror客户端
func NewFastMirrorClient() *FastMirrorClient {
	return &FastMirrorClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: FastMirrorAPIV3,
	}
}

// SetTimeout 设置请求超时时间
func (c *FastMirrorClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// makeRequest 发送HTTP请求
func (c *FastMirrorClient) makeRequest(endpoint string) (*FastMirrorResponse, error) {
	url := c.baseURL + endpoint
	
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d %s", resp.StatusCode, resp.Status)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}
	
	var response FastMirrorResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}
	
	if !response.Success {
		return nil, fmt.Errorf("API错误: %s (%s)", response.Message, response.Code)
	}
	
	return &response, nil
}

// GetServerList 获取支持的服务端列表
func (c *FastMirrorClient) GetServerList() ([]ServerInfo, error) {
	response, err := c.makeRequest("")
	if err != nil {
		return nil, fmt.Errorf("获取服务端列表失败: %w", err)
	}
	
	// 解析数据
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %w", err)
	}
	
	var servers []ServerInfo
	if err := json.Unmarshal(dataBytes, &servers); err != nil {
		return nil, fmt.Errorf("解析服务端列表失败: %w", err)
	}
	
	return servers, nil
}

// GetProjectInfo 获取项目详细信息
func (c *FastMirrorClient) GetProjectInfo(name string) (*ProjectInfo, error) {
	endpoint := "/" + url.PathEscape(name)
	
	response, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	
	// 解析数据
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %w", err)
	}
	
	var project ProjectInfo
	if err := json.Unmarshal(dataBytes, &project); err != nil {
		return nil, fmt.Errorf("解析项目信息失败: %w", err)
	}
	
	return &project, nil
}

// GetBuilds 获取构建版本列表
func (c *FastMirrorClient) GetBuilds(name, mcVersion string, offset, count int) (*BuildsResponse, error) {
	endpoint := fmt.Sprintf("/%s/%s", url.PathEscape(name), url.PathEscape(mcVersion))
	
	// 添加查询参数
	if offset > 0 || count > 0 {
		params := url.Values{}
		if offset > 0 {
			params.Add("offset", strconv.Itoa(offset))
		}
		if count > 0 {
			params.Add("count", strconv.Itoa(count))
		}
		endpoint += "?" + params.Encode()
	}
	
	response, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取构建列表失败: %w", err)
	}
	
	// 解析数据
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %w", err)
	}
	
	var builds BuildsResponse
	if err := json.Unmarshal(dataBytes, &builds); err != nil {
		return nil, fmt.Errorf("解析构建列表失败: %w", err)
	}
	
	return &builds, nil
}

// GetCoreInfo 获取指定核心详细信息
func (c *FastMirrorClient) GetCoreInfo(name, mcVersion, coreVersion string) (*CoreInfo, error) {
	endpoint := fmt.Sprintf("/%s/%s/%s", 
		url.PathEscape(name), 
		url.PathEscape(mcVersion), 
		url.PathEscape(coreVersion))
	
	response, err := c.makeRequest(endpoint)
	if err != nil {
		return nil, fmt.Errorf("获取核心信息失败: %w", err)
	}
	
	// 解析数据
	dataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, fmt.Errorf("序列化数据失败: %w", err)
	}
	
	var core CoreInfo
	if err := json.Unmarshal(dataBytes, &core); err != nil {
		return nil, fmt.Errorf("解析核心信息失败: %w", err)
	}
	
	return &core, nil
}

// GetLatestBuild 获取最新构建版本
func (c *FastMirrorClient) GetLatestBuild(name, mcVersion string) (*BuildInfo, error) {
	builds, err := c.GetBuilds(name, mcVersion, 0, 1)
	if err != nil {
		return nil, err
	}
	
	if len(builds.Builds) == 0 {
		return nil, fmt.Errorf("未找到 %s %s 的构建版本", name, mcVersion)
	}
	
	return &builds.Builds[0], nil
}

// GetRecommendedServers 获取推荐的服务端
func (c *FastMirrorClient) GetRecommendedServers() ([]ServerInfo, error) {
	servers, err := c.GetServerList()
	if err != nil {
		return nil, err
	}
	
	var recommended []ServerInfo
	for _, server := range servers {
		if server.Recommend {
			recommended = append(recommended, server)
		}
	}
	
	return recommended, nil
}

// GetServersByTag 根据标签获取服务端
func (c *FastMirrorClient) GetServersByTag(tag string) ([]ServerInfo, error) {
	servers, err := c.GetServerList()
	if err != nil {
		return nil, err
	}
	
	var filtered []ServerInfo
	for _, server := range servers {
		if server.Tag == tag {
			filtered = append(filtered, server)
		}
	}
	
	return filtered, nil
}

// SearchServers 搜索服务端
func (c *FastMirrorClient) SearchServers(keyword string) ([]ServerInfo, error) {
	servers, err := c.GetServerList()
	if err != nil {
		return nil, err
	}

	var results []ServerInfo
	keyword = strings.ToLower(keyword) // 转为小写进行搜索

	for _, server := range servers {
		serverName := strings.ToLower(server.Name)
		if strings.Contains(serverName, keyword) {
			results = append(results, server)
		}
	}

	return results, nil
}
