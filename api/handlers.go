package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"easilypanel5/config"
	"easilypanel5/frp"
	"easilypanel5/server"
	"easilypanel5/utils"
)

// handleJavaDetect 处理Java环境检测
func handleJavaDetect(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	javaInfo, err := server.GetJavaInfo()
	if err != nil {
		WriteStandardError(w, "JAVA_DETECTION_FAILED", fmt.Sprintf("Failed to detect Java: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, javaInfo)
}

// handleJavaConfig 处理Java配置
func handleJavaConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := config.Get()
		response := map[string]interface{}{
			"java_path":    cfg.Java.JavaPath,
			"auto_detect":  cfg.Java.AutoDetect,
			"default_args": cfg.Java.DefaultArgs,
			"min_version":  cfg.Java.MinVersion,
			"max_memory":   cfg.Java.MaxMemory,
		}
		writeJSONResponse(w, response)

	case http.MethodPost:
		var req struct {
			JavaPath   string   `json:"java_path"`
			AutoDetect bool     `json:"auto_detect"`
			DefaultArgs []string `json:"default_args"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// 数据验证
		var validationErrors utils.ValidationErrors

		// 验证Java路径
		if req.JavaPath != "" {
			if err := utils.ValidateFilePath("java_path", req.JavaPath); err != nil {
				if valErr, ok := err.(utils.ValidationError); ok {
					validationErrors = append(validationErrors, valErr)
				}
			} else {
				// 进一步验证Java路径是否有效
				if _, err := server.CheckJavaPath(req.JavaPath); err != nil {
					validationErrors = append(validationErrors, utils.ValidationError{
						Field:   "java_path",
						Message: fmt.Sprintf("Invalid Java path: %v", err),
						Code:    "INVALID_JAVA_PATH",
					})
				}
			}
		}

		// 验证Java参数
		if len(req.DefaultArgs) > 0 {
			if err := utils.ValidateJavaArgs("default_args", req.DefaultArgs); err != nil {
				if valErr, ok := err.(utils.ValidationError); ok {
					validationErrors = append(validationErrors, valErr)
				}
			}
		}

		// 如果有验证错误，返回错误信息
		if len(validationErrors) > 0 {
			WriteStandardError(w, "VALIDATION_FAILED", validationErrors.Error(), http.StatusBadRequest)
			return
		}

		// 更新配置
		cfg := config.Get()
		cfg.Java.JavaPath = req.JavaPath
		cfg.Java.AutoDetect = req.AutoDetect
		if len(req.DefaultArgs) > 0 {
			cfg.Java.DefaultArgs = req.DefaultArgs
		}

		if err := config.Update(cfg); err != nil {
			WriteStandardError(w, "CONFIG_UPDATE_FAILED", "Failed to update config", http.StatusInternalServerError)
			return
		}

		WriteStandardResponse(w, map[string]string{"status": "ok"})

	default:
		WriteStandardError(w, "METHOD_NOT_ALLOWED", "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCoresList 处理核心列表查询
func handleCoresList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cores, err := server.GetCoresList()
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Failed to get cores list: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, cores)
}

// handleCoresVersions 处理核心版本查询
func handleCoresVersions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	coreType := r.URL.Query().Get("type")
	if coreType == "" {
		writeErrorResponse(w, "Missing core type parameter", http.StatusBadRequest)
		return
	}

	versions, err := server.GetCoreVersions(coreType)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Failed to get core versions: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, versions)
}

// handleCoresDownload 处理核心下载
func handleCoresDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		CoreType    string `json:"core_type"`
		MCVersion   string `json:"mc_version"`
		CoreVersion string `json:"core_version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.CoreType == "" || req.MCVersion == "" || req.CoreVersion == "" {
		writeErrorResponse(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	taskID, err := server.StartDownload(req.CoreType, req.MCVersion, req.CoreVersion)
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Failed to start download: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"task_id": taskID,
		"status":  "started",
	}
	writeJSONResponse(w, response)
}

// handleServers 处理服务器列表
func handleServers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		servers := config.GetServers().GetAll()
		writeJSONResponse(w, servers)

	case http.MethodPost:
		var req config.MinecraftServer
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// 数据验证
		var validationErrors utils.ValidationErrors

		// 验证服务器名称
		if err := utils.ValidateServerName("name", req.Name); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}

		// 验证核心类型
		if err := utils.ValidateCoreType("core_type", req.CoreType); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}

		// 验证Minecraft版本
		if err := utils.ValidateMinecraftVersion("mc_version", req.MCVersion); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}

		// 验证内存大小
		if req.Memory > 0 {
			if err := utils.ValidateMemorySize("memory", req.Memory); err != nil {
				if valErr, ok := err.(utils.ValidationError); ok {
					validationErrors = append(validationErrors, valErr)
				}
			}
		}

		// 验证端口
		if req.Port > 0 {
			if err := utils.ValidatePort("port", req.Port); err != nil {
				if valErr, ok := err.(utils.ValidationError); ok {
					validationErrors = append(validationErrors, valErr)
				}
			}
		}

		// 如果有验证错误，返回错误信息
		if len(validationErrors) > 0 {
			WriteStandardError(w, "VALIDATION_FAILED", validationErrors.Error(), http.StatusBadRequest)
			return
		}

		// 创建服务器
		serverID, err := server.CreateServer(&req)
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to create server: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]string{
			"id":     serverID,
			"status": "created",
		}
		writeJSONResponse(w, response)

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleServerDetail 处理单个服务器操作
func handleServerDetail(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中提取服务器ID
	urlPath := strings.TrimPrefix(r.URL.Path, "/api/servers/")
	parts := strings.Split(urlPath, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeErrorResponse(w, "Missing server ID", http.StatusBadRequest)
		return
	}

	serverID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	// 验证服务器是否存在
	srv, exists := config.GetServers().Get(serverID)
	if !exists {
		writeErrorResponse(w, "Server not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if action == "" {
			// 获取服务器详情
			writeJSONResponse(w, srv)
		} else {
			writeErrorResponse(w, "Invalid action", http.StatusBadRequest)
		}

	case http.MethodPost:
		switch action {
		case "start":
			if err := server.StartServer(serverID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to start server: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "starting"})

		case "stop":
			if err := server.StopServer(serverID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to stop server: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "stopping"})

		case "restart":
			if err := server.RestartServer(serverID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to restart server: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "restarting"})

		case "command":
			var req struct {
				Command string `json:"command"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if err := server.SendCommand(serverID, req.Command); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to send command: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "sent"})

		default:
			writeErrorResponse(w, "Invalid action", http.StatusBadRequest)
		}

	case http.MethodDelete:
		if action == "" {
			// 删除服务器
			if err := server.DeleteServer(serverID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to delete server: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "deleted"})
		} else {
			writeErrorResponse(w, "Invalid action", http.StatusBadRequest)
		}

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleDaemonStatus 处理守护管理器状态查询
func handleDaemonStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.Get()

	response := map[string]interface{}{
		"enabled":              cfg.Daemon.EnableAutoRestart,
		"max_restart_attempts": cfg.Daemon.MaxRestartAttempts,
		"restart_delay_ms":     cfg.Daemon.RestartDelay.Milliseconds(),
		"resource_monitoring":  cfg.Daemon.ResourceMonitoring,
		"monitor_interval_ms":  cfg.Daemon.MonitorInterval.Milliseconds(),
		"log_rotation":         cfg.Daemon.LogRotation,
		"max_log_size":         cfg.Daemon.MaxLogSize,
		"max_log_files":        cfg.Daemon.MaxLogFiles,
	}

	writeJSONResponse(w, response)
}

// handleDaemonStats 处理守护统计信息查询
func handleDaemonStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	daemon := server.GetDaemon()
	stats := daemon.GetAllProcessStats()

	writeJSONResponse(w, stats)
}

// handleFRPStatus 处理FRP状态查询
func handleFRPStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := config.Get()

	response := map[string]interface{}{
		"enabled":             cfg.FRP.Enabled,
		"api_endpoint":        cfg.FRP.APIEndpoint,
		"default_node":        cfg.FRP.DefaultNode,
		"auto_start":          cfg.FRP.AutoStart,
		"auto_restart":        cfg.FRP.AutoRestart,
		"max_tunnels":         cfg.FRP.MaxTunnels,
		"monitor_interval_ms": cfg.FRP.MonitorInterval.Milliseconds(),
		"stats_retention_ms":  cfg.FRP.StatsRetention.Milliseconds(),
		"default_bandwidth":   cfg.FRP.DefaultBandwidth,
		"default_compression": cfg.FRP.DefaultCompression,
		"default_encryption":  cfg.FRP.DefaultEncryption,
		"is_running":          frp.IsManagerRunning(),
	}

	writeJSONResponse(w, response)
}

// handleFRPNodes 处理FRP节点查询
func handleFRPNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes, err := frp.GetNodes()
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Failed to get nodes: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, nodes)
}

// handleFRPTunnels 处理FRP隧道列表
func handleFRPTunnels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		tunnels, err := frp.GetAllTunnels()
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to get tunnels: %v", err), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, tunnels)

	case http.MethodPost:
		var req frp.TunnelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
			return
		}

		// 数据验证
		var validationErrors utils.ValidationErrors

		// 验证隧道名称
		if err := utils.ValidateRequired("name", req.Name); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		} else if err := utils.ValidateStringLength("name", req.Name, 1, 50); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}

		// 验证隧道类型
		if err := utils.ValidateRequired("type", req.Type); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		} else {
			validTypes := []string{"tcp", "udp", "http", "https"}
			if err := utils.ValidateStringInSlice("type", req.Type, validTypes); err != nil {
				if valErr, ok := err.(utils.ValidationError); ok {
					validationErrors = append(validationErrors, valErr)
				}
			}
		}

		// 验证Token
		if err := utils.ValidateRequired("token", req.Token); err != nil {
			if valErr, ok := err.(utils.ValidationError); ok {
				validationErrors = append(validationErrors, valErr)
			}
		}

		// 如果有验证错误，返回错误信息
		if len(validationErrors) > 0 {
			WriteStandardError(w, "VALIDATION_FAILED", validationErrors.Error(), http.StatusBadRequest)
			return
		}

		tunnel, err := frp.CreateTunnel(&req)
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to create tunnel: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, tunnel)

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleFRPTunnelAction 处理单个FRP隧道操作
func handleFRPTunnelAction(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中提取隧道ID
	urlPath := strings.TrimPrefix(r.URL.Path, "/api/frp/tunnels/")
	parts := strings.Split(urlPath, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeErrorResponse(w, "Missing tunnel ID", http.StatusBadRequest)
		return
	}

	tunnelID := parts[0]
	action := ""
	if len(parts) > 1 {
		action = parts[1]
	}

	switch r.Method {
	case http.MethodGet:
		if action == "" {
			// 获取隧道详情
			tunnel, err := frp.GetTunnel(tunnelID)
			if err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to get tunnel: %v", err), http.StatusNotFound)
				return
			}
			writeJSONResponse(w, tunnel)
		} else if action == "stats" {
			// 获取隧道统计
			stats, err := frp.GetTunnelStats(tunnelID)
			if err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to get tunnel stats: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, stats)
		} else {
			writeErrorResponse(w, "Invalid action", http.StatusBadRequest)
		}

	case http.MethodPost:
		switch action {
		case "start":
			if err := frp.StartTunnel(tunnelID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to start tunnel: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "starting"})

		case "stop":
			if err := frp.StopTunnel(tunnelID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to stop tunnel: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "stopping"})

		case "restart":
			if err := frp.RestartTunnel(tunnelID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to restart tunnel: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "restarting"})

		default:
			writeErrorResponse(w, "Invalid action", http.StatusBadRequest)
		}

	case http.MethodDelete:
		if action == "" {
			// 删除隧道
			if err := frp.DeleteTunnel(tunnelID); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Failed to delete tunnel: %v", err), http.StatusInternalServerError)
				return
			}
			writeJSONResponse(w, map[string]string{"status": "deleted"})
		} else {
			writeErrorResponse(w, "Invalid action", http.StatusBadRequest)
		}

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTemplates 处理模板列表
func handleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		templates := server.GetTemplateManager().GetAllTemplates()
		writeJSONResponse(w, templates)

	case http.MethodPost:
		var template config.ServerTemplate
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := server.GetTemplateManager().CreateTemplate(&template); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to create template: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, template)

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTemplateAction 处理单个模板操作
func handleTemplateAction(w http.ResponseWriter, r *http.Request) {
	templateID := strings.TrimPrefix(r.URL.Path, "/api/templates/")
	if templateID == "" {
		writeErrorResponse(w, "Missing template ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		template, err := server.GetTemplateManager().GetTemplate(templateID)
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Template not found: %v", err), http.StatusNotFound)
			return
		}
		writeJSONResponse(w, template)

	case http.MethodPut:
		var template config.ServerTemplate
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		template.ID = templateID
		if err := server.GetTemplateManager().UpdateTemplate(&template); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to update template: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, template)

	case http.MethodDelete:
		if err := server.GetTemplateManager().DeleteTemplate(templateID); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to delete template: %v", err), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "deleted"})

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGroups 处理分组列表
func handleGroups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		groups := server.GetGroupManager().GetAllGroups()
		writeJSONResponse(w, groups)

	case http.MethodPost:
		var group config.ServerGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := server.GetGroupManager().CreateGroup(&group); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to create group: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, group)

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGroupAction 处理单个分组操作
func handleGroupAction(w http.ResponseWriter, r *http.Request) {
	groupID := strings.TrimPrefix(r.URL.Path, "/api/groups/")
	if groupID == "" {
		writeErrorResponse(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		group, err := server.GetGroupManager().GetGroup(groupID)
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Group not found: %v", err), http.StatusNotFound)
			return
		}
		writeJSONResponse(w, group)

	case http.MethodPut:
		var group config.ServerGroup
		if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		group.ID = groupID
		if err := server.GetGroupManager().UpdateGroup(&group); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to update group: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, group)

	case http.MethodDelete:
		if err := server.GetGroupManager().DeleteGroup(groupID); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to delete group: %v", err), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "deleted"})

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleBatchOperations 处理批量操作
func handleBatchOperations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		operations := server.GetBatchManager().GetAllBatchOperations()
		writeJSONResponse(w, operations)

	case http.MethodPost:
		var req struct {
			Type      string   `json:"type"`
			ServerIDs []string `json:"server_ids"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		operation, err := server.GetBatchManager().StartBatchOperation(req.Type, req.ServerIDs)
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to start batch operation: %v", err), http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, operation)

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleBatchAction 处理单个批量操作
func handleBatchAction(w http.ResponseWriter, r *http.Request) {
	operationID := strings.TrimPrefix(r.URL.Path, "/api/batch/")
	if operationID == "" {
		writeErrorResponse(w, "Missing operation ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		operation, err := server.GetBatchManager().GetBatchOperation(operationID)
		if err != nil {
			writeErrorResponse(w, fmt.Sprintf("Operation not found: %v", err), http.StatusNotFound)
			return
		}
		writeJSONResponse(w, operation)

	case http.MethodDelete:
		if err := server.GetBatchManager().CancelBatchOperation(operationID); err != nil {
			writeErrorResponse(w, fmt.Sprintf("Failed to cancel operation: %v", err), http.StatusInternalServerError)
			return
		}
		writeJSONResponse(w, map[string]string{"status": "cancelled"})

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleServerLogs 处理服务器日志请求
func handleServerLogs(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中提取服务器ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteStandardError(w, "INVALID_PATH", "Invalid server ID", http.StatusBadRequest)
		return
	}
	serverID := pathParts[3]

	// 获取行数参数
	linesStr := r.URL.Query().Get("lines")
	lines := 100 // 默认100行
	if linesStr != "" {
		if parsedLines, err := strconv.Atoi(linesStr); err == nil && parsedLines > 0 {
			lines = parsedLines
		}
	}

	// 获取服务器日志
	logs, err := server.GetServerLogs(serverID, lines)
	if err != nil {
		WriteStandardError(w, "GET_LOGS_FAILED", fmt.Sprintf("Failed to get logs: %v", err), http.StatusInternalServerError)
		return
	}

	// 转换为结构化格式
	var logEntries []map[string]interface{}
	for _, logLine := range logs {
		logEntries = append(logEntries, map[string]interface{}{
			"message":   logLine,
			"level":     "INFO", // 简化处理，实际可以解析日志级别
			"timestamp": time.Now(), // 简化处理，实际应该解析时间戳
		})
	}

	WriteStandardResponse(w, logEntries)
}

// handleServerCommand 处理服务器命令请求
func handleServerCommand(w http.ResponseWriter, r *http.Request) {
	if err := ValidateMethod(r, http.MethodPost); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中提取服务器ID
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		WriteStandardError(w, "INVALID_PATH", "Invalid server ID", http.StatusBadRequest)
		return
	}
	serverID := pathParts[3]

	// 解析请求体
	var req struct {
		Command string `json:"command"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteStandardError(w, "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// 验证命令
	if err := utils.ValidateRequired("command", req.Command); err != nil {
		if valErr, ok := err.(utils.ValidationError); ok {
			WriteStandardError(w, "VALIDATION_FAILED", valErr.Message, http.StatusBadRequest)
			return
		}
	}

	// 发送命令到服务器
	err := server.SendCommand(serverID, req.Command)
	if err != nil {
		WriteStandardError(w, "SEND_COMMAND_FAILED", fmt.Sprintf("Failed to send command: %v", err), http.StatusInternalServerError)
		return
	}

	WriteStandardResponse(w, map[string]string{"status": "success"})
}

// writeJSON 写入JSON数据
func writeJSON(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
