package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 应用配置结构
type Config struct {
	// 应用设置
	App AppConfig `mapstructure:"app"`

	// 日志设置
	Log LogConfig `mapstructure:"log"`

	// Java设置
	Java JavaConfig `mapstructure:"java"`

	// 下载设置
	Download DownloadConfig `mapstructure:"download"`

	// 实例设置
	Instance InstanceConfig `mapstructure:"instance"`

	// FRP设置
	FRP FRPConfig `mapstructure:"frp"`

	// FTP设置
	FTP FTPConfig `mapstructure:"ftp"`

	// 备份设置
	Backup BackupConfig `mapstructure:"backup"`

	// 守护进程设置
	Daemon DaemonConfig `mapstructure:"daemon"`

	// 网络设置
	Network NetworkConfig `mapstructure:"network"`
}

// AppConfig 应用配置
type AppConfig struct {
	DataDir       string `mapstructure:"data_dir"`
	ConfigDir     string `mapstructure:"config_dir"`
	Language      string `mapstructure:"language"`
	AutoBackup    bool   `mapstructure:"auto_backup"`
	BackupCount   int    `mapstructure:"backup_count"`
	CheckUpdates  bool   `mapstructure:"check_updates"`
	FirstRun      bool   `mapstructure:"first_run"`
	Theme         string `mapstructure:"theme"`
	MaxInstances  int    `mapstructure:"max_instances"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

// JavaConfig Java配置
type JavaConfig struct {
	AutoDetect   bool     `mapstructure:"auto_detect"`
	DefaultPath  string   `mapstructure:"default_path"`
	SearchPaths  []string `mapstructure:"search_paths"`
	ExcludePaths []string `mapstructure:"exclude_paths"`
}

// DownloadConfig 下载配置
type DownloadConfig struct {
	DefaultSource   string            `mapstructure:"default_source"`
	Timeout         int               `mapstructure:"timeout"`
	Retry           int               `mapstructure:"retry"`
	MaxConcurrent   int               `mapstructure:"max_concurrent"`
	VerifyChecksum  bool              `mapstructure:"verify_checksum"`
	AutoCleanup     bool              `mapstructure:"auto_cleanup"`
	CleanupDays     int               `mapstructure:"cleanup_days"`
	Sources         map[string]string `mapstructure:"sources"`
}

// InstanceConfig 实例配置
type InstanceConfig struct {
	DefaultJavaArgs   []string          `mapstructure:"default_java_args"`
	DefaultServerArgs []string          `mapstructure:"default_server_args"`
	DefaultMemory     MemoryConfig      `mapstructure:"default_memory"`
	AutoEULA          bool              `mapstructure:"auto_eula"`
	AutoRestart       bool              `mapstructure:"auto_restart"`
	RestartDelay      int               `mapstructure:"restart_delay"`
	MaxRestarts       int               `mapstructure:"max_restarts"`
	LogRetention      int               `mapstructure:"log_retention"`
	Templates         map[string]string `mapstructure:"templates"`
}

// MemoryConfig 内存配置
type MemoryConfig struct {
	Min string `mapstructure:"min"`
	Max string `mapstructure:"max"`
}

// BackupConfig 备份配置
type BackupConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	AutoBackup      bool     `mapstructure:"auto_backup"`
	BackupInterval  string   `mapstructure:"backup_interval"`
	MaxBackups      int      `mapstructure:"max_backups"`
	CompressBackups bool     `mapstructure:"compress_backups"`
	BackupDir       string   `mapstructure:"backup_dir"`
	ExcludePatterns []string `mapstructure:"exclude_patterns"`
	IncludeWorlds   bool     `mapstructure:"include_worlds"`
	IncludePlugins  bool     `mapstructure:"include_plugins"`
	IncludeLogs     bool     `mapstructure:"include_logs"`
}

// DaemonConfig 守护进程配置
type DaemonConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	ServiceName   string `mapstructure:"service_name"`
	AutoStart     bool   `mapstructure:"auto_start"`
	RestartPolicy string `mapstructure:"restart_policy"`
	User          string `mapstructure:"user"`
	Group         string `mapstructure:"group"`
}

// NetworkConfig 网络配置
type NetworkConfig struct {
	ProxyEnabled bool   `mapstructure:"proxy_enabled"`
	ProxyType    string `mapstructure:"proxy_type"`
	ProxyHost    string `mapstructure:"proxy_host"`
	ProxyPort    int    `mapstructure:"proxy_port"`
	ProxyUser    string `mapstructure:"proxy_user"`
	ProxyPass    string `mapstructure:"proxy_pass"`
	Timeout      int    `mapstructure:"timeout"`
	UserAgent    string `mapstructure:"user_agent"`
}

// FRPConfig 内网穿透配置
type FRPConfig struct {
	// 基础设置
	Enabled    bool   `mapstructure:"enabled"`
	ServerAddr string `mapstructure:"server_addr"`
	Token      string `mapstructure:"token"`
	AutoConfig bool   `mapstructure:"auto_config"`

	// OpenFRP设置
	OpenFRP OpenFRPConfig `mapstructure:"openfrp"`

	// frpc客户端设置
	Client FRPClientConfig `mapstructure:"client"`

	// 隧道默认设置
	Defaults FRPDefaultsConfig `mapstructure:"defaults"`
}

// OpenFRPConfig OpenFRP特定配置
type OpenFRPConfig struct {
	APIBaseURL    string `mapstructure:"api_base_url"`
	Authorization string `mapstructure:"authorization"`
	AutoLogin     bool   `mapstructure:"auto_login"`
	AutoUpdate    bool   `mapstructure:"auto_update"`
	PreferredNode int    `mapstructure:"preferred_node"`
}

// FRPClientConfig frpc客户端配置
type FRPClientConfig struct {
	AutoDownload   bool   `mapstructure:"auto_download"`
	BinaryPath     string `mapstructure:"binary_path"`
	ConfigPath     string `mapstructure:"config_path"`
	LogPath        string `mapstructure:"log_path"`
	LogLevel       string `mapstructure:"log_level"`
	AutoStart      bool   `mapstructure:"auto_start"`
	RestartOnFail  bool   `mapstructure:"restart_on_fail"`
	MaxRestarts    int    `mapstructure:"max_restarts"`
	HealthCheck    bool   `mapstructure:"health_check"`
	CheckInterval  int    `mapstructure:"check_interval"`
}

// FRPDefaultsConfig 隧道默认配置
type FRPDefaultsConfig struct {
	LocalIP         string `mapstructure:"local_ip"`
	UseEncryption   bool   `mapstructure:"use_encryption"`
	UseCompression  bool   `mapstructure:"use_compression"`
	ProxyProtocol   bool   `mapstructure:"proxy_protocol"`
	AutoTLS         bool   `mapstructure:"auto_tls"`
	ForceHTTPS      bool   `mapstructure:"force_https"`
	CustomDomain    string `mapstructure:"custom_domain"`
	HealthCheckType string `mapstructure:"health_check_type"`
	HealthCheckURL  string `mapstructure:"health_check_url"`
}

// FTPConfig FTP配置
type FTPConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	RootDir  string `mapstructure:"root_dir"`
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	// 设置配置文件名和路径
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	
	// 设置环境变量前缀
	viper.SetEnvPrefix("EASILYPANEL")
	viper.AutomaticEnv()
	
	// 设置默认值
	setDefaults()
	
	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// 配置文件不存在，创建默认配置
			if err := createDefaultConfig(); err != nil {
				return nil, fmt.Errorf("创建默认配置失败: %w", err)
			}
		} else {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}
	}
	
	// 解析配置到结构体
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("解析配置失败: %w", err)
	}
	
	// 确保必要的目录存在
	if err := ensureDirectories(&config); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}
	
	return &config, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	// 应用默认设置
	viper.SetDefault("app.data_dir", "./data")
	viper.SetDefault("app.config_dir", "./configs")
	viper.SetDefault("app.language", "zh_CN")
	viper.SetDefault("app.auto_backup", true)
	viper.SetDefault("app.backup_count", 5)
	viper.SetDefault("app.check_updates", true)
	viper.SetDefault("app.first_run", true)
	viper.SetDefault("app.theme", "default")
	viper.SetDefault("app.max_instances", 10)

	// 日志默认设置
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file", "./logs/easilypanel.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
	viper.SetDefault("log.compress", true)

	// Java默认设置
	viper.SetDefault("java.auto_detect", true)
	viper.SetDefault("java.search_paths", []string{})
	viper.SetDefault("java.exclude_paths", []string{})

	// 下载默认设置
	viper.SetDefault("download.default_source", "fastmirror")
	viper.SetDefault("download.timeout", 300)
	viper.SetDefault("download.retry", 3)
	viper.SetDefault("download.max_concurrent", 3)
	viper.SetDefault("download.verify_checksum", true)
	viper.SetDefault("download.auto_cleanup", false)
	viper.SetDefault("download.cleanup_days", 30)
	viper.SetDefault("download.sources", map[string]string{
		"fastmirror": "https://download.fastmirror.net/api/v3",
		"mcsl":       "https://sync.mcsl.com.cn/api",
	})

	// 实例默认设置
	viper.SetDefault("instance.default_java_args", []string{
		"-Xmx2G", "-Xms1G",
		"-XX:+UseG1GC",
		"-XX:+ParallelRefProcEnabled",
		"-XX:MaxGCPauseMillis=200",
		"-XX:+UnlockExperimentalVMOptions",
		"-XX:+DisableExplicitGC",
		"-XX:+AlwaysPreTouch",
		"-XX:G1NewSizePercent=30",
		"-XX:G1MaxNewSizePercent=40",
		"-XX:G1HeapRegionSize=8M",
		"-XX:G1ReservePercent=20",
		"-XX:G1HeapWastePercent=5",
		"-XX:G1MixedGCCountTarget=4",
		"-XX:InitiatingHeapOccupancyPercent=15",
		"-XX:G1MixedGCLiveThresholdPercent=90",
		"-XX:G1RSetUpdatingPauseTimePercent=5",
		"-XX:SurvivorRatio=32",
		"-XX:+PerfDisableSharedMem",
		"-XX:MaxTenuringThreshold=1",
	})
	viper.SetDefault("instance.default_server_args", []string{"nogui"})
	viper.SetDefault("instance.default_memory.min", "1G")
	viper.SetDefault("instance.default_memory.max", "2G")
	viper.SetDefault("instance.auto_eula", false)
	viper.SetDefault("instance.auto_restart", false)
	viper.SetDefault("instance.restart_delay", 5)
	viper.SetDefault("instance.max_restarts", 3)
	viper.SetDefault("instance.log_retention", 7)
	viper.SetDefault("instance.templates", map[string]string{})

	// 备份默认设置
	viper.SetDefault("backup.enabled", true)
	viper.SetDefault("backup.auto_backup", false)
	viper.SetDefault("backup.backup_interval", "24h")
	viper.SetDefault("backup.max_backups", 10)
	viper.SetDefault("backup.compress_backups", true)
	viper.SetDefault("backup.backup_dir", "./data/backups")
	viper.SetDefault("backup.exclude_patterns", []string{"*.log", "*.tmp", "cache/*"})
	viper.SetDefault("backup.include_worlds", true)
	viper.SetDefault("backup.include_plugins", true)
	viper.SetDefault("backup.include_logs", false)

	// 守护进程默认设置
	viper.SetDefault("daemon.enabled", false)
	viper.SetDefault("daemon.service_name", "easilypanel")
	viper.SetDefault("daemon.auto_start", false)
	viper.SetDefault("daemon.restart_policy", "always")
	viper.SetDefault("daemon.user", "")
	viper.SetDefault("daemon.group", "")

	// 网络默认设置
	viper.SetDefault("network.proxy_enabled", false)
	viper.SetDefault("network.proxy_type", "http")
	viper.SetDefault("network.proxy_host", "")
	viper.SetDefault("network.proxy_port", 8080)
	viper.SetDefault("network.proxy_user", "")
	viper.SetDefault("network.proxy_pass", "")
	viper.SetDefault("network.timeout", 30)
	viper.SetDefault("network.user_agent", "EasilyPanel5/1.0.0")

	// FRP默认设置
	viper.SetDefault("frp.enabled", false)
	viper.SetDefault("frp.server_addr", "")
	viper.SetDefault("frp.token", "")
	viper.SetDefault("frp.auto_config", true)

	// OpenFRP默认设置
	viper.SetDefault("frp.openfrp.api_base_url", "https://api.openfrp.net")
	viper.SetDefault("frp.openfrp.authorization", "")
	viper.SetDefault("frp.openfrp.auto_login", false)
	viper.SetDefault("frp.openfrp.auto_update", true)
	viper.SetDefault("frp.openfrp.preferred_node", 0)

	// frpc客户端默认设置
	viper.SetDefault("frp.client.auto_download", true)
	viper.SetDefault("frp.client.binary_path", "./bin/frpc")
	viper.SetDefault("frp.client.config_path", "./configs/frpc.ini")
	viper.SetDefault("frp.client.log_path", "./logs/frpc.log")
	viper.SetDefault("frp.client.log_level", "info")
	viper.SetDefault("frp.client.auto_start", false)
	viper.SetDefault("frp.client.restart_on_fail", true)
	viper.SetDefault("frp.client.max_restarts", 3)
	viper.SetDefault("frp.client.health_check", true)
	viper.SetDefault("frp.client.check_interval", 30)

	// 隧道默认设置
	viper.SetDefault("frp.defaults.local_ip", "127.0.0.1")
	viper.SetDefault("frp.defaults.use_encryption", false)
	viper.SetDefault("frp.defaults.use_compression", false)
	viper.SetDefault("frp.defaults.proxy_protocol", false)
	viper.SetDefault("frp.defaults.auto_tls", false)
	viper.SetDefault("frp.defaults.force_https", false)
	viper.SetDefault("frp.defaults.custom_domain", "")
	viper.SetDefault("frp.defaults.health_check_type", "tcp")
	viper.SetDefault("frp.defaults.health_check_url", "")

	// FTP默认设置
	viper.SetDefault("ftp.enabled", false)
	viper.SetDefault("ftp.port", 21)
	viper.SetDefault("ftp.user", "admin")
	viper.SetDefault("ftp.root_dir", "./data")
}

// createDefaultConfig 创建默认配置文件
func createDefaultConfig() error {
	configDir := "./configs"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	configFile := filepath.Join(configDir, "config.yaml")
	return viper.WriteConfigAs(configFile)
}

// ensureDirectories 确保必要的目录存在
func ensureDirectories(config *Config) error {
	dirs := []string{
		config.App.DataDir,
		config.App.ConfigDir,
		filepath.Join(config.App.DataDir, "instances"),
		filepath.Join(config.App.DataDir, "backups"),
		filepath.Join(config.App.DataDir, "downloads"),
		filepath.Dir(config.Log.File),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	
	return nil
}

// SaveConfig 保存配置到文件
func (c *Config) Save() error {
	return viper.WriteConfig()
}

// SaveConfig 保存当前配置到文件
func SaveConfig() error {
	return viper.WriteConfig()
}

// GetString 获取字符串配置
func GetString(key string) string {
	return viper.GetString(key)
}

// GetInt 获取整数配置
func GetInt(key string) int {
	return viper.GetInt(key)
}

// GetBool 获取布尔配置
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetStringSlice 获取字符串数组配置
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}

// GetStringMap 获取字符串映射配置
func GetStringMap(key string) map[string]interface{} {
	return viper.GetStringMap(key)
}

// GetDuration 获取时间间隔配置
func GetDuration(key string) time.Duration {
	return viper.GetDuration(key)
}

// Set 设置配置值
func Set(key string, value interface{}) {
	viper.Set(key, value)
}

// SetString 设置字符串配置
func SetString(key, value string) {
	viper.Set(key, value)
}

// SetInt 设置整数配置
func SetInt(key string, value int) {
	viper.Set(key, value)
}

// SetBool 设置布尔配置
func SetBool(key string, value bool) {
	viper.Set(key, value)
}

// IsSet 检查配置是否已设置
func IsSet(key string) bool {
	return viper.IsSet(key)
}

// AllKeys 获取所有配置键
func AllKeys() []string {
	return viper.AllKeys()
}

// AllSettings 获取所有配置
func AllSettings() map[string]interface{} {
	return viper.AllSettings()
}

// WriteConfig 写入配置文件
func WriteConfig() error {
	return viper.WriteConfig()
}

// WriteConfigAs 写入配置到指定文件
func WriteConfigAs(filename string) error {
	return viper.WriteConfigAs(filename)
}

// ReloadConfig 重新加载配置
func ReloadConfig() error {
	return viper.ReadInConfig()
}

// ValidateConfig 验证配置
func ValidateConfig(config *Config) error {
	// 验证数据目录
	if config.App.DataDir == "" {
		return fmt.Errorf("数据目录不能为空")
	}

	// 验证日志级别
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, strings.ToLower(config.Log.Level)) {
		return fmt.Errorf("无效的日志级别: %s", config.Log.Level)
	}

	// 验证内存配置
	if config.Instance.DefaultMemory.Min == "" || config.Instance.DefaultMemory.Max == "" {
		return fmt.Errorf("默认内存配置不能为空")
	}

	// 验证备份间隔
	if config.Backup.BackupInterval != "" {
		if _, err := time.ParseDuration(config.Backup.BackupInterval); err != nil {
			return fmt.Errorf("无效的备份间隔: %s", config.Backup.BackupInterval)
		}
	}

	// 验证网络超时
	if config.Network.Timeout <= 0 {
		return fmt.Errorf("网络超时必须大于0")
	}

	return nil
}

// contains 检查字符串数组是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetConfigPath 获取配置文件路径
func GetConfigPath() string {
	return viper.ConfigFileUsed()
}

// BackupConfigFile 备份配置文件
func BackupConfigFile() error {
	configPath := GetConfigPath()
	if configPath == "" {
		return fmt.Errorf("未找到配置文件")
	}

	backupPath := configPath + ".backup." + time.Now().Format("20060102150405")

	// 读取原文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 写入备份文件
	if err := os.WriteFile(backupPath, data, 0644); err != nil {
		return fmt.Errorf("创建备份文件失败: %w", err)
	}

	return nil
}

// RestoreConfig 恢复配置文件
func RestoreConfig(backupPath string) error {
	configPath := GetConfigPath()
	if configPath == "" {
		return fmt.Errorf("未找到配置文件")
	}

	// 读取备份文件
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %w", err)
	}

	// 写入配置文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("恢复配置文件失败: %w", err)
	}

	// 重新加载配置
	return ReloadConfig()
}

// ResetConfig 重置配置为默认值
func ResetConfig() error {
	// 清除所有配置
	for _, key := range AllKeys() {
		viper.Set(key, nil)
	}

	// 重新设置默认值
	setDefaults()

	// 保存配置
	return WriteConfig()
}

// ExportConfig 导出配置到指定格式
func ExportConfig(format, filename string) error {
	switch strings.ToLower(format) {
	case "yaml", "yml":
		viper.SetConfigType("yaml")
	case "json":
		viper.SetConfigType("json")
	case "toml":
		viper.SetConfigType("toml")
	default:
		return fmt.Errorf("不支持的配置格式: %s", format)
	}

	return viper.WriteConfigAs(filename)
}

// ImportConfig 从文件导入配置
func ImportConfig(filename string) error {
	// 备份当前配置
	if err := BackupConfigFile(); err != nil {
		return fmt.Errorf("备份当前配置失败: %w", err)
	}

	// 设置配置文件
	viper.SetConfigFile(filename)

	// 读取配置
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 验证配置
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	if err := ValidateConfig(&config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 保存配置
	return WriteConfig()
}
