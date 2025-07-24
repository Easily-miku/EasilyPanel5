package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"easilypanel5/config"
)

// LogLevel 日志级别
type LogLevel int

const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LevelTrace:
		return "TRACE"
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel 解析日志级别字符串
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "TRACE":
		return LevelTrace
	case "DEBUG":
		return LevelDebug
	case "INFO":
		return LevelInfo
	case "WARN", "WARNING":
		return LevelWarn
	case "ERROR":
		return LevelError
	case "FATAL":
		return LevelFatal
	default:
		return LevelInfo
	}
}

// Logger 日志记录器
type Logger struct {
	level      LogLevel
	output     io.Writer
	file       *os.File
	mutex      sync.RWMutex
	maxSize    int64
	maxFiles   int
	maxAge     int
	logDir     string
	filename   string
	enableColors bool
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	File      string    `json:"file,omitempty"`
	Line      int       `json:"line,omitempty"`
	Function  string    `json:"function,omitempty"`
}

var (
	defaultLogger *Logger
	loggerOnce    sync.Once
)

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	loggerOnce.Do(func() {
		cfg := config.Get()
		defaultLogger = NewLogger(&LoggerConfig{
			Level:        ParseLogLevel(cfg.Logging.Level),
			LogDir:       cfg.Logging.LogsDir,
			MaxSize:      cfg.Logging.MaxFileSize,
			MaxFiles:     cfg.Logging.MaxFiles,
			MaxAge:       cfg.Logging.MaxAge,
			EnableColors: cfg.Logging.EnableColors,
		})
	})
	return defaultLogger
}

// LoggerConfig 日志记录器配置
type LoggerConfig struct {
	Level        LogLevel
	LogDir       string
	MaxSize      int64
	MaxFiles     int
	MaxAge       int
	EnableColors bool
}

// NewLogger 创建新的日志记录器
func NewLogger(config *LoggerConfig) *Logger {
	logger := &Logger{
		level:        config.Level,
		maxSize:      config.MaxSize,
		maxFiles:     config.MaxFiles,
		maxAge:       config.MaxAge,
		logDir:       config.LogDir,
		filename:     "easilypanel5.log",
		enableColors: config.EnableColors,
	}

	// 创建日志目录
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		logger.output = os.Stdout
		return logger
	}

	// 打开日志文件
	if err := logger.openLogFile(); err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		logger.output = os.Stdout
		return logger
	}

	return logger
}

// openLogFile 打开日志文件
func (l *Logger) openLogFile() error {
	logPath := filepath.Join(l.logDir, l.filename)
	
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	l.file = file
	l.output = file
	return nil
}

// shouldRotate 检查是否需要轮转日志
func (l *Logger) shouldRotate() bool {
	if l.file == nil {
		return false
	}

	stat, err := l.file.Stat()
	if err != nil {
		return false
	}

	return stat.Size() > l.maxSize
}

// rotateLog 轮转日志文件
func (l *Logger) rotateLog() error {
	if l.file == nil {
		return nil
	}

	// 关闭当前文件
	l.file.Close()

	// 重命名当前文件
	currentPath := filepath.Join(l.logDir, l.filename)
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	archivedPath := filepath.Join(l.logDir, fmt.Sprintf("easilypanel5-%s.log", timestamp))

	if err := os.Rename(currentPath, archivedPath); err != nil {
		return err
	}

	// 打开新文件
	if err := l.openLogFile(); err != nil {
		return err
	}

	// 清理旧文件
	go l.cleanupOldLogs()

	return nil
}

// cleanupOldLogs 清理旧的日志文件
func (l *Logger) cleanupOldLogs() {
	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return
	}

	var logFiles []os.FileInfo
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "easilypanel5-") && strings.HasSuffix(entry.Name(), ".log") {
			info, err := entry.Info()
			if err == nil {
				logFiles = append(logFiles, info)
			}
		}
	}

	// 按修改时间排序
	for i := 0; i < len(logFiles)-1; i++ {
		for j := i + 1; j < len(logFiles); j++ {
			if logFiles[i].ModTime().Before(logFiles[j].ModTime()) {
				logFiles[i], logFiles[j] = logFiles[j], logFiles[i]
			}
		}
	}

	// 删除超出限制的文件
	for i := l.maxFiles; i < len(logFiles); i++ {
		filePath := filepath.Join(l.logDir, logFiles[i].Name())
		os.Remove(filePath)
	}

	// 删除过期文件
	cutoff := time.Now().AddDate(0, 0, -l.maxAge)
	for _, file := range logFiles {
		if file.ModTime().Before(cutoff) {
			filePath := filepath.Join(l.logDir, file.Name())
			os.Remove(filePath)
		}
	}
}

// log 记录日志
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()

	// 检查是否需要轮转
	if l.shouldRotate() {
		l.rotateLog()
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)

	var logLine string
	if l.enableColors && l.output == os.Stdout {
		color := l.getLevelColor(level)
		logLine = fmt.Sprintf("%s [%s%s\033[0m] %s\n", timestamp, color, level.String(), message)
	} else {
		logLine = fmt.Sprintf("%s [%s] %s\n", timestamp, level.String(), message)
	}

	if l.output != nil {
		l.output.Write([]byte(logLine))
		if f, ok := l.output.(*os.File); ok {
			f.Sync()
		}
	}
}

// getLevelColor 获取日志级别对应的颜色代码
func (l *Logger) getLevelColor(level LogLevel) string {
	switch level {
	case LevelTrace:
		return "\033[37m" // 白色
	case LevelDebug:
		return "\033[36m" // 青色
	case LevelInfo:
		return "\033[32m" // 绿色
	case LevelWarn:
		return "\033[33m" // 黄色
	case LevelError:
		return "\033[31m" // 红色
	case LevelFatal:
		return "\033[35m" // 紫色
	default:
		return "\033[0m" // 默认
	}
}

// 便捷方法
func (l *Logger) Trace(format string, args ...interface{}) {
	l.log(LevelTrace, format, args...)
}

func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// 全局便捷函数
func Trace(format string, args ...interface{}) {
	GetLogger().Trace(format, args...)
}

func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}

// Close 关闭日志记录器
func (l *Logger) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.level = level
}

// GetLevel 获取当前日志级别
func (l *Logger) GetLevel() LogLevel {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.level
}
