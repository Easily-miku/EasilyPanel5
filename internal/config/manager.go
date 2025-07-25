package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Manager 配置管理器
type Manager struct {
	config     *Config
	configPath string
}

// NewManager 创建新的配置管理器
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Initialize 初始化配置管理器
func (m *Manager) Initialize() error {
	// 设置配置文件路径
	if m.configPath != "" {
		viper.SetConfigFile(m.configPath)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./configs")
		viper.AddConfigPath(".")
	}
	
	// 设置环境变量前缀
	viper.SetEnvPrefix("EASILYPANEL")
	viper.AutomaticEnv()
	
	// 设置默认值
	setDefaults()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，创建默认配置
			if err := m.CreateDefaultConfig(); err != nil {
				return fmt.Errorf("创建默认配置失败: %w", err)
			}
		} else {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
	}
	
	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	
	// 验证配置
	if err := ValidateConfig(&config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}
	
	// 确保必要的目录存在
	if err := m.ensureDirectories(&config); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	
	m.config = &config
	return nil
}

// GetConfig 获取配置
func (m *Manager) GetConfig() *Config {
	return m.config
}

// CreateDefaultConfig 创建默认配置文件
func (m *Manager) CreateDefaultConfig() error {
	configDir := "./configs"
	if m.configPath != "" {
		configDir = filepath.Dir(m.configPath)
	}
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	configFile := filepath.Join(configDir, "config.yaml")
	if m.configPath != "" {
		configFile = m.configPath
	}
	
	return viper.WriteConfigAs(configFile)
}

// ensureDirectories 确保必要的目录存在
func (m *Manager) ensureDirectories(config *Config) error {
	dirs := []string{
		config.App.DataDir,
		config.App.ConfigDir,
		filepath.Join(config.App.DataDir, "instances"),
		filepath.Join(config.App.DataDir, "backups"),
		filepath.Join(config.App.DataDir, "downloads"),
		filepath.Dir(config.Log.File),
		config.Backup.BackupDir,
	}
	
	for _, dir := range dirs {
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// UpdateConfig 更新配置
func (m *Manager) UpdateConfig(key string, value interface{}) error {
	viper.Set(key, value)
	
	// 重新解析配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}
	
	// 验证配置
	if err := ValidateConfig(&config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}
	
	// 保存配置
	if err := viper.WriteConfig(); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}
	
	m.config = &config
	return nil
}

// GetValue 获取配置值
func (m *Manager) GetValue(key string) interface{} {
	return viper.Get(key)
}

// SetValue 设置配置值
func (m *Manager) SetValue(key string, value interface{}) error {
	return m.UpdateConfig(key, value)
}

// ListConfigs 列出所有配置
func (m *Manager) ListConfigs() map[string]interface{} {
	return viper.AllSettings()
}

// SearchConfigs 搜索配置
func (m *Manager) SearchConfigs(keyword string) map[string]interface{} {
	allConfigs := viper.AllSettings()
	result := make(map[string]interface{})
	
	keyword = strings.ToLower(keyword)
	
	for key, value := range allConfigs {
		if strings.Contains(strings.ToLower(key), keyword) {
			result[key] = value
		}
	}
	
	return result
}

// GetConfigsByPrefix 根据前缀获取配置
func (m *Manager) GetConfigsByPrefix(prefix string) map[string]interface{} {
	allConfigs := viper.AllSettings()
	result := make(map[string]interface{})
	
	for key, value := range allConfigs {
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	
	return result
}

// ResetToDefaults 重置为默认配置
func (m *Manager) ResetToDefaults() error {
	if err := ResetConfig(); err != nil {
		return err
	}
	
	// 重新加载配置
	return m.Initialize()
}

// BackupCurrentConfig 备份当前配置
func (m *Manager) BackupCurrentConfig() (string, error) {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		return "", fmt.Errorf("未找到配置文件")
	}
	
	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")
	
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return "", fmt.Errorf("创建备份文件失败: %w", err)
	}
	
	return backupPath, nil
}

// RestoreFromBackup 从备份恢复配置
func (m *Manager) RestoreFromBackup(backupPath string) error {
	if err := RestoreConfig(backupPath); err != nil {
		return err
	}
	
	// 重新加载配置
	return m.Initialize()
}

// ValidateCurrentConfig 验证当前配置
func (m *Manager) ValidateCurrentConfig() error {
	return ValidateConfig(m.config)
}

// PrintConfig 打印配置信息
func (m *Manager) PrintConfig(section string) {
	var configs map[string]interface{}
	
	if section == "" {
		configs = viper.AllSettings()
		fmt.Println("所有配置:")
	} else {
		configs = m.GetConfigsByPrefix(section)
		fmt.Printf("%s 配置:\n", section)
	}
	
	if len(configs) == 0 {
		fmt.Println("未找到配置项")
		return
	}
	
	// 排序键名
	var keys []string
	for key := range configs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	
	fmt.Println("配置项                    | 值")
	fmt.Println("--------------------------|----")
	
	for _, key := range keys {
		value := configs[key]
		valueStr := fmt.Sprintf("%v", value)
		
		// 截断过长的值
		if len(valueStr) > 50 {
			valueStr = valueStr[:47] + "..."
		}
		
		fmt.Printf("%-25s | %s\n", key, valueStr)
	}
}

// GetConfigInfo 获取配置信息
func (m *Manager) GetConfigInfo() map[string]interface{} {
	return map[string]interface{}{
		"config_file":    viper.ConfigFileUsed(),
		"config_type":    viper.GetString("config_type"),
		"total_configs":  len(viper.AllKeys()),
		"data_dir":       m.config.App.DataDir,
		"config_dir":     m.config.App.ConfigDir,
		"language":       m.config.App.Language,
		"first_run":      m.config.App.FirstRun,
		"last_modified":  m.getConfigModTime(),
	}
}

// getConfigModTime 获取配置文件修改时间
func (m *Manager) getConfigModTime() string {
	configPath := viper.ConfigFileUsed()
	if configPath == "" {
		return "未知"
	}
	
	if info, err := os.Stat(configPath); err == nil {
		return info.ModTime().Format("2006-01-02 15:04:05")
	}
	
	return "未知"
}

// SetFirstRunComplete 设置首次运行完成
func (m *Manager) SetFirstRunComplete() error {
	return m.UpdateConfig("app.first_run", false)
}

// IsFirstRun 检查是否首次运行
func (m *Manager) IsFirstRun() bool {
	return m.config.App.FirstRun
}

// GetDataDir 获取数据目录
func (m *Manager) GetDataDir() string {
	return m.config.App.DataDir
}

// GetConfigDir 获取配置目录
func (m *Manager) GetConfigDir() string {
	return m.config.App.ConfigDir
}

// GetLanguage 获取语言设置
func (m *Manager) GetLanguage() string {
	return m.config.App.Language
}

// SetLanguage 设置语言
func (m *Manager) SetLanguage(language string) error {
	return m.UpdateConfig("app.language", language)
}
