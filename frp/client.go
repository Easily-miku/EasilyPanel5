package frp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenFRPClient OpenFRP API客户端
type OpenFRPClient struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
}

// NewOpenFRPClient 创建新的OpenFRP客户端
func NewOpenFRPClient(baseURL string) *OpenFRPClient {
	return &OpenFRPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		userAgent: "EasilyPanel5/1.1.0",
	}
}

// APIResponse OpenFRP API响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// doRequest 执行HTTP请求
func (c *OpenFRPClient) doRequest(method, endpoint string, body interface{}, token string) (*APIResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	url := c.baseURL + endpoint
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// 执行请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查API错误
	if apiResp.Code != 0 {
		return nil, &OpenFRPError{
			Code:    apiResp.Code,
			Message: apiResp.Message,
		}
	}

	return &apiResp, nil
}

// GetNodes 获取节点列表
func (c *OpenFRPClient) GetNodes() ([]*OpenFRPNode, error) {
	resp, err := c.doRequest("GET", "/api/nodes", nil, "")
	if err != nil {
		return nil, err
	}

	var nodes []*OpenFRPNode
	if err := c.parseData(resp.Data, &nodes); err != nil {
		return nil, fmt.Errorf("failed to parse nodes data: %v", err)
	}

	return nodes, nil
}

// GetUserInfo 获取用户信息
func (c *OpenFRPClient) GetUserInfo(token string) (*OpenFRPUser, error) {
	resp, err := c.doRequest("GET", "/api/user", nil, token)
	if err != nil {
		return nil, err
	}

	var user OpenFRPUser
	if err := c.parseData(resp.Data, &user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %v", err)
	}

	return &user, nil
}

// ValidateToken 验证令牌
func (c *OpenFRPClient) ValidateToken(token string) error {
	_, err := c.GetUserInfo(token)
	return err
}

// CreateTunnel 创建隧道
func (c *OpenFRPClient) CreateTunnel(token string, req *TunnelRequest) (*TunnelConfig, error) {
	resp, err := c.doRequest("POST", "/api/tunnels", req, token)
	if err != nil {
		return nil, err
	}

	var tunnel TunnelConfig
	if err := c.parseData(resp.Data, &tunnel); err != nil {
		return nil, fmt.Errorf("failed to parse tunnel data: %v", err)
	}

	return &tunnel, nil
}

// DeleteTunnel 删除隧道
func (c *OpenFRPClient) DeleteTunnel(token, tunnelID string) error {
	_, err := c.doRequest("DELETE", "/api/tunnels/"+tunnelID, nil, token)
	return err
}

// StartTunnel 启动隧道
func (c *OpenFRPClient) StartTunnel(token, tunnelID string) error {
	_, err := c.doRequest("POST", "/api/tunnels/"+tunnelID+"/start", nil, token)
	return err
}

// StopTunnel 停止隧道
func (c *OpenFRPClient) StopTunnel(token, tunnelID string) error {
	_, err := c.doRequest("POST", "/api/tunnels/"+tunnelID+"/stop", nil, token)
	return err
}

// GetTunnel 获取隧道信息
func (c *OpenFRPClient) GetTunnel(token, tunnelID string) (*TunnelConfig, error) {
	resp, err := c.doRequest("GET", "/api/tunnels/"+tunnelID, nil, token)
	if err != nil {
		return nil, err
	}

	var tunnel TunnelConfig
	if err := c.parseData(resp.Data, &tunnel); err != nil {
		return nil, fmt.Errorf("failed to parse tunnel data: %v", err)
	}

	return &tunnel, nil
}

// GetTunnels 获取隧道列表
func (c *OpenFRPClient) GetTunnels(token string) ([]*TunnelConfig, error) {
	resp, err := c.doRequest("GET", "/api/tunnels", nil, token)
	if err != nil {
		return nil, err
	}

	var tunnels []*TunnelConfig
	if err := c.parseData(resp.Data, &tunnels); err != nil {
		return nil, fmt.Errorf("failed to parse tunnels data: %v", err)
	}

	return tunnels, nil
}

// GetTunnelStats 获取隧道统计
func (c *OpenFRPClient) GetTunnelStats(token, tunnelID string) (*TunnelStats, error) {
	resp, err := c.doRequest("GET", "/api/tunnels/"+tunnelID+"/stats", nil, token)
	if err != nil {
		return nil, err
	}

	var stats TunnelStats
	if err := c.parseData(resp.Data, &stats); err != nil {
		return nil, fmt.Errorf("failed to parse stats data: %v", err)
	}

	return &stats, nil
}

// parseData 解析API响应数据
func (c *OpenFRPClient) parseData(data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, target)
}

// SetTimeout 设置请求超时时间
func (c *OpenFRPClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetUserAgent 设置User-Agent
func (c *OpenFRPClient) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// Health 检查API健康状态
func (c *OpenFRPClient) Health() error {
	resp, err := c.doRequest("GET", "/api/health", nil, "")
	if err != nil {
		return err
	}
	
	if resp.Code != 0 {
		return fmt.Errorf("API health check failed: %s", resp.Message)
	}
	
	return nil
}

// GetVersion 获取API版本信息
func (c *OpenFRPClient) GetVersion() (map[string]interface{}, error) {
	resp, err := c.doRequest("GET", "/api/version", nil, "")
	if err != nil {
		return nil, err
	}

	var version map[string]interface{}
	if err := c.parseData(resp.Data, &version); err != nil {
		return nil, fmt.Errorf("failed to parse version data: %v", err)
	}

	return version, nil
}
