package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AuthService 认证服务
type AuthService struct {
	config           *AuthConfig
	users            map[string]*User    // username -> User
	sessions         map[string]*Session // token -> Session
	securityLogs     []*SecurityLog
	totpManager      *TOTPManager
	backupManager    *BackupCodeManager
	dataDir          string
}

// NewAuthService 创建认证服务
func NewAuthService(dataDir string) *AuthService {
	service := &AuthService{
		config:        DefaultAuthConfig(),
		users:         make(map[string]*User),
		sessions:      make(map[string]*Session),
		securityLogs:  make([]*SecurityLog, 0),
		totpManager:   NewTOTPManager("EasilyPanel5"),
		backupManager: NewBackupCodeManager(),
		dataDir:       dataDir,
	}
	
	// 确保数据目录存在
	os.MkdirAll(dataDir, 0755)
	
	// 加载配置和数据
	service.loadConfig()
	service.loadUsers()
	service.loadSessions()
	
	// 如果没有用户，创建默认管理员
	if len(service.users) == 0 {
		service.createDefaultAdmin()
	}
	
	// 生成JWT密钥（如果不存在）
	if service.config.JWTSecret == "" {
		service.generateJWTSecret()
	}
	
	return service
}

// loadConfig 加载配置
func (s *AuthService) loadConfig() {
	configPath := filepath.Join(s.dataDir, "auth_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		// 使用默认配置
		s.saveConfig()
		return
	}
	
	if err := json.Unmarshal(data, s.config); err != nil {
		fmt.Printf("Failed to load auth config: %v\n", err)
		s.saveConfig()
	}
}

// saveConfig 保存配置
func (s *AuthService) saveConfig() {
	configPath := filepath.Join(s.dataDir, "auth_config.json")
	data, err := json.MarshalIndent(s.config, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal auth config: %v\n", err)
		return
	}
	
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		fmt.Printf("Failed to save auth config: %v\n", err)
	}
}

// loadUsers 加载用户数据
func (s *AuthService) loadUsers() {
	usersPath := filepath.Join(s.dataDir, "users.json")
	data, err := os.ReadFile(usersPath)
	if err != nil {
		return
	}
	
	var users []*User
	if err := json.Unmarshal(data, &users); err != nil {
		fmt.Printf("Failed to load users: %v\n", err)
		return
	}
	
	for _, user := range users {
		s.users[user.Username] = user
	}
}

// saveUsers 保存用户数据
func (s *AuthService) saveUsers() {
	usersPath := filepath.Join(s.dataDir, "users.json")
	
	users := make([]*User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal users: %v\n", err)
		return
	}
	
	if err := os.WriteFile(usersPath, data, 0600); err != nil {
		fmt.Printf("Failed to save users: %v\n", err)
	}
}

// loadSessions 加载会话数据
func (s *AuthService) loadSessions() {
	sessionsPath := filepath.Join(s.dataDir, "sessions.json")
	data, err := os.ReadFile(sessionsPath)
	if err != nil {
		return
	}
	
	var sessions []*Session
	if err := json.Unmarshal(data, &sessions); err != nil {
		fmt.Printf("Failed to load sessions: %v\n", err)
		return
	}
	
	// 只加载未过期的会话
	now := time.Now()
	for _, session := range sessions {
		if session.ExpiresAt.After(now) && session.IsActive {
			s.sessions[session.Token] = session
		}
	}
}

// saveSessions 保存会话数据
func (s *AuthService) saveSessions() {
	sessionsPath := filepath.Join(s.dataDir, "sessions.json")
	
	sessions := make([]*Session, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	
	data, err := json.MarshalIndent(sessions, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal sessions: %v\n", err)
		return
	}
	
	if err := os.WriteFile(sessionsPath, data, 0600); err != nil {
		fmt.Printf("Failed to save sessions: %v\n", err)
	}
}

// createDefaultAdmin 创建默认管理员账户
func (s *AuthService) createDefaultAdmin() {
	adminID := s.generateID()
	salt := s.generateSalt()
	passwordHash := s.hashPassword("admin123", salt)
	
	admin := &User{
		ID:           adminID,
		Username:     "admin",
		Email:        "admin@easilypanel.local",
		PasswordHash: passwordHash,
		Salt:         salt,
		Role:         RoleAdmin,
		Status:       StatusActive,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		TwoFactorEnabled: false,
		FailedLoginAttempts: 0,
	}
	
	s.users[admin.Username] = admin
	s.saveUsers()
	
	fmt.Println("Created default admin user: admin/admin123")
	fmt.Println("Please change the default password after first login!")
}

// generateJWTSecret 生成JWT密钥
func (s *AuthService) generateJWTSecret() {
	secret := make([]byte, 32)
	rand.Read(secret)
	s.config.JWTSecret = hex.EncodeToString(secret)
	s.saveConfig()
}

// generateID 生成唯一ID
func (s *AuthService) generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// generateSalt 生成盐值
func (s *AuthService) generateSalt() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// hashPassword 哈希密码
func (s *AuthService) hashPassword(password, salt string) string {
	hash := sha256.Sum256([]byte(password + salt))
	return hex.EncodeToString(hash[:])
}

// verifyPassword 验证密码
func (s *AuthService) verifyPassword(password, hash, salt string) bool {
	return s.hashPassword(password, salt) == hash
}

// generateToken 生成简单令牌
func (s *AuthService) generateToken(user *User, sessionID string) (string, error) {
	// 生成简单的随机令牌
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(tokenBytes), nil
}

// validateToken 验证令牌（通过会话）
func (s *AuthService) validateToken(tokenString string) (*User, error) {
	return s.ValidateSession(tokenString)
}

// logSecurityEvent 记录安全事件
func (s *AuthService) logSecurityEvent(userID, action, ipAddress, userAgent, details string) {
	log := &SecurityLog{
		ID:        s.generateID(),
		UserID:    userID,
		Action:    action,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		CreatedAt: time.Now(),
	}
	
	s.securityLogs = append(s.securityLogs, log)
	
	// 保持最近1000条日志
	if len(s.securityLogs) > 1000 {
		s.securityLogs = s.securityLogs[len(s.securityLogs)-1000:]
	}
}

// GetConfig 获取认证配置
func (s *AuthService) GetConfig() *AuthConfig {
	return s.config
}

// UpdateConfig 更新认证配置
func (s *AuthService) UpdateConfig(config *AuthConfig) error {
	s.config = config
	s.saveConfig()
	return nil
}

// IsEnabled 检查认证是否启用
func (s *AuthService) IsEnabled() bool {
	return s.config.Enabled
}

// EnableAuth 启用认证
func (s *AuthService) EnableAuth() error {
	s.config.Enabled = true
	s.saveConfig()
	return nil
}

// DisableAuth 禁用认证
func (s *AuthService) DisableAuth() error {
	s.config.Enabled = false
	s.saveConfig()
	return nil
}

// Login 用户登录
func (s *AuthService) Login(req *LoginRequest, ipAddress, userAgent string) (*LoginResponse, error) {
	// 查找用户
	user, exists := s.users[req.Username]
	if !exists {
		s.logSecurityEvent("", ActionLoginFailed, ipAddress, userAgent,
			fmt.Sprintf("User not found: %s", req.Username))
		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// 检查账户状态
	if user.Status == StatusDisabled {
		s.logSecurityEvent(user.ID, ActionLoginFailed, ipAddress, userAgent, "Account disabled")
		return &LoginResponse{
			Success: false,
			Message: "Account is disabled",
		}, nil
	}

	// 检查账户锁定
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		s.logSecurityEvent(user.ID, ActionLoginFailed, ipAddress, userAgent, "Account locked")
		return &LoginResponse{
			Success: false,
			Message: "Account is temporarily locked",
		}, nil
	}

	// 验证密码
	if !s.verifyPassword(req.Password, user.PasswordHash, user.Salt) {
		user.FailedLoginAttempts++
		user.LastFailedLoginAt = &time.Time{}
		*user.LastFailedLoginAt = time.Now()

		// 检查是否需要锁定账户
		if user.FailedLoginAttempts >= s.config.MaxFailedAttempts {
			lockUntil := time.Now().Add(s.config.LockoutDuration)
			user.LockedUntil = &lockUntil
			s.logSecurityEvent(user.ID, ActionAccountLocked, ipAddress, userAgent,
				fmt.Sprintf("Account locked after %d failed attempts", user.FailedLoginAttempts))
		}

		s.saveUsers()
		s.logSecurityEvent(user.ID, ActionLoginFailed, ipAddress, userAgent, "Invalid password")

		return &LoginResponse{
			Success: false,
			Message: "Invalid username or password",
		}, nil
	}

	// 重置失败计数
	user.FailedLoginAttempts = 0
	user.LastFailedLoginAt = nil
	user.LockedUntil = nil

	// 检查是否需要2FA
	if user.TwoFactorEnabled {
		if req.TOTPCode == "" && req.BackupCode == "" {
			return &LoginResponse{
				Success:      false,
				RequiresTOTP: true,
				Message:      "Two-factor authentication required",
			}, nil
		}

		// 验证TOTP或备用码
		if req.TOTPCode != "" {
			if !s.totpManager.ValidateCode(user.TwoFactorSecret, req.TOTPCode) {
				s.logSecurityEvent(user.ID, ActionLoginFailed, ipAddress, userAgent, "Invalid TOTP code")
				return &LoginResponse{
					Success: false,
					Message: "Invalid verification code",
				}, nil
			}
		} else if req.BackupCode != "" {
			valid, newCodes := s.backupManager.ValidateBackupCode(req.BackupCode, user.BackupCodes)
			if !valid {
				s.logSecurityEvent(user.ID, ActionLoginFailed, ipAddress, userAgent, "Invalid backup code")
				return &LoginResponse{
					Success: false,
					Message: "Invalid backup code",
				}, nil
			}

			// 更新备用码列表
			user.BackupCodes = newCodes
			s.logSecurityEvent(user.ID, ActionBackupUsed, ipAddress, userAgent, "Backup code used for login")
		}
	}

	// 创建会话
	sessionID := s.generateID()
	token, err := s.generateToken(user, sessionID)
	if err != nil {
		return &LoginResponse{
			Success: false,
			Message: "Failed to create session",
		}, nil
	}

	expiresAt := time.Now().Add(s.config.SessionTimeout)
	session := &Session{
		ID:        sessionID,
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	s.sessions[token] = session

	// 更新用户最后登录时间
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	s.saveUsers()
	s.saveSessions()

	s.logSecurityEvent(user.ID, ActionLogin, ipAddress, userAgent, "Successful login")

	return &LoginResponse{
		Success:   true,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
		Message:   "Login successful",
		User: &UserInfo{
			ID:               user.ID,
			Username:         user.Username,
			Email:            user.Email,
			Role:             user.Role,
			TwoFactorEnabled: user.TwoFactorEnabled,
			CreatedAt:        user.CreatedAt,
			LastLoginAt:      user.LastLoginAt,
		},
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) error {
	session, exists := s.sessions[token]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.IsActive = false
	delete(s.sessions, token)
	s.saveSessions()

	s.logSecurityEvent(session.UserID, ActionLogout, session.IPAddress, session.UserAgent, "User logout")

	return nil
}

// ValidateSession 验证会话
func (s *AuthService) ValidateSession(token string) (*User, error) {
	session, exists := s.sessions[token]
	if !exists {
		return nil, &AuthError{Code: ErrInvalidToken, Message: "Invalid session token"}
	}

	if !session.IsActive {
		return nil, &AuthError{Code: ErrSessionExpired, Message: "Session is not active"}
	}

	if time.Now().After(session.ExpiresAt) {
		session.IsActive = false
		delete(s.sessions, token)
		s.saveSessions()
		return nil, &AuthError{Code: ErrSessionExpired, Message: "Session has expired"}
	}

	// 查找用户
	for _, user := range s.users {
		if user.ID == session.UserID {
			return user, nil
		}
	}

	return nil, &AuthError{Code: ErrUserNotFound, Message: "User not found"}
}

// SetupTOTP 设置TOTP
func (s *AuthService) SetupTOTP(userID string) (*TOTPSetupResponse, error) {
	user := s.getUserByID(userID)
	if user == nil {
		return &TOTPSetupResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	if user.TwoFactorEnabled {
		return &TOTPSetupResponse{
			Success: false,
			Message: "Two-factor authentication is already enabled",
		}, nil
	}

	// 生成新的密钥
	secret, err := s.totpManager.GenerateSecret()
	if err != nil {
		return &TOTPSetupResponse{
			Success: false,
			Message: "Failed to generate secret",
		}, nil
	}

	// 生成QR码URL
	qrURL := s.totpManager.GenerateQRCodeURL(secret, user.Username)

	return &TOTPSetupResponse{
		Success:   true,
		Secret:    secret,
		QRCodeURL: qrURL,
		Message:   "TOTP setup initiated",
	}, nil
}

// ConfirmTOTP 确认TOTP设置
func (s *AuthService) ConfirmTOTP(userID, secret, code string) (*TOTPSetupResponse, error) {
	user := s.getUserByID(userID)
	if user == nil {
		return &TOTPSetupResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	if user.TwoFactorEnabled {
		return &TOTPSetupResponse{
			Success: false,
			Message: "Two-factor authentication is already enabled",
		}, nil
	}

	// 验证TOTP码
	if !s.totpManager.ValidateCode(secret, code) {
		return &TOTPSetupResponse{
			Success: false,
			Message: "Invalid verification code",
		}, nil
	}

	// 生成备用码
	backupCodes, err := s.backupManager.GenerateBackupCodes(s.config.BackupCodesCount)
	if err != nil {
		return &TOTPSetupResponse{
			Success: false,
			Message: "Failed to generate backup codes",
		}, nil
	}

	// 启用2FA
	user.TwoFactorEnabled = true
	user.TwoFactorSecret = secret
	user.BackupCodes = backupCodes
	now := time.Now()
	user.TwoFactorSetupAt = &now
	user.UpdatedAt = now

	s.saveUsers()
	s.logSecurityEvent(user.ID, Action2FAEnabled, "", "", "Two-factor authentication enabled")

	return &TOTPSetupResponse{
		Success:     true,
		BackupCodes: backupCodes,
		Message:     "Two-factor authentication enabled successfully",
	}, nil
}

// DisableTOTP 禁用TOTP
func (s *AuthService) DisableTOTP(userID, password string) error {
	user := s.getUserByID(userID)
	if user == nil {
		return &AuthError{Code: ErrUserNotFound, Message: "User not found"}
	}

	if !user.TwoFactorEnabled {
		return &AuthError{Code: Err2FANotEnabled, Message: "Two-factor authentication is not enabled"}
	}

	// 验证密码
	if !s.verifyPassword(password, user.PasswordHash, user.Salt) {
		return &AuthError{Code: ErrInvalidCredentials, Message: "Invalid password"}
	}

	// 禁用2FA
	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""
	user.BackupCodes = nil
	user.TwoFactorSetupAt = nil
	user.UpdatedAt = time.Now()

	s.saveUsers()
	s.logSecurityEvent(user.ID, Action2FADisabled, "", "", "Two-factor authentication disabled")

	return nil
}

// RegenerateBackupCodes 重新生成备用码
func (s *AuthService) RegenerateBackupCodes(userID string) (*BackupCodesResponse, error) {
	user := s.getUserByID(userID)
	if user == nil {
		return &BackupCodesResponse{
			Success: false,
			Message: "User not found",
		}, nil
	}

	if !user.TwoFactorEnabled {
		return &BackupCodesResponse{
			Success: false,
			Message: "Two-factor authentication is not enabled",
		}, nil
	}

	// 生成新的备用码
	backupCodes, err := s.backupManager.GenerateBackupCodes(s.config.BackupCodesCount)
	if err != nil {
		return &BackupCodesResponse{
			Success: false,
			Message: "Failed to generate backup codes",
		}, nil
	}

	user.BackupCodes = backupCodes
	user.UpdatedAt = time.Now()

	s.saveUsers()
	s.logSecurityEvent(user.ID, "backup_codes_regenerated", "", "", "Backup codes regenerated")

	return &BackupCodesResponse{
		Success:     true,
		BackupCodes: backupCodes,
		Message:     "Backup codes regenerated successfully",
	}, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(userID string, req *ChangePasswordRequest) error {
	user := s.getUserByID(userID)
	if user == nil {
		return &AuthError{Code: ErrUserNotFound, Message: "User not found"}
	}

	// 验证当前密码
	if !s.verifyPassword(req.CurrentPassword, user.PasswordHash, user.Salt) {
		return &AuthError{Code: ErrInvalidCredentials, Message: "Current password is incorrect"}
	}

	// 如果启用了2FA，验证TOTP码
	if user.TwoFactorEnabled && req.TOTPCode != "" {
		if !s.totpManager.ValidateCode(user.TwoFactorSecret, req.TOTPCode) {
			return &AuthError{Code: ErrInvalidTOTP, Message: "Invalid verification code"}
		}
	}

	// 验证新密码强度
	if err := s.validatePasswordStrength(req.NewPassword); err != nil {
		return err
	}

	// 更新密码
	newSalt := s.generateSalt()
	newHash := s.hashPassword(req.NewPassword, newSalt)

	user.PasswordHash = newHash
	user.Salt = newSalt
	user.UpdatedAt = time.Now()

	s.saveUsers()
	s.logSecurityEvent(user.ID, ActionPasswordChange, "", "", "Password changed")

	return nil
}

// validatePasswordStrength 验证密码强度
func (s *AuthService) validatePasswordStrength(password string) error {
	if len(password) < s.config.PasswordMinLength {
		return &AuthError{
			Code:    ErrPasswordTooWeak,
			Message: fmt.Sprintf("Password must be at least %d characters long", s.config.PasswordMinLength),
		}
	}

	if s.config.PasswordRequireSpecial {
		hasSpecial := false
		specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
		for _, char := range password {
			if strings.ContainsRune(specialChars, char) {
				hasSpecial = true
				break
			}
		}

		if !hasSpecial {
			return &AuthError{
				Code:    ErrPasswordTooWeak,
				Message: "Password must contain at least one special character",
			}
		}
	}

	return nil
}

// getUserByID 根据ID获取用户
func (s *AuthService) getUserByID(userID string) *User {
	for _, user := range s.users {
		if user.ID == userID {
			return user
		}
	}
	return nil
}

// GetSecurityLogs 获取安全日志
func (s *AuthService) GetSecurityLogs(userID string, limit int) []*SecurityLog {
	logs := make([]*SecurityLog, 0)

	for i := len(s.securityLogs) - 1; i >= 0 && len(logs) < limit; i-- {
		log := s.securityLogs[i]
		if userID == "" || log.UserID == userID {
			logs = append(logs, log)
		}
	}

	return logs
}

// CleanupExpiredSessions 清理过期会话
func (s *AuthService) CleanupExpiredSessions() {
	now := time.Now()
	for token, session := range s.sessions {
		if now.After(session.ExpiresAt) || !session.IsActive {
			delete(s.sessions, token)
		}
	}
	s.saveSessions()
}
