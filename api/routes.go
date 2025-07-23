package api

import (
	"net/http"
	"path/filepath"
)

// SetupRoutes 设置路由
func SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// 静态文件服务
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	
	// 主页
	mux.HandleFunc("/", serveIndex)

	// API路由
	mux.HandleFunc("/api/status", handleStatus)
	
	// Java环境相关
	mux.HandleFunc("/api/java/detect", handleJavaDetect)
	mux.HandleFunc("/api/java/config", handleJavaConfig)
	
	// 核心下载相关
	mux.HandleFunc("/api/cores/list", handleCoresList)
	mux.HandleFunc("/api/cores/versions", handleCoresVersions)
	mux.HandleFunc("/api/cores/download", handleCoresDownload)
	
	// 服务器管理相关
	mux.HandleFunc("/api/servers", handleServers)
	mux.HandleFunc("/api/servers/", handleServerDetail)

	// 守护管理
	mux.HandleFunc("/api/daemon/status", handleDaemonStatus)
	mux.HandleFunc("/api/daemon/stats", handleDaemonStats)

	// FRP管理
	mux.HandleFunc("/api/frp/status", handleFRPStatus)
	mux.HandleFunc("/api/frp/nodes", handleFRPNodes)
	mux.HandleFunc("/api/frp/tunnels", handleFRPTunnels)
	mux.HandleFunc("/api/frp/tunnels/", handleFRPTunnelAction)

	// 实例管理增强
	mux.HandleFunc("/api/templates", handleTemplates)
	mux.HandleFunc("/api/templates/", handleTemplateAction)
	mux.HandleFunc("/api/groups", handleGroups)
	mux.HandleFunc("/api/groups/", handleGroupAction)
	mux.HandleFunc("/api/batch", handleBatchOperations)
	mux.HandleFunc("/api/batch/", handleBatchAction)
	
	// WebSocket
	mux.HandleFunc("/ws", ServeWS)

	return mux
}

// serveIndex 服务主页
func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	indexPath := filepath.Join("web", "index.html")
	http.ServeFile(w, r, indexPath)
}

// handleStatus 处理状态查询
func handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"version": "1.0.0",
		"name":    "EasilyPanel5",
	}

	writeJSONResponse(w, response)
}



// writeJSONResponse 写入JSON响应
func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, data); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeErrorResponse 写入错误响应
func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    code,
	}
	
	writeJSON(w, response)
}
