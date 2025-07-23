package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// AuthHandlers 认证处理器
type AuthHandlers struct {
	service *AuthService
}

// NewAuthHandlers 创建认证处理器
func NewAuthHandlers(service *AuthService) *AuthHandlers {
	return &AuthHandlers{
		service: service,
	}
}

// RegisterRoutes 注册路由
func (h *AuthHandlers) RegisterRoutes(mux *http.ServeMux) {
	// 认证相关
	mux.HandleFunc("/api/auth/login", h.handleLogin)
	mux.HandleFunc("/api/auth/logout", h.handleLogout)
	mux.HandleFunc("/api/auth/status", h.handleAuthStatus)
	mux.HandleFunc("/api/auth/config", h.handleAuthConfig)
	
	// 2FA相关
	mux.HandleFunc("/api/auth/totp/setup", h.handleTOTPSetup)
	mux.HandleFunc("/api/auth/totp/confirm", h.handleTOTPConfirm)
	mux.HandleFunc("/api/auth/totp/disable", h.handleTOTPDisable)
	mux.HandleFunc("/api/auth/backup-codes/regenerate", h.handleRegenerateBackupCodes)
	
	// 用户管理
	mux.HandleFunc("/api/auth/password/change", h.handleChangePassword)
	mux.HandleFunc("/api/auth/profile", h.handleProfile)
	mux.HandleFunc("/api/auth/security-logs", h.handleSecurityLogs)

	// 调试接口（仅在开发环境启用）
	mux.HandleFunc("/api/auth/totp/debug", h.handleTOTPDebug)
}

// handleLogin 处理登录
func (h *AuthHandlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// 获取客户端信息
	ipAddress := h.getClientIP(r)
	userAgent := r.UserAgent()
	
	// 执行登录
	resp, err := h.service.Login(&req, ipAddress, userAgent)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	
	// 如果登录成功，设置cookie
	if resp.Success && resp.Token != "" {
		http.SetCookie(w, &http.Cookie{
			Name:     "auth_token",
			Value:    resp.Token,
			Path:     "/",
			HttpOnly: true,
			Secure:   r.TLS != nil,
			SameSite: http.SameSiteStrictMode,
		})
	}
	
	json.NewEncoder(w).Encode(resp)
}

// handleLogout 处理登出
func (h *AuthHandlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	token := h.getTokenFromRequest(r)
	if token == "" {
		http.Error(w, "No token provided", http.StatusUnauthorized)
		return
	}
	
	err := h.service.Logout(token)
	if err != nil {
		http.Error(w, "Failed to logout", http.StatusInternalServerError)
		return
	}
	
	// 清除cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Logged out successfully",
	})
}

// handleAuthStatus 处理认证状态查询
func (h *AuthHandlers) handleAuthStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"enabled": h.service.IsEnabled(),
		"config": map[string]interface{}{
			"require_two_factor":       h.service.config.RequireTwoFactor,
			"password_min_length":      h.service.config.PasswordMinLength,
			"password_require_special": h.service.config.PasswordRequireSpecial,
			"max_failed_attempts":      h.service.config.MaxFailedAttempts,
		},
	}
	
	// 如果用户已登录，返回用户信息
	token := h.getTokenFromRequest(r)
	if token != "" {
		user, err := h.service.ValidateSession(token)
		if err == nil {
			response["authenticated"] = true
			response["user"] = &UserInfo{
				ID:               user.ID,
				Username:         user.Username,
				Email:            user.Email,
				Role:             user.Role,
				TwoFactorEnabled: user.TwoFactorEnabled,
				CreatedAt:        user.CreatedAt,
				LastLoginAt:      user.LastLoginAt,
			}
		} else {
			response["authenticated"] = false
		}
	} else {
		response["authenticated"] = false
	}
	
	json.NewEncoder(w).Encode(response)
}

// handleAuthConfig 处理认证配置
func (h *AuthHandlers) handleAuthConfig(w http.ResponseWriter, r *http.Request) {
	// 需要管理员权限
	user, err := h.requireAuth(w, r, RoleAdmin)
	if err != nil {
		return
	}
	
	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(h.service.GetConfig())
		
	case http.MethodPut:
		var config AuthConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		
		if err := h.service.UpdateConfig(&config); err != nil {
			http.Error(w, "Failed to update config", http.StatusInternalServerError)
			return
		}
		
		h.service.logSecurityEvent(user.ID, "config_updated", h.getClientIP(r), r.UserAgent(), "Auth config updated")
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Configuration updated successfully",
		})
		
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTOTPSetup 处理TOTP设置
func (h *AuthHandlers) handleTOTPSetup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}
	
	resp, err := h.service.SetupTOTP(user.ID)
	if err != nil {
		http.Error(w, "Failed to setup TOTP", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleTOTPConfirm 处理TOTP确认
func (h *AuthHandlers) handleTOTPConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}
	
	var req TOTPSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	resp, err := h.service.ConfirmTOTP(user.ID, req.Secret, req.Code)
	if err != nil {
		http.Error(w, "Failed to confirm TOTP", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleTOTPDisable 处理TOTP禁用
func (h *AuthHandlers) handleTOTPDisable(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}
	
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.service.DisableTOTP(user.ID, req.Password); err != nil {
		if authErr, ok := err.(*AuthError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"code":    authErr.Code,
				"message": authErr.Message,
			})
			return
		}
		http.Error(w, "Failed to disable TOTP", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Two-factor authentication disabled successfully",
	})
}

// handleRegenerateBackupCodes 处理备用码重新生成
func (h *AuthHandlers) handleRegenerateBackupCodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}
	
	resp, err := h.service.RegenerateBackupCodes(user.ID)
	if err != nil {
		http.Error(w, "Failed to regenerate backup codes", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleChangePassword 处理密码修改
func (h *AuthHandlers) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}
	
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.service.ChangePassword(user.ID, &req); err != nil {
		if authErr, ok := err.(*AuthError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"code":    authErr.Code,
				"message": authErr.Message,
			})
			return
		}
		http.Error(w, "Failed to change password", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Password changed successfully",
	})
}

// handleProfile 处理用户资料
func (h *AuthHandlers) handleProfile(w http.ResponseWriter, r *http.Request) {
	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}

	switch r.Method {
	case http.MethodGet:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(&UserInfo{
			ID:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Role:             user.Role,
			TwoFactorEnabled: user.TwoFactorEnabled,
			CreatedAt:        user.CreatedAt,
			LastLoginAt:      user.LastLoginAt,
		})

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSecurityLogs 处理安全日志
func (h *AuthHandlers) handleSecurityLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}

	// 解析查询参数
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // 默认50条
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	// 管理员可以查看所有日志，普通用户只能查看自己的
	userID := user.ID
	if user.Role == RoleAdmin {
		if reqUserID := r.URL.Query().Get("user_id"); reqUserID != "" {
			userID = reqUserID
		} else {
			userID = "" // 查看所有用户的日志
		}
	}

	logs := h.service.GetSecurityLogs(userID, limit)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"logs":    logs,
	})
}

// 辅助方法

// requireAuth 要求认证
func (h *AuthHandlers) requireAuth(w http.ResponseWriter, r *http.Request, requiredRole string) (*User, error) {
	token := h.getTokenFromRequest(r)
	if token == "" {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return nil, fmt.Errorf("no token")
	}

	user, err := h.service.ValidateSession(token)
	if err != nil {
		if authErr, ok := err.(*AuthError); ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"code":    authErr.Code,
				"message": authErr.Message,
			})
		} else {
			http.Error(w, "Authentication failed", http.StatusUnauthorized)
		}
		return nil, err
	}

	// 检查角色权限
	if requiredRole != "" && user.Role != requiredRole {
		http.Error(w, "Insufficient permissions", http.StatusForbidden)
		return nil, fmt.Errorf("insufficient permissions")
	}

	return user, nil
}

// getTokenFromRequest 从请求中获取令牌
func (h *AuthHandlers) getTokenFromRequest(r *http.Request) string {
	// 首先尝试从Authorization头获取
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1]
		}
	}

	// 然后尝试从Cookie获取
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// getClientIP 获取客户端IP地址
func (h *AuthHandlers) getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// AuthMiddleware 认证中间件
func (h *AuthHandlers) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 如果认证未启用，直接通过
		if !h.service.IsEnabled() {
			next.ServeHTTP(w, r)
			return
		}

		// 检查是否是跳过认证的路径
		skipPaths := []string{
			"/api/auth/login",
			"/api/auth/status",
			"/static/",
			"/favicon.ico",
		}

		for _, path := range skipPaths {
			if strings.HasPrefix(r.URL.Path, path) {
				next.ServeHTTP(w, r)
				return
			}
		}

		// 验证认证
		token := h.getTokenFromRequest(r)
		if token == "" {
			// 对于API请求返回JSON错误
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"code":    "authentication_required",
					"message": "Authentication required",
				})
				return
			}

			// 对于页面请求重定向到登录页
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		user, err := h.service.ValidateSession(token)
		if err != nil {
			// 清除无效的cookie
			http.SetCookie(w, &http.Cookie{
				Name:     "auth_token",
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})

			// 对于API请求返回JSON错误
			if strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				if authErr, ok := err.(*AuthError); ok {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"code":    authErr.Code,
						"message": authErr.Message,
					})
				} else {
					json.NewEncoder(w).Encode(map[string]interface{}{
						"success": false,
						"code":    "authentication_failed",
						"message": "Authentication failed",
					})
				}
				return
			}

			// 对于页面请求重定向到登录页
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		// 将用户信息添加到请求上下文中
		r.Header.Set("X-User-ID", user.ID)
		r.Header.Set("X-User-Role", user.Role)

		next.ServeHTTP(w, r)
	})
}

// handleTOTPDebug 处理TOTP调试（仅用于开发调试）
func (h *AuthHandlers) handleTOTPDebug(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	user, err := h.requireAuth(w, r, "")
	if err != nil {
		return
	}

	if !user.TwoFactorEnabled {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Two-factor authentication is not enabled",
		})
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 使用调试方法验证TOTP
	valid, debugInfo := h.service.totpManager.ValidateCodeDebug(user.TwoFactorSecret, req.Code)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"valid":      valid,
		"debug_info": debugInfo,
	})
}
