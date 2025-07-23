package utils

import (
	"regexp"
	"strings"
)

// ANSI颜色代码映射
var ansiColorMap = map[string]string{
	"30": "black",
	"31": "red",
	"32": "green",
	"33": "yellow",
	"34": "blue",
	"35": "magenta",
	"36": "cyan",
	"37": "white",
	"90": "bright-black",
	"91": "bright-red",
	"92": "bright-green",
	"93": "bright-yellow",
	"94": "bright-blue",
	"95": "bright-magenta",
	"96": "bright-cyan",
	"97": "bright-white",
}

// ANSI背景颜色代码映射
var ansiBgColorMap = map[string]string{
	"40": "bg-black",
	"41": "bg-red",
	"42": "bg-green",
	"43": "bg-yellow",
	"44": "bg-blue",
	"45": "bg-magenta",
	"46": "bg-cyan",
	"47": "bg-white",
	"100": "bg-bright-black",
	"101": "bg-bright-red",
	"102": "bg-bright-green",
	"103": "bg-bright-yellow",
	"104": "bg-bright-blue",
	"105": "bg-bright-magenta",
	"106": "bg-bright-cyan",
	"107": "bg-bright-white",
}

// ANSI样式代码映射
var ansiStyleMap = map[string]string{
	"0":  "reset",
	"1":  "bold",
	"2":  "dim",
	"3":  "italic",
	"4":  "underline",
	"5":  "blink",
	"7":  "reverse",
	"8":  "hidden",
	"9":  "strikethrough",
	"22": "normal",
	"23": "no-italic",
	"24": "no-underline",
	"25": "no-blink",
	"27": "no-reverse",
	"28": "no-hidden",
	"29": "no-strikethrough",
}

// ANSIToHTML 将ANSI颜色代码转换为HTML
func ANSIToHTML(text string) string {
	// ANSI转义序列的正则表达式
	ansiRegex := regexp.MustCompile(`\x1b\[([0-9;]*)m`)
	
	// 当前活跃的样式
	var activeClasses []string
	var openSpans int
	
	result := ansiRegex.ReplaceAllStringFunc(text, func(match string) string {
		// 提取ANSI代码
		codes := ansiRegex.FindStringSubmatch(match)
		if len(codes) < 2 {
			return ""
		}
		
		codeStr := codes[1]
		if codeStr == "" {
			codeStr = "0" // 默认重置
		}
		
		// 分割多个代码
		codeParts := strings.Split(codeStr, ";")
		
		var html strings.Builder
		
		for _, code := range codeParts {
			switch code {
			case "0", "": // 重置
				// 关闭所有打开的span
				for i := 0; i < openSpans; i++ {
					html.WriteString("</span>")
				}
				openSpans = 0
				activeClasses = nil
				
			default:
				// 处理颜色和样式代码
				if className := getClassNameForCode(code); className != "" {
					// 检查是否已经有这个类
					if !contains(activeClasses, className) {
						activeClasses = append(activeClasses, className)
						html.WriteString(`<span class="ansi-` + className + `">`)
						openSpans++
					}
				}
			}
		}
		
		return html.String()
	})
	
	// 在文本末尾关闭所有打开的span
	for i := 0; i < openSpans; i++ {
		result += "</span>"
	}
	
	return result
}

// StripANSI 移除ANSI颜色代码
func StripANSI(text string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return ansiRegex.ReplaceAllString(text, "")
}

// getClassNameForCode 根据ANSI代码获取CSS类名
func getClassNameForCode(code string) string {
	// 检查前景色
	if color, exists := ansiColorMap[code]; exists {
		return color
	}
	
	// 检查背景色
	if bgColor, exists := ansiBgColorMap[code]; exists {
		return bgColor
	}
	
	// 检查样式
	if style, exists := ansiStyleMap[code]; exists {
		return style
	}
	
	return ""
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ParseMinecraftLog 解析Minecraft日志并提取关键信息
func ParseMinecraftLog(line string) map[string]interface{} {
	result := map[string]interface{}{
		"raw":       line,
		"message":   StripANSI(line),
		"level":     "INFO",
		"timestamp": "",
		"thread":    "",
		"logger":    "",
	}
	
	// Minecraft日志格式: [时间] [线程/级别] [日志器]: 消息
	logRegex := regexp.MustCompile(`^\[([^\]]+)\] \[([^/]+)/([^\]]+)\] \[([^\]]+)\]: (.+)$`)
	matches := logRegex.FindStringSubmatch(StripANSI(line))
	
	if len(matches) == 6 {
		result["timestamp"] = matches[1]
		result["thread"] = matches[2]
		result["level"] = matches[3]
		result["logger"] = matches[4]
		result["message"] = matches[5]
	} else {
		// 尝试简化格式: [级别]: 消息
		simpleRegex := regexp.MustCompile(`^\[([^\]]+)\]: (.+)$`)
		simpleMatches := simpleRegex.FindStringSubmatch(StripANSI(line))
		if len(simpleMatches) == 3 {
			result["level"] = simpleMatches[1]
			result["message"] = simpleMatches[2]
		}
	}
	
	// 根据内容判断日志级别
	message := strings.ToLower(result["message"].(string))
	if strings.Contains(message, "error") || strings.Contains(message, "exception") || 
	   strings.Contains(message, "failed") || strings.Contains(message, "crash") {
		result["level"] = "ERROR"
	} else if strings.Contains(message, "warn") {
		result["level"] = "WARN"
	} else if strings.Contains(message, "debug") {
		result["level"] = "DEBUG"
	}
	
	return result
}

// GetLogLevelColor 根据日志级别获取颜色类
func GetLogLevelColor(level string) string {
	switch strings.ToUpper(level) {
	case "ERROR", "SEVERE":
		return "log-error"
	case "WARN", "WARNING":
		return "log-warning"
	case "INFO":
		return "log-info"
	case "DEBUG":
		return "log-debug"
	case "TRACE":
		return "log-trace"
	default:
		return "log-default"
	}
}

// FormatLogForWeb 格式化日志用于Web显示
func FormatLogForWeb(line string) map[string]interface{} {
	parsed := ParseMinecraftLog(line)
	
	return map[string]interface{}{
		"raw":        line,
		"html":       ANSIToHTML(line),
		"plain":      parsed["message"],
		"level":      parsed["level"],
		"levelColor": GetLogLevelColor(parsed["level"].(string)),
		"timestamp":  parsed["timestamp"],
		"thread":     parsed["thread"],
		"logger":     parsed["logger"],
	}
}
