package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"easilypanel5/config"
	"easilypanel5/utils"
)

// BatchManager 批量操作管理器
type BatchManager struct {
	operations map[string]*config.BatchOperation
	mutex      sync.RWMutex
}

var (
	globalBatchManager *BatchManager
	batchOnce          sync.Once
)

// GetBatchManager 获取全局批量操作管理器
func GetBatchManager() *BatchManager {
	batchOnce.Do(func() {
		globalBatchManager = &BatchManager{
			operations: make(map[string]*config.BatchOperation),
		}
	})
	return globalBatchManager
}

// StartBatchOperation 开始批量操作
func (bm *BatchManager) StartBatchOperation(operationType string, serverIDs []string) (*config.BatchOperation, error) {
	if len(serverIDs) == 0 {
		return nil, fmt.Errorf("no servers specified")
	}

	// 验证操作类型
	validTypes := map[string]bool{
		"start":   true,
		"stop":    true,
		"restart": true,
		"delete":  true,
	}
	
	if !validTypes[operationType] {
		return nil, fmt.Errorf("invalid operation type: %s", operationType)
	}

	// 验证服务器ID
	servers := config.GetServers()
	for _, serverID := range serverIDs {
		if _, exists := servers.Get(serverID); !exists {
			return nil, fmt.Errorf("server not found: %s", serverID)
		}
	}

	// 创建批量操作
	operation := &config.BatchOperation{
		ID:        fmt.Sprintf("batch_%d", time.Now().Unix()),
		Type:      operationType,
		ServerIDs: serverIDs,
		Status:    "pending",
		Progress:  0,
		Results:   make(map[string]interface{}),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	bm.mutex.Lock()
	bm.operations[operation.ID] = operation
	bm.mutex.Unlock()

	// 异步执行操作
	go bm.executeBatchOperation(operation)

	utils.EmitEvent("batch_operation_started", "", operation)
	return operation, nil
}

// GetBatchOperation 获取批量操作
func (bm *BatchManager) GetBatchOperation(id string) (*config.BatchOperation, error) {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	operation, exists := bm.operations[id]
	if !exists {
		return nil, fmt.Errorf("batch operation not found: %s", id)
	}

	return operation, nil
}

// GetAllBatchOperations 获取所有批量操作
func (bm *BatchManager) GetAllBatchOperations() []*config.BatchOperation {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	operations := make([]*config.BatchOperation, 0, len(bm.operations))
	for _, operation := range bm.operations {
		operations = append(operations, operation)
	}

	return operations
}

// executeBatchOperation 执行批量操作
func (bm *BatchManager) executeBatchOperation(operation *config.BatchOperation) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Batch operation panic: %v", r)
			bm.updateOperationStatus(operation.ID, "failed", 0, fmt.Sprintf("Panic: %v", r))
		}
	}()

	log.Printf("开始执行批量操作: %s (%s)", operation.Type, operation.ID)

	bm.updateOperationStatus(operation.ID, "running", 0, "")

	totalServers := len(operation.ServerIDs)
	completed := 0
	results := make(map[string]interface{})

	for _, serverID := range operation.ServerIDs {
		log.Printf("对服务器 %s 执行 %s 操作", serverID, operation.Type)

		var err error
		switch operation.Type {
		case "start":
			err = StartServer(serverID)
		case "stop":
			err = StopServer(serverID)
		case "restart":
			err = RestartServer(serverID)
		case "delete":
			err = DeleteServer(serverID)
		}

		if err != nil {
			results[serverID] = map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			}
			log.Printf("服务器 %s 操作失败: %v", serverID, err)
		} else {
			results[serverID] = map[string]interface{}{
				"success": true,
			}
			log.Printf("服务器 %s 操作成功", serverID)
		}

		completed++
		progress := (completed * 100) / totalServers

		// 更新进度
		bm.mutex.Lock()
		if op, exists := bm.operations[operation.ID]; exists {
			op.Progress = progress
			op.Results = results
			op.UpdatedAt = time.Now()
		}
		bm.mutex.Unlock()

		// 发送进度更新事件
		utils.EmitEvent("batch_operation_progress", "", map[string]interface{}{
			"operation_id": operation.ID,
			"progress":     progress,
			"completed":    completed,
			"total":        totalServers,
		})

		// 操作间隔，避免系统负载过高
		if completed < totalServers {
			time.Sleep(1 * time.Second)
		}
	}

	// 完成操作
	status := "completed"
	errorCount := 0
	for _, result := range results {
		if resultMap, ok := result.(map[string]interface{}); ok {
			if success, ok := resultMap["success"].(bool); ok && !success {
				errorCount++
			}
		}
	}

	if errorCount > 0 {
		if errorCount == totalServers {
			status = "failed"
		} else {
			status = "partial_success"
		}
	}

	bm.updateOperationStatus(operation.ID, status, 100, "")

	log.Printf("批量操作完成: %s (%s) - 成功: %d, 失败: %d", 
		operation.Type, operation.ID, totalServers-errorCount, errorCount)

	utils.EmitEvent("batch_operation_completed", "", map[string]interface{}{
		"operation_id": operation.ID,
		"status":       status,
		"total":        totalServers,
		"success":      totalServers - errorCount,
		"failed":       errorCount,
	})
}

// updateOperationStatus 更新操作状态
func (bm *BatchManager) updateOperationStatus(id, status string, progress int, errorMsg string) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	if operation, exists := bm.operations[id]; exists {
		operation.Status = status
		operation.Progress = progress
		operation.UpdatedAt = time.Now()
		if errorMsg != "" {
			operation.Error = errorMsg
		}
	}
}

// CleanupOldOperations 清理旧的操作记录
func (bm *BatchManager) CleanupOldOperations(maxAge time.Duration) {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	cutoff := time.Now().Add(-maxAge)
	toDelete := make([]string, 0)

	for id, operation := range bm.operations {
		if operation.CreatedAt.Before(cutoff) && 
		   (operation.Status == "completed" || operation.Status == "failed" || operation.Status == "partial_success") {
			toDelete = append(toDelete, id)
		}
	}

	for _, id := range toDelete {
		delete(bm.operations, id)
	}

	if len(toDelete) > 0 {
		log.Printf("清理了 %d 个旧的批量操作记录", len(toDelete))
	}
}

// GetOperationStats 获取操作统计
func (bm *BatchManager) GetOperationStats() map[string]interface{} {
	bm.mutex.RLock()
	defer bm.mutex.RUnlock()

	stats := map[string]interface{}{
		"total":           len(bm.operations),
		"pending":         0,
		"running":         0,
		"completed":       0,
		"failed":          0,
		"partial_success": 0,
	}

	for _, operation := range bm.operations {
		if count, exists := stats[operation.Status]; exists {
			stats[operation.Status] = count.(int) + 1
		}
	}

	return stats
}

// CancelBatchOperation 取消批量操作（仅对pending状态有效）
func (bm *BatchManager) CancelBatchOperation(id string) error {
	bm.mutex.Lock()
	defer bm.mutex.Unlock()

	operation, exists := bm.operations[id]
	if !exists {
		return fmt.Errorf("batch operation not found: %s", id)
	}

	if operation.Status != "pending" {
		return fmt.Errorf("cannot cancel operation in status: %s", operation.Status)
	}

	operation.Status = "cancelled"
	operation.UpdatedAt = time.Now()

	utils.EmitEvent("batch_operation_cancelled", "", operation)
	return nil
}
