package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"easilypanel5/config"
	"easilypanel5/utils"
)

// SystemSettings 系统设置结构
type SystemSettings struct {
	// 常规设置
	ServerName        string `json:"server_name"`
	ServerDescription string `json:"server_description"`
	Language          string `json:"language"`
	Theme             string `json:"theme"`
	TimeZone          string `json:"timezone"`
	
	// Java设置
	JavaPath        string   `json:"java_path"`
	JavaAutoDetect  bool     `json:"java_auto_detect"`
	JavaDefaultArgs []string `json:"java_default_args"`
	
	// 下载设置
	DownloadMirror     string `json:"download_mirror"`
	DownloadThreads    int    `json:"download_threads"`
	DownloadTimeout    int    `json:"download_timeout"`
	DownloadRetries    int    `json:"download_retries"`
	
	// 安全设置
	EnableTwoFactor    bool   `json:"enable_two_factor"`
	SessionTimeout     int    `json:"session_timeout"`
	MaxLoginAttempts   int    `json:"max_login_attempts"`
	EnableIPWhitelist  bool   `json:"enable_ip_whitelist"`
	IPWhitelist        []string `json:"ip_whitelist"`
	
	// 通知设置
	EnableNotifications     bool   `json:"enable_notifications"`
	NotificationEmail       string `json:"notification_email"`
	NotifyServerStart       bool   `json:"notify_server_start"`
	NotifyServerStop        bool   `json:"notify_server_stop"`
	NotifyServerCrash       bool   `json:"notify_server_crash"`
	NotifyHighCPU           bool   `json:"notify_high_cpu"`
	NotifyHighMemory        bool   `json:"notify_high_memory"`
	NotifyDiskSpace         bool   `json:"notify_disk_space"`
	
	// 高级设置
	EnableDebugMode         bool   `json:"enable_debug_mode"`
	LogLevel                string `json:"log_level"`
	LogRetentionDays        int    `json:"log_retention_days"`
	BackupRetentionDays     int    `json:"backup_retention_days"`
	EnableAutoBackup        bool   `json:"enable_auto_backup"`
	AutoBackupInterval      int    `json:"auto_backup_interval"`
	EnableMetrics           bool   `json:"enable_metrics"`
	MetricsRetentionDays    int    `json:"metrics_retention_days"`
}

// handleSettings 处理设置请求
func handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetSettings(w, r)
	case http.MethodPut:
		handleUpdateSettings(w, r)
	default:
		WriteStandardError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetSettings 获取系统设置
func handleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := getSystemSettings()
	if err != nil {
		WriteStandardError(w, "GET_SETTINGS_FAILED", fmt.Sprintf("Failed to get settings: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, settings)
}

// handleUpdateSettings 更新系统设置
func handleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	var newSettings SystemSettings
	if err := json.NewDecoder(r.Body).Decode(&newSettings); err != nil {
		WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证设置
	if err := validateSettings(&newSettings); err != nil {
		WriteStandardError(w, "VALIDATION_FAILED", err.Error(), http.StatusBadRequest)
		return
	}

	// 保存设置
	if err := saveSystemSettings(&newSettings); err != nil {
		WriteStandardError(w, "SAVE_SETTINGS_FAILED", fmt.Sprintf("Failed to save settings: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, map[string]string{"status": "success"})
}

// getSystemSettings 获取系统设置
func getSystemSettings() (*SystemSettings, error) {
	cfg := config.Get()
	
	settings := &SystemSettings{
		// 常规设置（使用默认值，因为配置中没有这些字段）
		ServerName:        "EasilyPanel5",
		ServerDescription: "Minecraft服务器管理面板",
		Language:          "zh-CN",
		Theme:             "dark",
		TimeZone:          "Asia/Shanghai",

		// Java设置
		JavaPath:        cfg.Java.JavaPath,
		JavaAutoDetect:  cfg.Java.AutoDetect,
		JavaDefaultArgs: cfg.Java.DefaultArgs,

		// 下载设置
		DownloadMirror:     cfg.Download.FastMirrorAPI,
		DownloadThreads:    4, // 默认值
		DownloadTimeout:    int(cfg.Download.Timeout.Seconds()),
		DownloadRetries:    cfg.Download.MaxRetries,
		
		// 安全设置（默认值）
		EnableTwoFactor:    true,
		SessionTimeout:     3600,
		MaxLoginAttempts:   5,
		EnableIPWhitelist:  false,
		IPWhitelist:        []string{},
		
		// 通知设置（默认值）
		EnableNotifications:     true,
		NotificationEmail:       "",
		NotifyServerStart:       true,
		NotifyServerStop:        true,
		NotifyServerCrash:       true,
		NotifyHighCPU:           true,
		NotifyHighMemory:        true,
		NotifyDiskSpace:         true,
		
		// 高级设置（默认值）
		EnableDebugMode:         false,
		LogLevel:                "INFO",
		LogRetentionDays:        30,
		BackupRetentionDays:     7,
		EnableAutoBackup:        true,
		AutoBackupInterval:      24,
		EnableMetrics:           true,
		MetricsRetentionDays:    30,
	}

	return settings, nil
}

// saveSystemSettings 保存系统设置
func saveSystemSettings(settings *SystemSettings) error {
	cfg := config.Get()

	// 更新配置（只更新实际存在的字段）
	cfg.Java.JavaPath = settings.JavaPath
	cfg.Java.AutoDetect = settings.JavaAutoDetect
	cfg.Java.DefaultArgs = settings.JavaDefaultArgs
	cfg.Download.FastMirrorAPI = settings.DownloadMirror
	cfg.Download.Timeout = time.Duration(settings.DownloadTimeout) * time.Second
	cfg.Download.MaxRetries = settings.DownloadRetries

	// 保存配置
	return config.Update(cfg)
}

// validateSettings 验证设置
func validateSettings(settings *SystemSettings) error {
	var validationErrors utils.ValidationErrors
	
	// 验证服务器名称
	if err := utils.ValidateRequired("server_name", settings.ServerName); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	} else if err := utils.ValidateStringLength("server_name", settings.ServerName, 1, 100); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	// 验证Java路径
	if settings.JavaPath != "" {
		if err := utils.ValidateFilePath("java_path", settings.JavaPath); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}
	
	// 验证Java参数
	if len(settings.JavaDefaultArgs) > 0 {
		if err := utils.ValidateJavaArgs("java_default_args", settings.JavaDefaultArgs); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}
	
	// 验证下载设置
	if err := utils.ValidateRange("download_threads", settings.DownloadThreads, 1, 16); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("download_timeout", settings.DownloadTimeout, 10, 300); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("download_retries", settings.DownloadRetries, 0, 10); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	// 验证安全设置
	if err := utils.ValidateRange("session_timeout", settings.SessionTimeout, 300, 86400); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("max_login_attempts", settings.MaxLoginAttempts, 1, 20); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	// 验证IP白名单
	for i, ip := range settings.IPWhitelist {
		if err := utils.ValidateIP(fmt.Sprintf("ip_whitelist[%d]", i), ip); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}
	
	// 验证通知邮箱
	if settings.NotificationEmail != "" {
		if err := utils.ValidateEmail("notification_email", settings.NotificationEmail); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}
	}
	
	// 验证高级设置
	logLevels := []string{"DEBUG", "INFO", "WARN", "ERROR"}
	if err := utils.ValidateStringInSlice("log_level", settings.LogLevel, logLevels); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("log_retention_days", settings.LogRetentionDays, 1, 365); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("backup_retention_days", settings.BackupRetentionDays, 1, 90); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("auto_backup_interval", settings.AutoBackupInterval, 1, 168); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if err := utils.ValidateRange("metrics_retention_days", settings.MetricsRetentionDays, 1, 365); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			validationErrors = append(validationErrors, valErr)
		}
	}
	
	if len(validationErrors) > 0 {
		return validationErrors
	}
	
	return nil
}
