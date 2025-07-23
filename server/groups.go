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

// GroupManager 分组管理器
type GroupManager struct {
	groups map[string]*config.ServerGroup
	mutex  sync.RWMutex
	file   string
}

var (
	globalGroupManager *GroupManager
	groupOnce          sync.Once
)

// GetGroupManager 获取全局分组管理器
func GetGroupManager() *GroupManager {
	groupOnce.Do(func() {
		globalGroupManager = &GroupManager{
			groups: make(map[string]*config.ServerGroup),
			file:   "data/groups.json",
		}
		globalGroupManager.load()
		globalGroupManager.createDefaultGroups()
	})
	return globalGroupManager
}

// CreateGroup 创建分组
func (gm *GroupManager) CreateGroup(group *config.ServerGroup) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if group.ID == "" {
		group.ID = fmt.Sprintf("group_%d", time.Now().Unix())
	}

	// 检查名称是否重复
	for _, existingGroup := range gm.groups {
		if existingGroup.Name == group.Name {
			return fmt.Errorf("group name already exists: %s", group.Name)
		}
	}

	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()
	group.ServerCount = 0

	gm.groups[group.ID] = group
	
	utils.EmitEvent("group_created", "", group)
	return gm.save()
}

// GetGroup 获取分组
func (gm *GroupManager) GetGroup(id string) (*config.ServerGroup, error) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	group, exists := gm.groups[id]
	if !exists {
		return nil, fmt.Errorf("group not found: %s", id)
	}

	return group, nil
}

// GetAllGroups 获取所有分组
func (gm *GroupManager) GetAllGroups() []*config.ServerGroup {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	groups := make([]*config.ServerGroup, 0, len(gm.groups))
	for _, group := range gm.groups {
		groups = append(groups, group)
	}

	return groups
}

// UpdateGroup 更新分组
func (gm *GroupManager) UpdateGroup(group *config.ServerGroup) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	if _, exists := gm.groups[group.ID]; !exists {
		return fmt.Errorf("group not found: %s", group.ID)
	}

	// 检查名称是否与其他分组重复
	for id, existingGroup := range gm.groups {
		if id != group.ID && existingGroup.Name == group.Name {
			return fmt.Errorf("group name already exists: %s", group.Name)
		}
	}

	group.UpdatedAt = time.Now()
	gm.groups[group.ID] = group
	
	utils.EmitEvent("group_updated", "", group)
	return gm.save()
}

// DeleteGroup 删除分组
func (gm *GroupManager) DeleteGroup(id string) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	group, exists := gm.groups[id]
	if !exists {
		return fmt.Errorf("group not found: %s", id)
	}

	// 检查是否有服务器在使用此分组
	if group.ServerCount > 0 {
		return fmt.Errorf("cannot delete group with servers: %s", group.Name)
	}

	// 不能删除默认分组
	if id == "default" {
		return fmt.Errorf("cannot delete default group")
	}

	delete(gm.groups, id)
	
	utils.EmitEvent("group_deleted", "", map[string]interface{}{
		"group_id": id,
		"name":     group.Name,
	})
	return gm.save()
}

// UpdateServerCount 更新分组中的服务器数量
func (gm *GroupManager) UpdateServerCount(groupID string, delta int) error {
	gm.mutex.Lock()
	defer gm.mutex.Unlock()

	group, exists := gm.groups[groupID]
	if !exists {
		return fmt.Errorf("group not found: %s", groupID)
	}

	group.ServerCount += delta
	if group.ServerCount < 0 {
		group.ServerCount = 0
	}

	group.UpdatedAt = time.Now()
	return gm.save()
}

// GetGroupByName 根据名称获取分组
func (gm *GroupManager) GetGroupByName(name string) (*config.ServerGroup, error) {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	for _, group := range gm.groups {
		if group.Name == name {
			return group, nil
		}
	}

	return nil, fmt.Errorf("group not found: %s", name)
}

// GetGroupStats 获取分组统计信息
func (gm *GroupManager) GetGroupStats() map[string]interface{} {
	gm.mutex.RLock()
	defer gm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total_groups":   len(gm.groups),
		"total_servers":  0,
		"groups":         make([]map[string]interface{}, 0),
	}

	totalServers := 0
	for _, group := range gm.groups {
		totalServers += group.ServerCount
		
		groupInfo := map[string]interface{}{
			"id":           group.ID,
			"name":         group.Name,
			"server_count": group.ServerCount,
			"color":        group.Color,
			"icon":         group.Icon,
		}
		
		stats["groups"] = append(stats["groups"].([]map[string]interface{}), groupInfo)
	}

	stats["total_servers"] = totalServers
	return stats
}

// load 加载分组
func (gm *GroupManager) load() error {
	if _, err := os.Stat(gm.file); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(gm.file)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &gm.groups)
}

// save 保存分组
func (gm *GroupManager) save() error {
	if err := os.MkdirAll(filepath.Dir(gm.file), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(gm.groups, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(gm.file, data, 0644)
}

// createDefaultGroups 创建默认分组
func (gm *GroupManager) createDefaultGroups() {
	// 检查是否已有默认分组
	if _, exists := gm.groups["default"]; exists {
		return
	}

	// 创建默认分组
	defaultGroups := []*config.ServerGroup{
		{
			ID:          "default",
			Name:        "默认分组",
			Description: "默认服务器分组",
			Color:       "#6366f1",
			Icon:        "server",
			ServerCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "production",
			Name:        "生产环境",
			Description: "生产环境服务器",
			Color:       "#ef4444",
			Icon:        "shield-check",
			ServerCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "development",
			Name:        "开发环境",
			Description: "开发测试服务器",
			Color:       "#10b981",
			Icon:        "code",
			ServerCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "modded",
			Name:        "模组服务器",
			Description: "安装了模组的服务器",
			Color:       "#f59e0b",
			Icon:        "puzzle",
			ServerCount: 0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	for _, group := range defaultGroups {
		gm.groups[group.ID] = group
	}

	gm.save()
	utils.EmitEvent("groups_initialized", "", map[string]interface{}{
		"count": len(defaultGroups),
	})
}
