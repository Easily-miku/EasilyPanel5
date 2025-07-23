package server

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"easilypanel5/config"
	"easilypanel5/utils"
)

// TemplateManager 模板管理器
type TemplateManager struct {
	templates map[string]*config.ServerTemplate
	mutex     sync.RWMutex
	file      string
}

var (
	globalTemplateManager *TemplateManager
	templateOnce          sync.Once
)

// GetTemplateManager 获取全局模板管理器
func GetTemplateManager() *TemplateManager {
	templateOnce.Do(func() {
		globalTemplateManager = &TemplateManager{
			templates: make(map[string]*config.ServerTemplate),
			file:      "data/templates.json",
		}
		globalTemplateManager.load()
		globalTemplateManager.createDefaultTemplates()
	})
	return globalTemplateManager
}

// CreateTemplate 创建模板
func (tm *TemplateManager) CreateTemplate(template *config.ServerTemplate) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if template.ID == "" {
		template.ID = fmt.Sprintf("template_%d", time.Now().Unix())
	}

	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()

	tm.templates[template.ID] = template
	return tm.save()
}

// GetTemplate 获取模板
func (tm *TemplateManager) GetTemplate(id string) (*config.ServerTemplate, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	template, exists := tm.templates[id]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", id)
	}

	return template, nil
}

// GetAllTemplates 获取所有模板
func (tm *TemplateManager) GetAllTemplates() []*config.ServerTemplate {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	templates := make([]*config.ServerTemplate, 0, len(tm.templates))
	for _, template := range tm.templates {
		templates = append(templates, template)
	}

	return templates
}

// UpdateTemplate 更新模板
func (tm *TemplateManager) UpdateTemplate(template *config.ServerTemplate) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.templates[template.ID]; !exists {
		return fmt.Errorf("template not found: %s", template.ID)
	}

	template.UpdatedAt = time.Now()
	tm.templates[template.ID] = template
	return tm.save()
}

// DeleteTemplate 删除模板
func (tm *TemplateManager) DeleteTemplate(id string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	template, exists := tm.templates[id]
	if !exists {
		return fmt.Errorf("template not found: %s", id)
	}

	// 不能删除默认模板
	if template.IsDefault {
		return fmt.Errorf("cannot delete default template")
	}

	delete(tm.templates, id)
	return tm.save()
}

// CreateServerFromTemplate 从模板创建服务器
func (tm *TemplateManager) CreateServerFromTemplate(templateID, serverName string) (*config.MinecraftServer, error) {
	template, err := tm.GetTemplate(templateID)
	if err != nil {
		return nil, err
	}

	server := &config.MinecraftServer{
		Name:        serverName,
		CoreType:    template.CoreType,
		MCVersion:   template.MCVersion,
		Memory:      template.Memory,
		Port:        25565, // 默认端口，后续可以自动分配
		JavaArgs:    make([]string, len(template.JavaArgs)),
		Properties:  make(map[string]string),
		TemplateID:  templateID,
		Status:      config.StatusStopped,
		AutoStart:   false,
		AutoRestart: true,
		DaemonEnabled: true,
		MonitoringEnabled: true,
		BackupEnabled: false,
		Group:       "default",
		Tags:        []string{},
		CustomConfig: make(map[string]interface{}),
	}

	// 复制Java参数
	copy(server.JavaArgs, template.JavaArgs)

	// 复制属性
	for k, v := range template.Properties {
		server.Properties[k] = v
	}

	// 复制自定义配置
	if template.Config != nil {
		for k, v := range template.Config {
			server.CustomConfig[k] = v
		}
	}

	return server, nil
}

// load 加载模板
func (tm *TemplateManager) load() error {
	if _, err := os.Stat(tm.file); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(tm.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &tm.templates)
}

// save 保存模板
func (tm *TemplateManager) save() error {
	if err := os.MkdirAll(filepath.Dir(tm.file), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(tm.templates, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tm.file, data, 0644)
}

// createDefaultTemplates 创建默认模板
func (tm *TemplateManager) createDefaultTemplates() {
	// 检查是否已有默认模板
	for _, template := range tm.templates {
		if template.IsDefault {
			return
		}
	}

	// 创建默认模板
	defaultTemplates := []*config.ServerTemplate{
		{
			ID:          "vanilla_latest",
			Name:        "Vanilla Latest",
			Description: "最新版本的原版Minecraft服务器",
			CoreType:    "vanilla",
			MCVersion:   "1.20.1",
			Memory:      2048,
			JavaArgs:    []string{"-Xms1G", "-Xmx2G", "-XX:+UseG1GC"},
			Properties: map[string]string{
				"server-port":      "25565",
				"max-players":      "20",
				"difficulty":       "normal",
				"gamemode":         "survival",
				"pvp":             "true",
				"online-mode":     "true",
				"white-list":      "false",
			},
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "paper_latest",
			Name:        "Paper Latest",
			Description: "最新版本的Paper服务器（高性能）",
			CoreType:    "paper",
			MCVersion:   "1.20.1",
			Memory:      2048,
			JavaArgs:    []string{"-Xms1G", "-Xmx2G", "-XX:+UseG1GC", "-XX:+ParallelRefProcEnabled"},
			Properties: map[string]string{
				"server-port":      "25565",
				"max-players":      "50",
				"difficulty":       "normal",
				"gamemode":         "survival",
				"pvp":             "true",
				"online-mode":     "true",
				"white-list":      "false",
			},
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "fabric_latest",
			Name:        "Fabric Latest",
			Description: "最新版本的Fabric模组服务器",
			CoreType:    "fabric",
			MCVersion:   "1.20.1",
			Memory:      3072,
			JavaArgs:    []string{"-Xms1G", "-Xmx3G", "-XX:+UseG1GC"},
			Properties: map[string]string{
				"server-port":      "25565",
				"max-players":      "30",
				"difficulty":       "normal",
				"gamemode":         "survival",
				"pvp":             "true",
				"online-mode":     "true",
				"white-list":      "false",
			},
			IsDefault: true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, template := range defaultTemplates {
		tm.templates[template.ID] = template
	}

	tm.save()
	utils.EmitEvent("templates_initialized", "", map[string]interface{}{
		"count": len(defaultTemplates),
	})
}
