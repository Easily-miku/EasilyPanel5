package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

// TOTPManager TOTP管理器
type TOTPManager struct {
	issuer string
	period int // 时间步长（秒）
	digits int // 验证码位数
}

// NewTOTPManager 创建TOTP管理器
func NewTOTPManager(issuer string) *TOTPManager {
	return &TOTPManager{
		issuer: issuer,
		period: 30, // 30秒
		digits: 6,  // 6位数字
	}
}

// GenerateSecret 生成TOTP密钥
func (t *TOTPManager) GenerateSecret() (string, error) {
	// 生成20字节的随机密钥
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", fmt.Errorf("failed to generate secret: %v", err)
	}
	
	// 使用Base32编码
	return base32.StdEncoding.EncodeToString(secret), nil
}

// GenerateCode 生成TOTP验证码
func (t *TOTPManager) GenerateCode(secret string, timestamp int64) (string, error) {
	// 解码密钥
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", fmt.Errorf("invalid secret: %v", err)
	}
	
	// 计算时间步数
	timeStep := timestamp / int64(t.period)
	
	// 生成HOTP
	return t.generateHOTP(key, timeStep), nil
}

// ValidateCode 验证TOTP验证码
func (t *TOTPManager) ValidateCode(secret, code string) bool {
	return t.ValidateCodeWithWindow(secret, code, 2)
}

// ValidateCodeWithWindow 验证TOTP验证码（带时间窗口）
func (t *TOTPManager) ValidateCodeWithWindow(secret, code string, window int) bool {
	now := time.Now().Unix()
	currentTimeStep := now / int64(t.period)

	// 检查当前时间步和前后窗口内的验证码
	for i := -window; i <= window; i++ {
		timeStep := currentTimeStep + int64(i)
		timestamp := timeStep * int64(t.period)
		expectedCode, err := t.GenerateCode(secret, timestamp)
		if err != nil {
			return false
		}

		if code == expectedCode {
			return true
		}
	}

	return false
}

// ValidateCodeWithReplayProtection 验证TOTP验证码（带防重放攻击）
func (t *TOTPManager) ValidateCodeWithReplayProtection(secret, code string, lastUsedTimeStep int64, window int) (bool, int64) {
	now := time.Now().Unix()
	currentTimeStep := now / int64(t.period)

	// 检查当前时间步和前后窗口内的验证码
	for i := -window; i <= window; i++ {
		timeStep := currentTimeStep + int64(i)

		// 防重放攻击：不允许使用已经使用过的时间步
		if timeStep <= lastUsedTimeStep {
			continue
		}

		timestamp := timeStep * int64(t.period)
		expectedCode, err := t.GenerateCode(secret, timestamp)
		if err != nil {
			continue
		}

		if code == expectedCode {
			return true, timeStep
		}
	}

	return false, lastUsedTimeStep
}

// GenerateQRCodeURL 生成QR码URL
func (t *TOTPManager) GenerateQRCodeURL(secret, accountName string) string {
	// 构建otpauth URL
	params := url.Values{}
	params.Set("secret", secret)
	params.Set("issuer", t.issuer)
	params.Set("algorithm", "SHA1")
	params.Set("digits", fmt.Sprintf("%d", t.digits))
	params.Set("period", fmt.Sprintf("%d", t.period))
	
	// 格式: otpauth://totp/Issuer:AccountName?secret=...&issuer=...
	otpauthURL := fmt.Sprintf("otpauth://totp/%s:%s?%s",
		url.QueryEscape(t.issuer),
		url.QueryEscape(accountName),
		params.Encode())
	
	return otpauthURL
}

// generateHOTP 生成HOTP验证码
func (t *TOTPManager) generateHOTP(key []byte, counter int64) string {
	// 将计数器转换为8字节大端序
	counterBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(counterBytes, uint64(counter))
	
	// 使用HMAC-SHA1计算哈希
	h := hmac.New(sha1.New, key)
	h.Write(counterBytes)
	hash := h.Sum(nil)
	
	// 动态截取
	offset := hash[len(hash)-1] & 0x0F
	truncated := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF
	
	// 生成指定位数的验证码
	code := truncated % uint32(pow10(t.digits))
	
	// 格式化为指定位数的字符串（前导零）
	format := fmt.Sprintf("%%0%dd", t.digits)
	return fmt.Sprintf(format, code)
}

// pow10 计算10的n次方
func pow10(n int) int {
	result := 1
	for i := 0; i < n; i++ {
		result *= 10
	}
	return result
}

// BackupCodeManager 备用码管理器
type BackupCodeManager struct {
	codeLength int
	codeCount  int
}

// NewBackupCodeManager 创建备用码管理器
func NewBackupCodeManager() *BackupCodeManager {
	return &BackupCodeManager{
		codeLength: 8,  // 8位备用码
		codeCount:  10, // 10个备用码
	}
}

// GenerateBackupCodes 生成备用码
func (b *BackupCodeManager) GenerateBackupCodes(count int) ([]string, error) {
	if count <= 0 {
		count = b.codeCount
	}
	
	codes := make([]string, count)
	
	for i := 0; i < count; i++ {
		code, err := b.generateSingleCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %v", err)
		}
		codes[i] = code
	}
	
	return codes, nil
}

// generateSingleCode 生成单个备用码
func (b *BackupCodeManager) generateSingleCode() (string, error) {
	// 使用数字和字母（排除容易混淆的字符）
	charset := "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	
	bytes := make([]byte, b.codeLength)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	
	code := make([]byte, b.codeLength)
	for i, b := range bytes {
		code[i] = charset[int(b)%len(charset)]
	}
	
	// 格式化为 XXXX-XXXX 的形式
	if b.codeLength == 8 {
		return fmt.Sprintf("%s-%s", string(code[:4]), string(code[4:])), nil
	}
	
	return string(code), nil
}

// ValidateBackupCode 验证备用码
func (b *BackupCodeManager) ValidateBackupCode(code string, validCodes []string) (bool, []string) {
	// 标准化输入码（移除空格和连字符，转大写）
	normalizedInput := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(code, " ", ""), "-", ""))
	
	// 查找匹配的备用码
	for i, validCode := range validCodes {
		normalizedValid := strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(validCode, " ", ""), "-", ""))
		
		if normalizedInput == normalizedValid {
			// 找到匹配的码，从列表中移除它
			newCodes := make([]string, 0, len(validCodes)-1)
			newCodes = append(newCodes, validCodes[:i]...)
			newCodes = append(newCodes, validCodes[i+1:]...)
			
			return true, newCodes
		}
	}
	
	return false, validCodes
}

// FormatBackupCode 格式化备用码显示
func (b *BackupCodeManager) FormatBackupCode(code string) string {
	// 移除现有格式
	clean := strings.ReplaceAll(strings.ReplaceAll(code, " ", ""), "-", "")
	
	// 重新格式化
	if len(clean) == 8 {
		return fmt.Sprintf("%s-%s", clean[:4], clean[4:])
	}
	
	return clean
}

// QRCodeGenerator QR码生成器接口
type QRCodeGenerator interface {
	GenerateQRCode(data string, size int) ([]byte, error)
}

// SimpleQRCodeGenerator 简单的QR码生成器（用于演示）
type SimpleQRCodeGenerator struct{}

// GenerateQRCode 生成QR码（这里返回URL，实际应用中可以集成QR码库）
func (g *SimpleQRCodeGenerator) GenerateQRCode(data string, size int) ([]byte, error) {
	// 这里可以集成第三方QR码生成库
	// 为了简化，我们返回一个指向在线QR码生成器的URL
	qrURL := fmt.Sprintf("https://api.qrserver.com/v1/create-qr-code/?size=%dx%d&data=%s",
		size, size, url.QueryEscape(data))
	
	return []byte(qrURL), nil
}

// GetCurrentTOTPCode 获取当前时间的TOTP验证码（用于测试）
func (t *TOTPManager) GetCurrentTOTPCode(secret string) (string, error) {
	return t.GenerateCode(secret, time.Now().Unix())
}

// GetTOTPCodeForTime 获取指定时间的TOTP验证码（用于测试）
func (t *TOTPManager) GetTOTPCodeForTime(secret string, timestamp int64) (string, error) {
	return t.GenerateCode(secret, timestamp)
}

// GetCurrentTimeStep 获取当前时间步（用于调试）
func (t *TOTPManager) GetCurrentTimeStep() int64 {
	return time.Now().Unix() / int64(t.period)
}

// ValidateCodeDebug 验证TOTP验证码（带调试信息）
func (t *TOTPManager) ValidateCodeDebug(secret, code string) (bool, map[string]interface{}) {
	now := time.Now().Unix()
	currentTimeStep := now / int64(t.period)

	debugInfo := map[string]interface{}{
		"current_time":      now,
		"current_timestep":  currentTimeStep,
		"period":           t.period,
		"window":           2,
		"tested_codes":     []map[string]interface{}{},
	}

	// 检查当前时间步和前后窗口内的验证码
	for i := -2; i <= 2; i++ {
		timeStep := currentTimeStep + int64(i)
		timestamp := timeStep * int64(t.period)
		expectedCode, err := t.GenerateCode(secret, timestamp)

		codeInfo := map[string]interface{}{
			"timestep":      timeStep,
			"timestamp":     timestamp,
			"expected_code": expectedCode,
			"matches":       code == expectedCode,
			"error":         nil,
		}

		if err != nil {
			codeInfo["error"] = err.Error()
		}

		debugInfo["tested_codes"] = append(debugInfo["tested_codes"].([]map[string]interface{}), codeInfo)

		if err == nil && code == expectedCode {
			debugInfo["match_found"] = true
			debugInfo["match_timestep"] = timeStep
			return true, debugInfo
		}
	}

	debugInfo["match_found"] = false
	return false, debugInfo
}
