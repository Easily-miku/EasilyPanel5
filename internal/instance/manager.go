package instance

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Manager 实例管理器
type Manager struct {
	dataDir string
}

// NewManager 创建新的实例管理器
func NewManager(dataDir string) *Manager {
	return &Manager{
		dataDir: dataDir,
	}
}

// CreateMinecraftInstance 创建Minecraft实例
func (m *Manager) CreateMinecraftInstance(name, mcVersion, serverType, javaPath string) (*Instance, error) {
	// 检查实例是否已存在
	if m.InstanceExists(name) {
		return nil, fmt.Errorf("实例 '%s' 已存在", name)
	}
	
	// 验证名称
	if err := m.validateInstanceName(name); err != nil {
		return nil, err
	}
	
	// 创建实例
	instance := NewMinecraftInstance(name, mcVersion, serverType, javaPath)
	instance.SetWorkDir(m.dataDir)
	
	// 保存配置
	if err := instance.Save(m.dataDir); err != nil {
		return nil, fmt.Errorf("保存实例配置失败: %w", err)
	}
	
	return instance, nil
}

// CreateBlankInstance 创建空白实例
func (m *Manager) CreateBlankInstance(name, description, startCmd string) (*Instance, error) {
	// 检查实例是否已存在
	if m.InstanceExists(name) {
		return nil, fmt.Errorf("实例 '%s' 已存在", name)
	}
	
	// 验证名称
	if err := m.validateInstanceName(name); err != nil {
		return nil, err
	}
	
	// 创建实例
	instance := NewBlankInstance(name, description, startCmd)
	instance.SetWorkDir(m.dataDir)
	
	// 保存配置
	if err := instance.Save(m.dataDir); err != nil {
		return nil, fmt.Errorf("保存实例配置失败: %w", err)
	}
	
	return instance, nil
}

// GetInstance 获取实例
func (m *Manager) GetInstance(name string) (*Instance, error) {
	return Load(name, m.dataDir)
}

// ListInstances 列出所有实例
func (m *Manager) ListInstances() ([]*Instance, error) {
	instancesDir := filepath.Join(m.dataDir, "instances")
	
	// 确保目录存在
	if err := os.MkdirAll(instancesDir, 0755); err != nil {
		return nil, fmt.Errorf("创建实例目录失败: %w", err)
	}
	
	entries, err := os.ReadDir(instancesDir)
	if err != nil {
		return nil, fmt.Errorf("读取实例目录失败: %w", err)
	}
	
	var instances []*Instance
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		// 只处理.json文件
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		
		// 提取实例名称
		name := strings.TrimSuffix(entry.Name(), ".json")
		
		// 加载实例
		instance, err := Load(name, m.dataDir)
		if err != nil {
			// 记录错误但继续处理其他实例
			fmt.Printf("警告: 加载实例 '%s' 失败: %v\n", name, err)
			continue
		}
		
		instances = append(instances, instance)
	}
	
	return instances, nil
}

// DeleteInstance 删除实例
func (m *Manager) DeleteInstance(name string, deleteFiles bool) error {
	// 检查实例是否存在
	instance, err := m.GetInstance(name)
	if err != nil {
		return err
	}
	
	// 如果实例正在运行，先停止
	if instance.IsRunning() {
		return fmt.Errorf("实例 '%s' 正在运行，请先停止实例", name)
	}
	
	// 删除配置文件
	if err := instance.Delete(m.dataDir); err != nil {
		return fmt.Errorf("删除实例配置失败: %w", err)
	}
	
	// 如果需要，删除工作目录
	if deleteFiles {
		workDir := instance.GetWorkDir(m.dataDir)
		if err := os.RemoveAll(workDir); err != nil {
			return fmt.Errorf("删除实例文件失败: %w", err)
		}
	}
	
	return nil
}

// InstanceExists 检查实例是否存在
func (m *Manager) InstanceExists(name string) bool {
	configFile := filepath.Join(m.dataDir, "instances", fmt.Sprintf("%s.json", name))
	_, err := os.Stat(configFile)
	return err == nil
}

// UpdateInstance 更新实例配置
func (m *Manager) UpdateInstance(instance *Instance) error {
	return instance.Save(m.dataDir)
}

// GetInstancesByType 根据类型获取实例列表
func (m *Manager) GetInstancesByType(instanceType InstanceType) ([]*Instance, error) {
	allInstances, err := m.ListInstances()
	if err != nil {
		return nil, err
	}
	
	var filteredInstances []*Instance
	for _, instance := range allInstances {
		if instance.Type == instanceType {
			filteredInstances = append(filteredInstances, instance)
		}
	}
	
	return filteredInstances, nil
}

// GetRunningInstances 获取正在运行的实例
func (m *Manager) GetRunningInstances() ([]*Instance, error) {
	allInstances, err := m.ListInstances()
	if err != nil {
		return nil, err
	}
	
	var runningInstances []*Instance
	for _, instance := range allInstances {
		if instance.IsRunning() {
			runningInstances = append(runningInstances, instance)
		}
	}
	
	return runningInstances, nil
}

// validateInstanceName 验证实例名称
func (m *Manager) validateInstanceName(name string) error {
	if name == "" {
		return fmt.Errorf("实例名称不能为空")
	}
	
	// 检查名称长度
	if len(name) > 50 {
		return fmt.Errorf("实例名称不能超过50个字符")
	}
	
	// 检查非法字符
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", " "}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("实例名称不能包含字符: %s", char)
		}
	}
	
	// 检查保留名称
	reservedNames := []string{"con", "prn", "aux", "nul", "com1", "com2", "com3", "com4", "com5", "com6", "com7", "com8", "com9", "lpt1", "lpt2", "lpt3", "lpt4", "lpt5", "lpt6", "lpt7", "lpt8", "lpt9"}
	lowerName := strings.ToLower(name)
	for _, reserved := range reservedNames {
		if lowerName == reserved {
			return fmt.Errorf("实例名称不能使用保留名称: %s", name)
		}
	}
	
	return nil
}

// PrintInstanceList 打印实例列表
func (m *Manager) PrintInstanceList() error {
	instances, err := m.ListInstances()
	if err != nil {
		return err
	}
	
	if len(instances) == 0 {
		fmt.Println("未找到任何实例")
		return nil
	}
	
	fmt.Printf("找到 %d 个实例:\n\n", len(instances))
	fmt.Println("名称          | 类型      | 状态    | 版本      | 描述")
	fmt.Println("--------------|-----------|---------|-----------|----")
	
	for _, instance := range instances {
		status := string(instance.Status)
		version := instance.MCVersion
		if version == "" {
			version = "-"
		}
		
		// 截断过长的描述
		description := instance.Description
		if len(description) > 30 {
			description = description[:27] + "..."
		}
		
		fmt.Printf("%-12s | %-9s | %-7s | %-9s | %s\n",
			instance.Name,
			instance.Type,
			status,
			version,
			description)
	}
	
	return nil
}

// GetInstanceInfo 获取实例详细信息
func (m *Manager) GetInstanceInfo(name string) (map[string]interface{}, error) {
	instance, err := m.GetInstance(name)
	if err != nil {
		return nil, err
	}
	
	info := map[string]interface{}{
		"name":         instance.Name,
		"type":         instance.Type,
		"status":       instance.Status,
		"description":  instance.Description,
		"created_at":   instance.CreatedAt.Format("2006-01-02 15:04:05"),
		"updated_at":   instance.UpdatedAt.Format("2006-01-02 15:04:05"),
		"work_dir":     instance.GetWorkDir(m.dataDir),
		"auto_start":   instance.AutoStart,
		"auto_restart": instance.AutoRestart,
	}
	
	if instance.Type == TypeMinecraft {
		info["mc_version"] = instance.MCVersion
		info["server_type"] = instance.ServerType
		info["server_jar"] = instance.ServerJar
		info["java_path"] = instance.JavaPath
		info["port"] = instance.Port
		info["max_memory"] = instance.MaxMemory
		info["min_memory"] = instance.MinMemory
	} else {
		info["start_cmd"] = instance.StartCmd
	}
	
	if instance.PID > 0 {
		info["pid"] = instance.PID
	}
	
	if instance.LastStarted != nil {
		info["last_started"] = instance.LastStarted.Format("2006-01-02 15:04:05")
	}
	
	if instance.LastStopped != nil {
		info["last_stopped"] = instance.LastStopped.Format("2006-01-02 15:04:05")
	}
	
	return info, nil
}
