package api

import (
	"net/http"
	"path/filepath"
	"strings"
)

// AuthHandlers 认证处理器接口
type AuthHandlers interface {
	RegisterRoutes(mux *http.ServeMux)
	AuthMiddleware(next http.Handler) http.Handler
}

// SetupRoutes 设置路由
func SetupRoutes(authHandlers AuthHandlers) http.Handler {
	mux := http.NewServeMux()

	// 注册认证路由
	if authHandlers != nil {
		authHandlers.RegisterRoutes(mux)
	}

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

	// 插件管理相关
	mux.HandleFunc("/api/plugins/list", handlePluginsList)
	mux.HandleFunc("/api/plugins/search", handlePluginsSearch)
	mux.HandleFunc("/api/plugins/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/versions") {
			handlePluginVersions(w, r)
		} else {
			handlePluginInfo(w, r)
		}
	})
	mux.HandleFunc("/api/plugins/download", handlePluginDownload)

	// 服务器管理相关
	mux.HandleFunc("/api/servers", handleServers)

	// 服务器详细操作（使用更具体的路由模式）
	mux.HandleFunc("/api/servers/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasSuffix(path, "/logs") {
			handleServerLogs(w, r)
		} else if strings.HasSuffix(path, "/command") {
			handleServerCommand(w, r)
		} else {
			handleServerDetail(w, r)
		}
	})

	// 守护管理
	mux.HandleFunc("/api/daemon/status", handleDaemonStatus)
	mux.HandleFunc("/api/daemon/stats", handleDaemonStats)

	// FRP管理
	mux.HandleFunc("/api/frp/status", handleFRPStatus)
	mux.HandleFunc("/api/frp/nodes", handleFRPNodes)
	mux.HandleFunc("/api/frp/tunnels", handleFRPTunnels)
	mux.HandleFunc("/api/frp/tunnels/", handleFRPTunnelAction)

	// 实例管理增强
	mux.HandleFunc("/api/groups", handleGroups)
	mux.HandleFunc("/api/groups/", handleGroupAction)
	mux.HandleFunc("/api/batch", handleBatchOperations)
	mux.HandleFunc("/api/batch/", handleBatchAction)

	// 文件管理
	mux.HandleFunc("/api/files", handleFiles)
	mux.HandleFunc("/api/files/download", handleFileDownload)
	mux.HandleFunc("/api/files/content", handleFileContent)

	// 系统监控
	mux.HandleFunc("/api/monitoring/system", handleSystemStats)
	mux.HandleFunc("/api/monitoring/servers", handleServerStats)
	mux.HandleFunc("/api/monitoring/historical", handleHistoricalStats)

	// 系统设置
	mux.HandleFunc("/api/settings", handleSettings)

	// 系统日志
	mux.HandleFunc("/api/logs/system", handleSystemLogs)

	// WebSocket
	mux.HandleFunc("/ws", ServeWS)

	// 构建中间件链
	var handler http.Handler = mux

	// 应用认证中间件（如果存在）
	if authHandlers != nil {
		handler = authHandlers.AuthMiddleware(handler)
	}

	// 应用通用中间件链
	handler = ChainMiddleware(handler,
		RecoveryMiddleware,    // 最外层：panic恢复
		LoggingMiddleware,     // 日志记录
		SecurityMiddleware,    // 安全头
		CORSMiddleware,        // CORS处理
		ContentTypeMiddleware, // 内容类型设置
	)

	return handler
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
	if err := ValidateMethod(r, http.MethodGet); err != nil {
		WriteStandardError(w, "METHOD_NOT_ALLOWED", err.Error(), http.StatusMethodNotAllowed)
		return
	}

	response := map[string]interface{}{
		"status":  "ok",
		"version": "1.1.0",
		"name":    "EasilyPanel5",
	}

	WriteStandardResponse(w, response)
}



// writeJSONResponse 写入JSON响应（保持向后兼容）
func writeJSONResponse(w http.ResponseWriter, data interface{}) {
	WriteStandardResponse(w, data)
}

// writeErrorResponse 写入错误响应（保持向后兼容）
func writeErrorResponse(w http.ResponseWriter, message string, code int) {
	WriteStandardError(w, "API_ERROR", message, code)
}
