package auth

import (
	"time"
)

// User 用户信息
type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	PasswordHash string   `json:"password_hash"`
	Salt        string    `json:"salt"`
	Role        string    `json:"role"`        // admin, user
	Status      string    `json:"status"`      // active, disabled, locked
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at"`
	
	// 2FA相关
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	TwoFactorSecret  string    `json:"two_factor_secret"`
	BackupCodes      []string  `json:"backup_codes"`
	TwoFactorSetupAt *time.Time `json:"two_factor_setup_at"`
	
	// 安全相关
	FailedLoginAttempts int       `json:"failed_login_attempts"`
	LastFailedLoginAt   *time.Time `json:"last_failed_login_at"`
	LockedUntil         *time.Time `json:"locked_until"`
}

// Session 会话信息
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	IsActive  bool      `json:"is_active"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	TOTPCode   string `json:"totp_code,omitempty"`
	BackupCode string `json:"backup_code,omitempty"`
	RememberMe bool   `json:"remember_me"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success      bool   `json:"success"`
	Token        string `json:"token,omitempty"`
	ExpiresAt    int64  `json:"expires_at,omitempty"`
	RequiresTOTP bool   `json:"requires_totp"`
	Message      string `json:"message"`
	User         *UserInfo `json:"user,omitempty"`
}

// UserInfo 用户信息（不包含敏感数据）
type UserInfo struct {
	ID               string    `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Role             string    `json:"role"`
	TwoFactorEnabled bool      `json:"two_factor_enabled"`
	CreatedAt        time.Time `json:"created_at"`
	LastLoginAt      *time.Time `json:"last_login_at"`
}

// TOTPSetupRequest TOTP设置请求
type TOTPSetupRequest struct {
	Secret string `json:"secret"`
	Code   string `json:"code"`
}

// TOTPSetupResponse TOTP设置响应
type TOTPSetupResponse struct {
	Success     bool     `json:"success"`
	Secret      string   `json:"secret,omitempty"`
	QRCodeURL   string   `json:"qr_code_url,omitempty"`
	BackupCodes []string `json:"backup_codes,omitempty"`
	Message     string   `json:"message"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	TOTPCode        string `json:"totp_code,omitempty"`
}

// BackupCodesResponse 备用码响应
type BackupCodesResponse struct {
	Success     bool     `json:"success"`
	BackupCodes []string `json:"backup_codes,omitempty"`
	Message     string   `json:"message"`
}

// SecurityLog 安全日志
type SecurityLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Action    string    `json:"action"`    // login, logout, login_failed, password_changed, 2fa_enabled, etc.
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Details   string    `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled                bool          `json:"enabled"`                  // 启用认证
	RequireTwoFactor       bool          `json:"require_two_factor"`       // 强制2FA
	SessionTimeout         time.Duration `json:"session_timeout"`          // 会话超时
	MaxFailedAttempts      int           `json:"max_failed_attempts"`      // 最大失败尝试次数
	LockoutDuration        time.Duration `json:"lockout_duration"`         // 锁定时长
	PasswordMinLength      int           `json:"password_min_length"`      // 密码最小长度
	PasswordRequireSpecial bool          `json:"password_require_special"` // 密码需要特殊字符
	BackupCodesCount       int           `json:"backup_codes_count"`       // 备用码数量
	TOTPIssuer             string        `json:"totp_issuer"`              // TOTP发行者名称
	JWTSecret              string        `json:"jwt_secret"`               // JWT密钥
}

// 用户角色常量
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// 用户状态常量
const (
	StatusActive   = "active"
	StatusDisabled = "disabled"
	StatusLocked   = "locked"
)

// 安全操作常量
const (
	ActionLogin          = "login"
	ActionLogout         = "logout"
	ActionLoginFailed    = "login_failed"
	ActionPasswordChange = "password_changed"
	Action2FAEnabled     = "2fa_enabled"
	Action2FADisabled    = "2fa_disabled"
	ActionBackupUsed     = "backup_code_used"
	ActionAccountLocked  = "account_locked"
	ActionAccountUnlocked = "account_unlocked"
)

// AuthError 认证错误
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AuthError) Error() string {
	return e.Message
}

// 错误代码常量
const (
	ErrInvalidCredentials = "invalid_credentials"
	ErrAccountLocked      = "account_locked"
	ErrAccountDisabled    = "account_disabled"
	ErrTOTPRequired       = "totp_required"
	ErrInvalidTOTP        = "invalid_totp"
	ErrInvalidBackupCode  = "invalid_backup_code"
	ErrSessionExpired     = "session_expired"
	ErrInvalidToken       = "invalid_token"
	ErrPasswordTooWeak    = "password_too_weak"
	ErrUserNotFound       = "user_not_found"
	ErrUserExists         = "user_exists"
	Err2FAAlreadyEnabled  = "2fa_already_enabled"
	Err2FANotEnabled      = "2fa_not_enabled"
)

// TokenClaims JWT令牌声明
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	SessionID string `json:"session_id"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
}

// AuthMiddleware 认证中间件配置
type AuthMiddleware struct {
	SkipPaths []string `json:"skip_paths"` // 跳过认证的路径
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Enabled:                false, // 默认禁用，需要手动启用
		RequireTwoFactor:       false,
		SessionTimeout:         24 * time.Hour,
		MaxFailedAttempts:      5,
		LockoutDuration:        15 * time.Minute,
		PasswordMinLength:      8,
		PasswordRequireSpecial: true,
		BackupCodesCount:       10,
		TOTPIssuer:             "EasilyPanel5",
		JWTSecret:              "", // 需要生成
	}
}
