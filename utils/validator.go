package utils

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors 多个验证错误
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	
	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// Validator 验证器接口
type Validator interface {
	Validate(value interface{}) error
}

// ValidateRequired 验证必填字段
func ValidateRequired(field string, value interface{}) error {
	if value == nil {
		return ValidationError{
			Field:   field,
			Message: "This field is required",
			Code:    "REQUIRED",
		}
	}

	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return ValidationError{
				Field:   field,
				Message: "This field cannot be empty",
				Code:    "REQUIRED",
			}
		}
	case []string:
		if len(v) == 0 {
			return ValidationError{
				Field:   field,
				Message: "This field must contain at least one item",
				Code:    "REQUIRED",
			}
		}
	}

	return nil
}

// ValidateStringLength 验证字符串长度
func ValidateStringLength(field string, value string, min, max int) error {
	length := len(value)
	
	if min > 0 && length < min {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Must be at least %d characters long", min),
			Code:    "MIN_LENGTH",
		}
	}
	
	if max > 0 && length > max {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Must be no more than %d characters long", max),
			Code:    "MAX_LENGTH",
		}
	}
	
	return nil
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(field string, email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return ValidationError{
			Field:   field,
			Message: "Invalid email format",
			Code:    "INVALID_EMAIL",
		}
	}
	return nil
}

// ValidatePort 验证端口号
func ValidatePort(field string, port int) error {
	if port < 1 || port > 65535 {
		return ValidationError{
			Field:   field,
			Message: "Port must be between 1 and 65535",
			Code:    "INVALID_PORT",
		}
	}
	return nil
}

// ValidateIP 验证IP地址
func ValidateIP(field string, ip string) error {
	if net.ParseIP(ip) == nil {
		return ValidationError{
			Field:   field,
			Message: "Invalid IP address format",
			Code:    "INVALID_IP",
		}
	}
	return nil
}

// ValidateMemorySize 验证内存大小（MB）
func ValidateMemorySize(field string, memory int) error {
	if memory < 512 {
		return ValidationError{
			Field:   field,
			Message: "Memory must be at least 512MB",
			Code:    "INVALID_MEMORY_SIZE",
		}
	}
	
	if memory > 32768 { // 32GB
		return ValidationError{
			Field:   field,
			Message: "Memory cannot exceed 32GB",
			Code:    "INVALID_MEMORY_SIZE",
		}
	}
	
	return nil
}

// ValidateServerName 验证服务器名称
func ValidateServerName(field string, name string) error {
	if err := ValidateRequired(field, name); err != nil {
		return err
	}
	
	if err := ValidateStringLength(field, name, 1, 50); err != nil {
		return err
	}
	
	// 检查是否包含非法字符
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '-' && char != '_' && char != ' ' {
			return ValidationError{
				Field:   field,
				Message: "Server name can only contain letters, numbers, spaces, hyphens and underscores",
				Code:    "INVALID_CHARACTERS",
			}
		}
	}
	
	return nil
}

// ValidateJavaArgs 验证Java启动参数
func ValidateJavaArgs(field string, args []string) error {
	for i, arg := range args {
		if strings.TrimSpace(arg) == "" {
			return ValidationError{
				Field:   fmt.Sprintf("%s[%d]", field, i),
				Message: "Java argument cannot be empty",
				Code:    "INVALID_JAVA_ARG",
			}
		}
		
		// 检查危险参数
		dangerousArgs := []string{
			"-XX:+UnlockExperimentalVMOptions",
			"-XX:+UseZGC",
			"-XX:+UseShenandoahGC",
		}
		
		for _, dangerous := range dangerousArgs {
			if strings.Contains(arg, dangerous) {
				return ValidationError{
					Field:   fmt.Sprintf("%s[%d]", field, i),
					Message: fmt.Sprintf("Potentially dangerous Java argument: %s", dangerous),
					Code:    "DANGEROUS_JAVA_ARG",
				}
			}
		}
	}
	
	return nil
}

// ValidateMinecraftVersion 验证Minecraft版本格式
func ValidateMinecraftVersion(field string, version string) error {
	if err := ValidateRequired(field, version); err != nil {
		return err
	}
	
	// 基本版本格式验证 (例如: 1.20.1, 1.19.4)
	versionRegex := regexp.MustCompile(`^1\.\d+(\.\d+)?$`)
	if !versionRegex.MatchString(version) {
		return ValidationError{
			Field:   field,
			Message: "Invalid Minecraft version format (expected: 1.x.x)",
			Code:    "INVALID_MC_VERSION",
		}
	}
	
	return nil
}

// ValidateCoreType 验证核心类型
func ValidateCoreType(field string, coreType string) error {
	validCoreTypes := []string{
		"vanilla", "paper", "spigot", "bukkit", "forge", "fabric", "quilt",
	}
	
	coreType = strings.ToLower(coreType)
	for _, valid := range validCoreTypes {
		if coreType == valid {
			return nil
		}
	}
	
	return ValidationError{
		Field:   field,
		Message: fmt.Sprintf("Invalid core type. Valid types: %s", strings.Join(validCoreTypes, ", ")),
		Code:    "INVALID_CORE_TYPE",
	}
}

// ValidateFilePath 验证文件路径
func ValidateFilePath(field string, path string) error {
	if err := ValidateRequired(field, path); err != nil {
		return err
	}
	
	// 检查路径遍历攻击
	if strings.Contains(path, "..") {
		return ValidationError{
			Field:   field,
			Message: "Path traversal is not allowed",
			Code:    "INVALID_PATH",
		}
	}
	
	// 检查绝对路径（在某些情况下可能不安全）
	if strings.HasPrefix(path, "/") || strings.Contains(path, ":") {
		return ValidationError{
			Field:   field,
			Message: "Absolute paths are not allowed",
			Code:    "INVALID_PATH",
		}
	}
	
	return nil
}

// ValidateURL 验证URL格式
func ValidateURL(field string, url string) error {
	if err := ValidateRequired(field, url); err != nil {
		return err
	}
	
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return ValidationError{
			Field:   field,
			Message: "Invalid URL format",
			Code:    "INVALID_URL",
		}
	}
	
	return nil
}

// ValidateRange 验证数值范围
func ValidateRange(field string, value, min, max int) error {
	if value < min {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be at least %d", min),
			Code:    "VALUE_TOO_SMALL",
		}
	}
	
	if value > max {
		return ValidationError{
			Field:   field,
			Message: fmt.Sprintf("Value must be no more than %d", max),
			Code:    "VALUE_TOO_LARGE",
		}
	}
	
	return nil
}

// ValidateStringInSlice 验证字符串是否在允许的列表中
func ValidateStringInSlice(field string, value string, allowed []string) error {
	for _, item := range allowed {
		if value == item {
			return nil
		}
	}
	
	return ValidationError{
		Field:   field,
		Message: fmt.Sprintf("Invalid value. Allowed values: %s", strings.Join(allowed, ", ")),
		Code:    "INVALID_VALUE",
	}
}

// ValidateStruct 验证结构体（使用反射）
func ValidateStruct(data interface{}) ValidationErrors {
	// 这里可以实现基于反射的结构体验证
	// 暂时返回空错误列表
	return ValidationErrors{}
}

// ParseMemoryString 解析内存字符串（如 "2G", "1024M"）
func ParseMemoryString(memStr string) (int, error) {
	memStr = strings.TrimSpace(strings.ToUpper(memStr))
	
	if memStr == "" {
		return 0, fmt.Errorf("memory string cannot be empty")
	}
	
	// 提取数字部分和单位
	var numStr string
	var unit string
	
	for i, char := range memStr {
		if unicode.IsDigit(char) {
			numStr += string(char)
		} else {
			unit = memStr[i:]
			break
		}
	}
	
	if numStr == "" {
		return 0, fmt.Errorf("invalid memory format: %s", memStr)
	}
	
	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0, fmt.Errorf("invalid number in memory string: %s", numStr)
	}
	
	// 转换为MB
	switch unit {
	case "", "M", "MB":
		return num, nil
	case "G", "GB":
		return num * 1024, nil
	case "K", "KB":
		return num / 1024, nil
	default:
		return 0, fmt.Errorf("unsupported memory unit: %s", unit)
	}
}
