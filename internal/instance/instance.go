package instance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// InstanceType 实例类型
type InstanceType string

const (
	TypeMinecraft InstanceType = "minecraft"
	TypeBlank     InstanceType = "blank"
)

// InstanceStatus 实例状态
type InstanceStatus string

const (
	StatusStopped InstanceStatus = "stopped"
	StatusRunning InstanceStatus = "running"
	StatusStarting InstanceStatus = "starting"
	StatusStopping InstanceStatus = "stopping"
	StatusError   InstanceStatus = "error"
)

// Instance 服务器实例结构
type Instance struct {
	// 基本信息
	Name        string         `json:"name"`
	Type        InstanceType   `json:"type"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	
	// 路径信息
	WorkDir     string `json:"work_dir"`
	ServerJar   string `json:"server_jar,omitempty"`
	
	// 启动配置
	JavaPath    string   `json:"java_path"`
	JavaArgs    []string `json:"java_args"`
	ServerArgs  []string `json:"server_args"`
	StartCmd    string   `json:"start_cmd,omitempty"` // 自定义启动命令（覆盖默认启动方式）
	UseCustomCmd bool    `json:"use_custom_cmd"`     // 是否使用自定义启动命令
	
	// Minecraft特定配置
	MCVersion   string `json:"mc_version,omitempty"`
	ServerType  string `json:"server_type,omitempty"` // paper, fabric, forge等
	
	// 运行时信息
	Status      InstanceStatus `json:"status"`
	PID         int           `json:"pid,omitempty"`
	Port        int           `json:"port,omitempty"`
	LastStarted *time.Time    `json:"last_started,omitempty"`
	LastStopped *time.Time    `json:"last_stopped,omitempty"`
	
	// 资源限制
	MaxMemory   string `json:"max_memory,omitempty"`   // 如 "2G", "1024M"
	MinMemory   string `json:"min_memory,omitempty"`   // 如 "1G", "512M"
	
	// 自动化设置
	AutoStart   bool `json:"auto_start"`
	AutoRestart bool `json:"auto_restart"`
}

// NewMinecraftInstance 创建新的Minecraft实例
func NewMinecraftInstance(name, mcVersion, serverType, javaPath string) *Instance {
	now := time.Now()
	return &Instance{
		Name:        name,
		Type:        TypeMinecraft,
		Description: fmt.Sprintf("Minecraft %s 服务器 (%s)", mcVersion, serverType),
		CreatedAt:   now,
		UpdatedAt:   now,
		MCVersion:   mcVersion,
		ServerType:  serverType,
		JavaPath:    javaPath,
		Status:      StatusStopped,
		AutoStart:   false,
		AutoRestart: false,
		Port:        25565, // 默认Minecraft端口
		MaxMemory:   "2G",
		MinMemory:   "1G",
		JavaArgs: []string{
			"-Xmx2G",
			"-Xms1G",
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
		},
		ServerArgs: []string{"nogui"},
	}
}

// NewBlankInstance 创建新的空白实例
func NewBlankInstance(name, description, startCmd string) *Instance {
	now := time.Now()
	return &Instance{
		Name:        name,
		Type:        TypeBlank,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		StartCmd:    startCmd,
		Status:      StatusStopped,
		AutoStart:   false,
		AutoRestart: false,
	}
}

// GetConfigFile 获取实例配置文件路径
func (i *Instance) GetConfigFile(dataDir string) string {
	return filepath.Join(dataDir, "instances", fmt.Sprintf("%s.json", i.Name))
}

// GetWorkDir 获取实例工作目录
func (i *Instance) GetWorkDir(dataDir string) string {
	if i.WorkDir != "" {
		return i.WorkDir
	}
	return filepath.Join(dataDir, "instances", i.Name)
}

// SetWorkDir 设置工作目录
func (i *Instance) SetWorkDir(dataDir string) {
	i.WorkDir = i.GetWorkDir(dataDir)
	i.UpdatedAt = time.Now()
}

// UpdateMemorySettings 更新内存设置
func (i *Instance) UpdateMemorySettings(minMem, maxMem string) {
	i.MinMemory = minMem
	i.MaxMemory = maxMem
	
	// 更新Java参数中的内存设置
	var newJavaArgs []string
	for _, arg := range i.JavaArgs {
		if !strings.HasPrefix(arg, "-Xmx") && !strings.HasPrefix(arg, "-Xms") {
			newJavaArgs = append(newJavaArgs, arg)
		}
	}
	
	// 添加新的内存参数
	newJavaArgs = append([]string{fmt.Sprintf("-Xmx%s", maxMem), fmt.Sprintf("-Xms%s", minMem)}, newJavaArgs...)
	i.JavaArgs = newJavaArgs
	i.UpdatedAt = time.Now()
}

// GetStartCommand 获取启动命令
func (i *Instance) GetStartCommand() (string, []string, error) {
	// 如果启用了自定义启动命令，优先使用自定义命令
	if i.UseCustomCmd && i.StartCmd != "" {
		// 解析自定义启动命令
		parts := strings.Fields(i.StartCmd)
		if len(parts) == 0 {
			return "", nil, fmt.Errorf("自定义启动命令为空")
		}
		return parts[0], parts[1:], nil
	}

	// 空白实例必须使用自定义启动命令
	if i.Type == TypeBlank {
		if i.StartCmd == "" {
			return "", nil, fmt.Errorf("空白实例未设置启动命令")
		}
		// 解析自定义启动命令
		parts := strings.Fields(i.StartCmd)
		if len(parts) == 0 {
			return "", nil, fmt.Errorf("启动命令为空")
		}
		return parts[0], parts[1:], nil
	}

	// Minecraft实例使用默认Java启动方式
	if i.JavaPath == "" {
		return "", nil, fmt.Errorf("未设置Java路径")
	}

	if i.ServerJar == "" {
		return "", nil, fmt.Errorf("未设置服务器JAR文件")
	}

	var args []string
	args = append(args, i.JavaArgs...)
	args = append(args, "-jar", i.ServerJar)
	args = append(args, i.ServerArgs...)

	return i.JavaPath, args, nil
}

// UpdateStatus 更新实例状态
func (i *Instance) UpdateStatus(status InstanceStatus) {
	i.Status = status
	i.UpdatedAt = time.Now()
	
	now := time.Now()
	switch status {
	case StatusRunning:
		i.LastStarted = &now
	case StatusStopped:
		i.LastStopped = &now
		i.PID = 0
	}
}

// SetPID 设置进程ID
func (i *Instance) SetPID(pid int) {
	i.PID = pid
	i.UpdatedAt = time.Now()
}

// IsRunning 检查实例是否正在运行
func (i *Instance) IsRunning() bool {
	return i.Status == StatusRunning || i.Status == StatusStarting
}

// Validate 验证实例配置
func (i *Instance) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("实例名称不能为空")
	}
	
	if i.Type == TypeMinecraft {
		if i.JavaPath == "" {
			return fmt.Errorf("Minecraft实例必须设置Java路径")
		}
		if i.MCVersion == "" {
			return fmt.Errorf("Minecraft实例必须设置MC版本")
		}
	} else if i.Type == TypeBlank {
		if i.StartCmd == "" {
			return fmt.Errorf("空白实例必须设置启动命令")
		}
	} else {
		return fmt.Errorf("未知的实例类型: %s", i.Type)
	}
	
	return nil
}

// Save 保存实例配置到文件
func (i *Instance) Save(dataDir string) error {
	if err := i.Validate(); err != nil {
		return fmt.Errorf("实例配置验证失败: %w", err)
	}
	
	configFile := i.GetConfigFile(dataDir)
	
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(configFile), 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	
	// 确保工作目录存在
	workDir := i.GetWorkDir(dataDir)
	if err := os.MkdirAll(workDir, 0755); err != nil {
		return fmt.Errorf("创建工作目录失败: %w", err)
	}
	
	// 序列化配置
	data, err := json.MarshalIndent(i, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化实例配置失败: %w", err)
	}
	
	// 写入文件
	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	return nil
}

// Load 从文件加载实例配置
func Load(name, dataDir string) (*Instance, error) {
	configFile := filepath.Join(dataDir, "instances", fmt.Sprintf("%s.json", name))
	
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("实例 '%s' 不存在", name)
		}
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	var instance Instance
	if err := json.Unmarshal(data, &instance); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	return &instance, nil
}

// Delete 删除实例
func (i *Instance) Delete(dataDir string) error {
	// 删除配置文件
	configFile := i.GetConfigFile(dataDir)
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除配置文件失败: %w", err)
	}
	
	// 询问是否删除工作目录
	// 这里暂时不自动删除，避免误删用户数据
	
	return nil
}
