package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"easilypanel5/config"
	"easilypanel5/server"
)

// handleJavaDetect 处理Java环境检测
func handleJavaDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	javaInfo, err := server.GetJavaInfo()
	if err != nil {
		writeErrorResponse(w, fmt.Sprintf("Failed to detect Java: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSONResponse(w, javaInfo)
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
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 验证Java路径
		if req.JavaPath != "" {
			if _, err := server.CheckJavaPath(req.JavaPath); err != nil {
				writeErrorResponse(w, fmt.Sprintf("Invalid Java path: %v", err), http.StatusBadRequest)
				return
			}
		}

		// 更新配置
		cfg := config.Get()
		cfg.Java.JavaPath = req.JavaPath
		cfg.Java.AutoDetect = req.AutoDetect
		if len(req.DefaultArgs) > 0 {
			cfg.Java.DefaultArgs = req.DefaultArgs
		}

		if err := config.Update(cfg); err != nil {
			writeErrorResponse(w, "Failed to update config", http.StatusInternalServerError)
			return
		}

		writeJSONResponse(w, map[string]string{"status": "ok"})

	default:
		writeErrorResponse(w, "Method not allowed", http.StatusMethodNotAllowed)
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
			writeErrorResponse(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// 验证必要字段
		if req.Name == "" || req.CoreType == "" || req.MCVersion == "" {
			writeErrorResponse(w, "Missing required fields", http.StatusBadRequest)
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

// writeJSON 写入JSON数据
func writeJSON(w http.ResponseWriter, data interface{}) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}
