package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"easilypanel5/utils"
)

// StandardResponse 标准响应结构
type StandardResponse struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Code      string      `json:"code,omitempty"`
	Message   string      `json:"message,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// CORSMiddleware CORS跨域中间件
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置CORS头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware 请求日志中间件
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 创建响应记录器
		recorder := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 执行请求
		next.ServeHTTP(recorder, r)

		// 记录请求日志
		duration := time.Since(start)

		// 使用新的日志系统
		logger := utils.GetLogger()

		// 根据状态码选择日志级别
		if recorder.statusCode >= 500 {
			logger.Error("[%s] %s %s - %d - %v", r.Method, r.RequestURI, r.RemoteAddr, recorder.statusCode, duration)
		} else if recorder.statusCode >= 400 {
			logger.Warn("[%s] %s %s - %d - %v", r.Method, r.RequestURI, r.RemoteAddr, recorder.statusCode, duration)
		} else {
			logger.Info("[%s] %s %s - %d - %v", r.Method, r.RequestURI, r.RemoteAddr, recorder.statusCode, duration)
		}
	})
}

// RecoveryMiddleware Panic恢复中间件
func RecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 记录panic信息
				log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

				// 返回500错误
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				errorResp := ErrorResponse{
					Success:   false,
					Error:     "internal_server_error",
					Code:      "INTERNAL_ERROR",
					Message:   "Internal server error occurred",
					Timestamp: time.Now(),
				}

				json.NewEncoder(w).Encode(errorResp)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SecurityMiddleware 安全头中间件
func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置安全头
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// ContentTypeMiddleware 内容类型中间件
func ContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 为API请求设置JSON内容类型
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
		}

		next.ServeHTTP(w, r)
	})
}

// responseRecorder 响应记录器，用于记录状态码
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *responseRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// ChainMiddleware 中间件链构建器
func ChainMiddleware(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}

// WriteStandardResponse 写入标准响应
func WriteStandardResponse(w http.ResponseWriter, data interface{}) {
	response := StandardResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// WriteStandardError 写入标准错误响应
func WriteStandardError(w http.ResponseWriter, code string, message string, httpStatus int) {
	response := ErrorResponse{
		Success:   false,
		Error:     strings.ToLower(strings.ReplaceAll(code, " ", "_")),
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// ValidateContentType 验证请求内容类型
func ValidateContentType(r *http.Request, expectedType string) error {
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		return fmt.Errorf("missing content-type header")
	}

	if !strings.Contains(contentType, expectedType) {
		return fmt.Errorf("invalid content-type: expected %s, got %s", expectedType, contentType)
	}

	return nil
}

// ValidateMethod 验证HTTP方法
func ValidateMethod(r *http.Request, allowedMethods ...string) error {
	for _, method := range allowedMethods {
		if r.Method == method {
			return nil
		}
	}
	return fmt.Errorf("method %s not allowed", r.Method)
}

// GetClientIP 获取客户端IP地址
func GetClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}
