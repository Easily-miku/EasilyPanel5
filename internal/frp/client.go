package frp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	DefaultAPIBaseURL = "https://api.openfrp.net"
	DefaultUserAgent  = "EasilyPanel5/1.0.0"
)

// OpenFRPClient OpenFRP API客户端
type OpenFRPClient struct {
	httpClient    *http.Client
	baseURL       string
	authorization string
	userAgent     string
}

// NewOpenFRPClient 创建新的OpenFRP客户端
func NewOpenFRPClient(baseURL, authorization string) *OpenFRPClient {
	if baseURL == "" {
		baseURL = DefaultAPIBaseURL
	}
	
	return &OpenFRPClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:       baseURL,
		authorization: authorization,
		userAgent:     DefaultUserAgent,
	}
}

// SetTimeout 设置请求超时时间
func (c *OpenFRPClient) SetTimeout(timeout time.Duration) {
	c.httpClient.Timeout = timeout
}

// SetAuthorization 设置认证令牌
func (c *OpenFRPClient) SetAuthorization(auth string) {
	c.authorization = auth
}

// GetAuthorization 获取认证令牌
func (c *OpenFRPClient) GetAuthorization() string {
	return c.authorization
}

// makeRequest 发送HTTP请求
func (c *OpenFRPClient) makeRequest(method, endpoint string, body interface{}) (*APIResponse, error) {
	url := c.baseURL + endpoint
	
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("序列化请求数据失败: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	if c.authorization != "" {
		req.Header.Set("Authorization", c.authorization)
	}

	// 调试信息（生产环境可注释掉）
	// fmt.Printf("DEBUG: 发送请求 %s %s\n", method, url)
	// if c.authorization != "" && len(c.authorization) > 8 {
	//     fmt.Printf("DEBUG: Authorization: %s...%s\n", c.authorization[:8], c.authorization[len(c.authorization)-4:])
	// }

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	// fmt.Printf("DEBUG: 响应状态: %d %s\n", resp.StatusCode, resp.Status)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// fmt.Printf("DEBUG: 响应内容: %s\n", string(respBody))
	
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return &apiResp, fmt.Errorf("HTTP错误: %d %s", resp.StatusCode, resp.Status)
	}
	
	// 检查API响应状态
	if !apiResp.Flag {
		return &apiResp, fmt.Errorf("API错误: %s", apiResp.Msg)
	}
	
	return &apiResp, nil
}

// GetUserInfo 获取用户信息
func (c *OpenFRPClient) GetUserInfo() (*UserInfo, error) {
	resp, err := c.makeRequest("POST", "/frp/api/getUserInfo", nil)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	var userInfo UserInfo
	if err := mapToStruct(resp.Data, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %w", err)
	}

	return &userInfo, nil
}

// GetProxies 获取用户隧道列表
func (c *OpenFRPClient) GetProxies() (*ProxyListResponse, error) {
	// 首先尝试标准API
	resp, err := c.makeRequest("POST", "/frp/api/getUserProxies", nil)
	if err != nil {
		// 如果失败，尝试使用简单的token API
		return c.getProxiesWithToken()
	}

	var proxyList ProxyListResponse
	if err := mapToStruct(resp.Data, &proxyList); err != nil {
		return nil, fmt.Errorf("解析隧道列表失败: %w", err)
	}

	return &proxyList, nil
}

// getProxiesWithToken 使用token方式获取隧道列表（备用方法）
func (c *OpenFRPClient) getProxiesWithToken() (*ProxyListResponse, error) {
	if c.authorization == "" {
		return nil, fmt.Errorf("未设置认证令牌")
	}

	// 构建URL
	url := fmt.Sprintf("%s/api?action=getallproxies&user=%s", c.baseURL, c.authorization)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d %s", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 这个API返回的是直接的隧道数组，不是标准的APIResponse格式
	var proxies []ProxyInfo
	if err := json.Unmarshal(respBody, &proxies); err != nil {
		return nil, fmt.Errorf("解析隧道列表失败: %w", err)
	}

	return &ProxyListResponse{
		List: proxies,
	}, nil
}

// GetNodes 获取节点列表
func (c *OpenFRPClient) GetNodes() (*NodeListResponse, error) {
	// 首先尝试标准API
	resp, err := c.makeRequest("POST", "/frp/api/getNodeList", nil)
	if err != nil {
		// 如果失败，尝试使用GET方式获取节点列表
		return c.getNodesWithGET()
	}

	var nodeList NodeListResponse
	if err := mapToStruct(resp.Data, &nodeList); err != nil {
		return nil, fmt.Errorf("解析节点列表失败: %w", err)
	}

	return &nodeList, nil
}

// getNodesWithGET 使用GET方式获取节点列表（备用方法）
func (c *OpenFRPClient) getNodesWithGET() (*NodeListResponse, error) {
	// 构建URL
	url := fmt.Sprintf("%s/api?action=getnodelist", c.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)
	if c.authorization != "" {
		req.Header.Set("Authorization", c.authorization)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP错误: %d %s", resp.StatusCode, resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	// 尝试解析为标准API响应格式
	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err == nil && apiResp.Flag {
		var nodeList NodeListResponse
		if err := mapToStruct(apiResp.Data, &nodeList); err != nil {
			return nil, fmt.Errorf("解析节点列表失败: %w", err)
		}
		return &nodeList, nil
	}

	// 如果不是标准格式，尝试直接解析为节点数组
	var nodes []NodeInfo
	if err := json.Unmarshal(respBody, &nodes); err != nil {
		return nil, fmt.Errorf("解析节点列表失败: %w", err)
	}

	return &NodeListResponse{
		List: nodes,
	}, nil
}

// CreateProxy 创建新隧道
func (c *OpenFRPClient) CreateProxy(req *CreateProxyRequest) error {
	_, err := c.makeRequest("POST", "/frp/api/newProxy", req)
	if err != nil {
		return fmt.Errorf("创建隧道失败: %w", err)
	}
	
	return nil
}

// EditProxy 编辑隧道
func (c *OpenFRPClient) EditProxy(req *EditProxyRequest) error {
	_, err := c.makeRequest("POST", "/frp/api/editProxy", req)
	if err != nil {
		return fmt.Errorf("编辑隧道失败: %w", err)
	}
	
	return nil
}

// DeleteProxy 删除隧道
func (c *OpenFRPClient) DeleteProxy(proxyID int) error {
	req := map[string]interface{}{
		"proxy_id": proxyID,
	}
	
	_, err := c.makeRequest("POST", "/frp/api/removeProxy", req)
	if err != nil {
		return fmt.Errorf("删除隧道失败: %w", err)
	}
	
	return nil
}

// GetProxiesByToken 通过Token获取隧道列表（简化API）
func (c *OpenFRPClient) GetProxiesByToken(token string) (*TokenProxyResponse, error) {
	endpoint := fmt.Sprintf("/api?action=getallproxies&user=%s", token)
	
	resp, err := c.makeRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("获取隧道列表失败: %w", err)
	}
	
	var tokenResp TokenProxyResponse
	if err := mapToStruct(resp.Data, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析隧道列表失败: %w", err)
	}
	
	return &tokenResp, nil
}

// TestConnection 测试连接
func (c *OpenFRPClient) TestConnection() error {
	// 首先尝试获取用户信息
	_, err := c.GetUserInfo()
	if err != nil {
		// 如果失败，尝试使用简单的token API验证
		return c.testConnectionWithToken()
	}
	return nil
}

// testConnectionWithToken 使用token方式测试连接
func (c *OpenFRPClient) testConnectionWithToken() error {
	if c.authorization == "" {
		return fmt.Errorf("未设置认证令牌")
	}

	// 尝试获取隧道列表来验证token
	_, err := c.getProxiesWithToken()
	return err
}

// mapToStruct 将map转换为结构体
func mapToStruct(data interface{}, target interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(jsonData, target)
}
