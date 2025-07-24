package core

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// PluginAPI 插件API接口
type PluginAPI interface {
	GetPluginList(category string, page int) (*PluginListResponse, error)
	SearchPlugins(query string, category string) (*PluginListResponse, error)
	GetPluginInfo(pluginID string) (*PluginInfo, error)
	GetPluginVersions(pluginID string) (*PluginVersionsResponse, error)
	DownloadPlugin(pluginID string, version string) (string, error)
}

// PluginInfo 插件信息
type PluginInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Homepage    string   `json:"homepage"`
	Repository  string   `json:"repository"`
	Downloads   int64    `json:"downloads"`
	Rating      float64  `json:"rating"`
	LastUpdate  string   `json:"last_update"`
	Versions    []string `json:"versions"`
	LatestVersion string `json:"latest_version"`
	MinecraftVersions []string `json:"minecraft_versions"`
	Dependencies []string `json:"dependencies"`
	Icon        string   `json:"icon"`
	Screenshots []string `json:"screenshots"`
}

// PluginListResponse 插件列表响应
type PluginListResponse struct {
	Plugins    []PluginInfo `json:"plugins"`
	Total      int          `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// PluginVersionsResponse 插件版本响应
type PluginVersionsResponse struct {
	PluginID string          `json:"plugin_id"`
	Versions []PluginVersion `json:"versions"`
}

// PluginVersion 插件版本信息
type PluginVersion struct {
	Version         string   `json:"version"`
	ReleaseDate     string   `json:"release_date"`
	MinecraftVersions []string `json:"minecraft_versions"`
	DownloadURL     string   `json:"download_url"`
	Changelog       string   `json:"changelog"`
	FileSize        int64    `json:"file_size"`
	SHA1            string   `json:"sha1"`
	Dependencies    []string `json:"dependencies"`
}

// SpigotMCAPI SpigotMC API实现
type SpigotMCAPI struct {
	baseURL string
	client  *http.Client
}

// NewSpigotMCAPI 创建SpigotMC API实例
func NewSpigotMCAPI() *SpigotMCAPI {
	return &SpigotMCAPI{
		baseURL: "https://api.spigotmc.org/v2",
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetPluginList 获取插件列表
func (api *SpigotMCAPI) GetPluginList(category string, page int) (*PluginListResponse, error) {
	// SpigotMC API不直接支持分类和分页，这里模拟实现
	// 实际应用中可能需要使用其他数据源或缓存
	
	// 构建请求URL
	reqURL := fmt.Sprintf("%s/resources", api.baseURL)
	if category != "" && category != "all" {
		reqURL += fmt.Sprintf("?category=%s", url.QueryEscape(category))
	}
	
	resp, err := api.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin list: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	// 解析响应（这里需要根据实际API响应格式调整）
	var plugins []PluginInfo
	if err := json.Unmarshal(body, &plugins); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	// 分页处理
	pageSize := 20
	start := (page - 1) * pageSize
	end := start + pageSize
	
	if start >= len(plugins) {
		return &PluginListResponse{
			Plugins:    []PluginInfo{},
			Total:      len(plugins),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: (len(plugins) + pageSize - 1) / pageSize,
		}, nil
	}
	
	if end > len(plugins) {
		end = len(plugins)
	}
	
	return &PluginListResponse{
		Plugins:    plugins[start:end],
		Total:      len(plugins),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: (len(plugins) + pageSize - 1) / pageSize,
	}, nil
}

// SearchPlugins 搜索插件
func (api *SpigotMCAPI) SearchPlugins(query string, category string) (*PluginListResponse, error) {
	// 构建搜索URL
	reqURL := fmt.Sprintf("%s/search/resources/%s", api.baseURL, url.QueryEscape(query))
	
	resp, err := api.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to search plugins: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	var plugins []PluginInfo
	if err := json.Unmarshal(body, &plugins); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	// 按分类过滤
	if category != "" && category != "all" {
		var filtered []PluginInfo
		for _, plugin := range plugins {
			if strings.EqualFold(plugin.Category, category) {
				filtered = append(filtered, plugin)
			}
		}
		plugins = filtered
	}
	
	return &PluginListResponse{
		Plugins:    plugins,
		Total:      len(plugins),
		Page:       1,
		PageSize:   len(plugins),
		TotalPages: 1,
	}, nil
}

// GetPluginInfo 获取插件详细信息
func (api *SpigotMCAPI) GetPluginInfo(pluginID string) (*PluginInfo, error) {
	reqURL := fmt.Sprintf("%s/resources/%s", api.baseURL, pluginID)
	
	resp, err := api.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin info: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plugin not found or API error: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	var plugin PluginInfo
	if err := json.Unmarshal(body, &plugin); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	return &plugin, nil
}

// GetPluginVersions 获取插件版本列表
func (api *SpigotMCAPI) GetPluginVersions(pluginID string) (*PluginVersionsResponse, error) {
	reqURL := fmt.Sprintf("%s/resources/%s/versions", api.baseURL, pluginID)
	
	resp, err := api.client.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin versions: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("versions not found or API error: %d", resp.StatusCode)
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}
	
	var versions []PluginVersion
	if err := json.Unmarshal(body, &versions); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}
	
	return &PluginVersionsResponse{
		PluginID: pluginID,
		Versions: versions,
	}, nil
}

// DownloadPlugin 下载插件
func (api *SpigotMCAPI) DownloadPlugin(pluginID string, version string) (string, error) {
	// SpigotMC通常需要登录才能下载，这里返回下载URL
	downloadURL := fmt.Sprintf("https://www.spigotmc.org/resources/%s/download?version=%s", pluginID, version)
	return downloadURL, nil
}

// 全局插件API实例
var pluginAPI PluginAPI

// GetPluginAPI 获取插件API实例
func GetPluginAPI() PluginAPI {
	if pluginAPI == nil {
		pluginAPI = NewSpigotMCAPI()
	}
	return pluginAPI
}
