package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 中文化日志器
type Logger struct {
	*logrus.Logger
}

// ChineseFormatter 中文格式化器
type ChineseFormatter struct {
	TimestampFormat string
}

// Format 格式化日志输出
func (f *ChineseFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timestamp := entry.Time.Format(f.TimestampFormat)
	
	// 中文化日志级别
	var levelText string
	switch entry.Level {
	case logrus.PanicLevel:
		levelText = "严重错误"
	case logrus.FatalLevel:
		levelText = "致命错误"
	case logrus.ErrorLevel:
		levelText = "错误"
	case logrus.WarnLevel:
		levelText = "警告"
	case logrus.InfoLevel:
		levelText = "信息"
	case logrus.DebugLevel:
		levelText = "调试"
	case logrus.TraceLevel:
		levelText = "跟踪"
	default:
		levelText = "未知"
	}
	
	// 构建日志消息
	message := fmt.Sprintf("[%s] [%s] %s", timestamp, levelText, entry.Message)
	
	// 添加字段信息
	if len(entry.Data) > 0 {
		message += " |"
		for key, value := range entry.Data {
			message += fmt.Sprintf(" %s=%v", key, value)
		}
	}
	
	message += "\n"
	return []byte(message), nil
}

// NewLogger 创建新的中文化日志器
func NewLogger() *Logger {
	log := logrus.New()
	
	// 设置中文格式化器
	log.SetFormatter(&ChineseFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	// 设置日志级别
	log.SetLevel(logrus.InfoLevel)
	
	// 设置输出到控制台
	log.SetOutput(os.Stdout)
	
	return &Logger{Logger: log}
}

// NewLoggerWithConfig 根据配置创建日志器
func NewLoggerWithConfig(logFile string, level string, maxSize, maxBackups, maxAge int, compress bool) (*Logger, error) {
	log := logrus.New()
	
	// 设置中文格式化器
	log.SetFormatter(&ChineseFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
	
	// 设置日志级别
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	log.SetLevel(logLevel)
	
	// 确保日志目录存在
	if err := os.MkdirAll(filepath.Dir(logFile), 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	
	// 配置日志轮转
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    maxSize,    // MB
		MaxBackups: maxBackups,
		MaxAge:     maxAge,     // days
		Compress:   compress,
	}
	
	// 同时输出到文件和控制台
	multiWriter := io.MultiWriter(os.Stdout, lumberjackLogger)
	log.SetOutput(multiWriter)
	
	return &Logger{Logger: log}, nil
}

// 中文化的日志方法

// 信息 记录信息级别日志
func (l *Logger) 信息(format string, args ...interface{}) {
	l.Infof(format, args...)
}

// 警告 记录警告级别日志
func (l *Logger) 警告(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

// 错误 记录错误级别日志
func (l *Logger) 错误(format string, args ...interface{}) {
	l.Errorf(format, args...)
}

// 调试 记录调试级别日志
func (l *Logger) 调试(format string, args ...interface{}) {
	l.Debugf(format, args...)
}

// 致命错误 记录致命错误并退出程序
func (l *Logger) 致命错误(format string, args ...interface{}) {
	l.Fatalf(format, args...)
}

// 带字段的日志方法

// 信息字段 记录带字段的信息日志
func (l *Logger) 信息字段(fields map[string]interface{}, format string, args ...interface{}) {
	l.WithFields(logrus.Fields(fields)).Infof(format, args...)
}

// 警告字段 记录带字段的警告日志
func (l *Logger) 警告字段(fields map[string]interface{}, format string, args ...interface{}) {
	l.WithFields(logrus.Fields(fields)).Warnf(format, args...)
}

// 错误字段 记录带字段的错误日志
func (l *Logger) 错误字段(fields map[string]interface{}, format string, args ...interface{}) {
	l.WithFields(logrus.Fields(fields)).Errorf(format, args...)
}

// 操作日志方法

// 开始操作 记录操作开始
func (l *Logger) 开始操作(operation string, details map[string]interface{}) {
	fields := map[string]interface{}{
		"操作": operation,
		"状态": "开始",
		"时间": time.Now().Format("15:04:05"),
	}
	for k, v := range details {
		fields[k] = v
	}
	l.信息字段(fields, "开始执行操作: %s", operation)
}

// 完成操作 记录操作完成
func (l *Logger) 完成操作(operation string, duration time.Duration, details map[string]interface{}) {
	fields := map[string]interface{}{
		"操作": operation,
		"状态": "完成",
		"耗时": duration.String(),
	}
	for k, v := range details {
		fields[k] = v
	}
	l.信息字段(fields, "操作完成: %s (耗时: %s)", operation, duration.String())
}

// 操作失败 记录操作失败
func (l *Logger) 操作失败(operation string, err error, details map[string]interface{}) {
	fields := map[string]interface{}{
		"操作": operation,
		"状态": "失败",
		"错误": err.Error(),
	}
	for k, v := range details {
		fields[k] = v
	}
	l.错误字段(fields, "操作失败: %s - %v", operation, err)
}

// 系统事件日志

// 系统启动 记录系统启动事件
func (l *Logger) 系统启动(version string) {
	l.信息字段(map[string]interface{}{
		"版本": version,
		"事件": "系统启动",
	}, "EasilyPanel5 系统启动 - 版本: %s", version)
}

// 系统关闭 记录系统关闭事件
func (l *Logger) 系统关闭() {
	l.信息字段(map[string]interface{}{
		"事件": "系统关闭",
	}, "EasilyPanel5 系统正常关闭")
}

// 配置加载 记录配置加载事件
func (l *Logger) 配置加载(configFile string) {
	l.信息字段(map[string]interface{}{
		"配置文件": configFile,
		"事件":   "配置加载",
	}, "配置文件加载成功: %s", configFile)
}

// 实例事件日志

// 实例创建 记录实例创建事件
func (l *Logger) 实例创建(name, instanceType string) {
	l.信息字段(map[string]interface{}{
		"实例名称": name,
		"实例类型": instanceType,
		"事件":   "实例创建",
	}, "实例创建成功: %s (%s)", name, instanceType)
}

// 实例启动 记录实例启动事件
func (l *Logger) 实例启动(name string, pid int) {
	l.信息字段(map[string]interface{}{
		"实例名称": name,
		"进程ID":  pid,
		"事件":   "实例启动",
	}, "实例启动成功: %s (PID: %d)", name, pid)
}

// 实例停止 记录实例停止事件
func (l *Logger) 实例停止(name string) {
	l.信息字段(map[string]interface{}{
		"实例名称": name,
		"事件":   "实例停止",
	}, "实例停止成功: %s", name)
}
