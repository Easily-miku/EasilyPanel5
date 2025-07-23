package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	globalConfig *AppConfig
	configMutex  sync.RWMutex
	configFile   = "data/config.json"
)

// Initialize 初始化配置
func Initialize() error {
	configMutex.Lock()
	defer configMutex.Unlock()

	// 创建默认配置
	globalConfig = getDefaultConfig()

	// 尝试加载现有配置
	if err := loadConfig(); err != nil {
		// 如果加载失败，保存默认配置
		if saveErr := saveConfig(); saveErr != nil {
			return fmt.Errorf("failed to save default config: %v", saveErr)
		}
	}

	return nil
}

// Get 获取全局配置
func Get() *AppConfig {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return globalConfig
}

// Update 更新配置
func Update(newConfig *AppConfig) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	globalConfig = newConfig
	return saveConfig()
}

// getDefaultConfig 获取默认配置
func getDefaultConfig() *AppConfig {
	return &AppConfig{
		Server: ServerConfig{
			Port:         8080,
			Host:         "0.0.0.0",
			StaticDir:    "web",
			MaxServers:   10,
			DefaultMemory: 2048,
		},
		Java: JavaConfig{
			AutoDetect:  true,
			JavaPath:    "",
			DefaultArgs: []string{"-Xms1G", "-Xmx2G", "-XX:+UseG1GC"},
			MinVersion:  8,
			MaxMemory:   8192,
		},
		Download: DownloadConfig{
			CoresDir:      "data/cores",
			Timeout:       30 * time.Minute,
			MaxRetries:    3,
			ChunkSize:     1024 * 1024, // 1MB
			FastMirrorAPI: "https://download.fastmirror.net/api/v3",
		},
		Logging: LoggingConfig{
			Level:        "info",
			LogsDir:      "data/logs",
			MaxFileSize:  10 * 1024 * 1024, // 10MB
			MaxFiles:     5,
			MaxAge:       7,
			EnableColors: true,
		},
	}
}

// loadConfig 加载配置文件
func loadConfig() error {
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config file does not exist")
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var config AppConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	globalConfig = &config
	return nil
}

// saveConfig 保存配置文件
func saveConfig() error {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	data, err := json.MarshalIndent(globalConfig, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// Servers 服务器配置管理
type Servers struct {
	mutex   sync.RWMutex
	servers map[string]*MinecraftServer
	file    string
}

var globalServers *Servers

// InitServers 初始化服务器配置
func InitServers() error {
	globalServers = &Servers{
		servers: make(map[string]*MinecraftServer),
		file:    "data/servers.json",
	}
	return globalServers.load()
}

// GetServers 获取服务器管理器
func GetServers() *Servers {
	return globalServers
}

// GetAll 获取所有服务器
func (s *Servers) GetAll() map[string]*MinecraftServer {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	result := make(map[string]*MinecraftServer)
	for k, v := range s.servers {
		result[k] = v
	}
	return result
}

// Get 获取指定服务器
func (s *Servers) Get(id string) (*MinecraftServer, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	server, exists := s.servers[id]
	return server, exists
}

// Add 添加服务器
func (s *Servers) Add(server *MinecraftServer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	server.UpdatedAt = time.Now()
	s.servers[server.ID] = server
	return s.save()
}

// Update 更新服务器
func (s *Servers) Update(server *MinecraftServer) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.servers[server.ID]; !exists {
		return fmt.Errorf("server not found: %s", server.ID)
	}
	
	server.UpdatedAt = time.Now()
	s.servers[server.ID] = server
	return s.save()
}

// Delete 删除服务器
func (s *Servers) Delete(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.servers[id]; !exists {
		return fmt.Errorf("server not found: %s", id)
	}
	
	delete(s.servers, id)
	return s.save()
}

// load 加载服务器配置
func (s *Servers) load() error {
	if _, err := os.Stat(s.file); os.IsNotExist(err) {
		return s.save() // 创建空文件
	}

	data, err := os.ReadFile(s.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.servers)
}

// save 保存服务器配置
func (s *Servers) save() error {
	data, err := json.MarshalIndent(s.servers, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.file, data, 0644)
}
