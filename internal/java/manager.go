package java

import (
	"fmt"
	"path/filepath"
	"time"
)

// Manager Java管理器
type Manager struct {
	detector     *Detector
	configDir    string
	javaListFile string
	javaList     []*Java
}

// NewManager 创建新的Java管理器
func NewManager(configDir string) *Manager {
	javaListFile := filepath.Join(configDir, "detected_java.json")
	
	return &Manager{
		detector:     NewDetector(),
		configDir:    configDir,
		javaListFile: javaListFile,
		javaList:     make([]*Java, 0),
	}
}

// DetectAndSave 检测Java并保存到文件
func (m *Manager) DetectAndSave(fullSearch bool) ([]*Java, error) {
	fmt.Println("正在检测Java环境...")
	start := time.Now()
	
	javaList, err := m.detector.DetectJava(fullSearch)
	if err != nil {
		return nil, fmt.Errorf("Java检测失败: %w", err)
	}
	
	// 排序Java列表
	SortJavaList(javaList, true) // 按版本降序排序
	
	// 保存到文件
	if err := SaveJavaList(javaList, m.javaListFile); err != nil {
		return nil, fmt.Errorf("保存Java列表失败: %w", err)
	}
	
	m.javaList = javaList
	
	duration := time.Since(start)
	fmt.Printf("Java检测完成，耗时: %v，找到 %d 个Java版本\n", duration, len(javaList))
	
	return javaList, nil
}

// LoadJavaList 加载已保存的Java列表
func (m *Manager) LoadJavaList() ([]*Java, error) {
	javaList, err := LoadJavaList(m.javaListFile)
	if err != nil {
		return nil, fmt.Errorf("加载Java列表失败: %w", err)
	}
	
	m.javaList = javaList
	return javaList, nil
}

// GetJavaList 获取当前Java列表
func (m *Manager) GetJavaList() []*Java {
	return m.javaList
}

// ValidateJavaList 验证Java列表中的Java是否仍然可用
func (m *Manager) ValidateJavaList() ([]*Java, []*Java) {
	var validJava []*Java
	var invalidJava []*Java
	
	for _, java := range m.javaList {
		if m.detector.CheckJavaAvailability(java) {
			validJava = append(validJava, java)
		} else {
			invalidJava = append(invalidJava, java)
		}
	}
	
	// 更新有效的Java列表
	m.javaList = validJava
	
	// 保存更新后的列表
	if len(invalidJava) > 0 {
		SaveJavaList(validJava, m.javaListFile)
	}
	
	return validJava, invalidJava
}

// FindJavaByVersion 根据版本查找Java
func (m *Manager) FindJavaByVersion(version string) *Java {
	for _, java := range m.javaList {
		if java.Version == version {
			return java
		}
	}
	return nil
}

// FindJavaByPath 根据路径查找Java
func (m *Manager) FindJavaByPath(path string) *Java {
	for _, java := range m.javaList {
		if java.Path == path {
			return java
		}
	}
	return nil
}

// GetBestJava 获取最佳Java版本（最新版本）
func (m *Manager) GetBestJava() *Java {
	if len(m.javaList) == 0 {
		return nil
	}
	
	// 假设列表已经按版本排序，返回第一个
	return m.javaList[0]
}

// GetJavaForMinecraft 获取适合Minecraft的Java版本
func (m *Manager) GetJavaForMinecraft(mcVersion string) *Java {
	// 根据Minecraft版本推荐Java版本
	// 这里可以根据实际需求实现更复杂的逻辑
	
	// 简单的版本映射
	var recommendedJavaVersion string
	switch {
	case mcVersion >= "1.17":
		recommendedJavaVersion = "17" // Java 17+
	case mcVersion >= "1.12":
		recommendedJavaVersion = "8" // Java 8+
	default:
		recommendedJavaVersion = "8" // Java 8
	}
	
	// 查找匹配的Java版本
	for _, java := range m.javaList {
		if java.Version >= recommendedJavaVersion {
			return java
		}
	}
	
	// 如果没有找到推荐版本，返回最新版本
	return m.GetBestJava()
}

// AddJava 手动添加Java
func (m *Manager) AddJava(javaPath string) (*Java, error) {
	version := m.detector.GetJavaVersion(javaPath)
	if version == "" {
		return nil, fmt.Errorf("无法获取Java版本: %s", javaPath)
	}
	
	java := &Java{
		Path:    javaPath,
		Version: version,
	}
	
	// 检查是否已存在
	for _, existing := range m.javaList {
		if existing.Equal(java) {
			return existing, fmt.Errorf("Java已存在: %s", javaPath)
		}
	}
	
	// 添加到列表
	m.javaList = append(m.javaList, java)
	
	// 重新排序
	SortJavaList(m.javaList, true)
	
	// 保存到文件
	if err := SaveJavaList(m.javaList, m.javaListFile); err != nil {
		return nil, fmt.Errorf("保存Java列表失败: %w", err)
	}
	
	return java, nil
}

// RemoveJava 移除Java
func (m *Manager) RemoveJava(javaPath string) error {
	var newList []*Java
	found := false
	
	for _, java := range m.javaList {
		if java.Path != javaPath {
			newList = append(newList, java)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("Java不存在: %s", javaPath)
	}
	
	m.javaList = newList
	
	// 保存到文件
	if err := SaveJavaList(m.javaList, m.javaListFile); err != nil {
		return fmt.Errorf("保存Java列表失败: %w", err)
	}
	
	return nil
}

// PrintJavaList 打印Java列表
func (m *Manager) PrintJavaList() {
	if len(m.javaList) == 0 {
		fmt.Println("未找到Java安装")
		return
	}
	
	fmt.Printf("找到 %d 个Java版本:\n", len(m.javaList))
	fmt.Println("序号 | 版本    | 路径")
	fmt.Println("-----|---------|----")
	
	for i, java := range m.javaList {
		status := "✓"
		if !m.detector.CheckJavaAvailability(java) {
			status = "✗"
		}
		fmt.Printf("%-4d | %-7s | %s %s\n", i+1, java.Version, java.Path, status)
	}
	
	fmt.Println("\n✓ = 可用, ✗ = 不可用")
}

// GetJavaInfo 获取Java详细信息
func (m *Manager) GetJavaInfo(java *Java) map[string]interface{} {
	info := map[string]interface{}{
		"path":      java.Path,
		"version":   java.Version,
		"available": m.detector.CheckJavaAvailability(java),
	}
	
	// 可以添加更多信息，如架构、供应商等
	return info
}
