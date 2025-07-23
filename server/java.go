package server

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"easilypanel5/config"
)

// detectJavaEnvironment 检测系统中的Java环境
func detectJavaEnvironment() ([]config.JavaInfo, error) {
	var javaInfos []config.JavaInfo

	// 检查环境变量中的Java
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		javaPath := filepath.Join(javaHome, "bin", "java")
		if runtime.GOOS == "windows" {
			javaPath += ".exe"
		}
		
		if info := checkJavaPath(javaPath); info.IsValid {
			javaInfos = append(javaInfos, info)
		}
	}

	// 检查PATH中的Java
	if javaPath, err := exec.LookPath("java"); err == nil {
		if info := checkJavaPath(javaPath); info.IsValid {
			// 避免重复添加
			found := false
			for _, existing := range javaInfos {
				if existing.Path == info.Path {
					found = true
					break
				}
			}
			if !found {
				javaInfos = append(javaInfos, info)
			}
		}
	}

	// 检查常见的Java安装路径
	commonPaths := getCommonJavaPaths()
	for _, path := range commonPaths {
		if info := checkJavaPath(path); info.IsValid {
			// 避免重复添加
			found := false
			for _, existing := range javaInfos {
				if existing.Path == info.Path {
					found = true
					break
				}
			}
			if !found {
				javaInfos = append(javaInfos, info)
			}
		}
	}

	if len(javaInfos) == 0 {
		return nil, fmt.Errorf("no valid Java installation found")
	}

	return javaInfos, nil
}

// validateJavaPath 验证指定路径的Java
func validateJavaPath(javaPath string) (*config.JavaInfo, error) {
	info := checkJavaPath(javaPath)
	if !info.IsValid {
		return nil, fmt.Errorf("invalid Java: %s", info.Error)
	}
	return &info, nil
}

// checkJavaPath 检查指定路径的Java
func checkJavaPath(javaPath string) config.JavaInfo {
	info := config.JavaInfo{
		Path:    javaPath,
		IsValid: false,
	}

	// 检查文件是否存在
	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		info.Error = "Java executable not found"
		return info
	}

	// 执行java -version命令
	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		info.Error = fmt.Sprintf("Failed to execute java -version: %v", err)
		return info
	}

	// 解析版本信息
	versionStr := string(output)
	info.Version = parseJavaVersion(versionStr)
	info.Vendor = parseJavaVendor(versionStr)
	info.Architecture = parseJavaArchitecture(versionStr)

	// 验证版本是否满足最低要求
	if !isValidJavaVersion(info.Version) {
		info.Error = fmt.Sprintf("Java version %s is not supported (minimum: Java 8)", info.Version)
		return info
	}

	info.IsValid = true
	return info
}

// parseJavaVersion 解析Java版本
func parseJavaVersion(versionStr string) string {
	// 匹配版本号模式
	patterns := []string{
		`version "(\d+\.\d+\.\d+[^"]*)"`,  // Java 8及以下: "1.8.0_XXX"
		`version "(\d+[^"]*)"`,           // Java 9及以上: "11.0.1"
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(versionStr); len(matches) > 1 {
			version := matches[1]
			// 标准化版本格式
			if strings.HasPrefix(version, "1.") {
				// Java 8及以下，提取主版本号
				parts := strings.Split(version, ".")
				if len(parts) >= 2 {
					return "8" // 1.8.x -> 8
				}
			} else {
				// Java 9及以上，提取主版本号
				parts := strings.Split(version, ".")
				if len(parts) > 0 {
					return parts[0]
				}
			}
			return version
		}
	}

	return "unknown"
}

// parseJavaVendor 解析Java供应商
func parseJavaVendor(versionStr string) string {
	vendors := map[string]string{
		"OpenJDK":     "OpenJDK",
		"Oracle":      "Oracle",
		"Adoptium":    "Eclipse Adoptium",
		"AdoptOpenJDK": "AdoptOpenJDK",
		"Amazon":      "Amazon Corretto",
		"Azul":        "Azul Zulu",
		"IBM":         "IBM",
		"Microsoft":   "Microsoft",
	}

	lowerStr := strings.ToLower(versionStr)
	for key, vendor := range vendors {
		if strings.Contains(lowerStr, strings.ToLower(key)) {
			return vendor
		}
	}

	return "Unknown"
}

// parseJavaArchitecture 解析Java架构
func parseJavaArchitecture(versionStr string) string {
	if strings.Contains(versionStr, "64-Bit") || strings.Contains(versionStr, "amd64") {
		return "64-bit"
	} else if strings.Contains(versionStr, "32-Bit") || strings.Contains(versionStr, "i386") {
		return "32-bit"
	}
	return runtime.GOARCH
}

// isValidJavaVersion 检查Java版本是否有效
func isValidJavaVersion(version string) bool {
	if version == "unknown" {
		return false
	}

	// 提取主版本号
	var majorVersion int
	if strings.Contains(version, ".") {
		parts := strings.Split(version, ".")
		if len(parts) > 0 {
			if v, err := strconv.Atoi(parts[0]); err == nil {
				majorVersion = v
			}
		}
	} else {
		if v, err := strconv.Atoi(version); err == nil {
			majorVersion = v
		}
	}

	// Java 8及以上版本
	return majorVersion >= 8
}

// getCommonJavaPaths 获取常见的Java安装路径
func getCommonJavaPaths() []string {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Windows常见路径
		programFiles := []string{
			os.Getenv("ProgramFiles"),
			os.Getenv("ProgramFiles(x86)"),
			"C:\\Program Files",
			"C:\\Program Files (x86)",
		}

		for _, pf := range programFiles {
			if pf == "" {
				continue
			}
			
			// Oracle JDK/JRE
			javaDir := filepath.Join(pf, "Java")
			if entries, err := os.ReadDir(javaDir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						javaPath := filepath.Join(javaDir, entry.Name(), "bin", "java.exe")
						paths = append(paths, javaPath)
					}
				}
			}

			// OpenJDK
			openJDKDirs := []string{"OpenJDK", "Eclipse Adoptium", "Zulu"}
			for _, dir := range openJDKDirs {
				javaDir := filepath.Join(pf, dir)
				if entries, err := os.ReadDir(javaDir); err == nil {
					for _, entry := range entries {
						if entry.IsDir() {
							javaPath := filepath.Join(javaDir, entry.Name(), "bin", "java.exe")
							paths = append(paths, javaPath)
						}
					}
				}
			}
		}

	case "darwin":
		// macOS常见路径
		commonDirs := []string{
			"/Library/Java/JavaVirtualMachines",
			"/System/Library/Java/JavaVirtualMachines",
			"/usr/libexec/java_home",
		}

		for _, dir := range commonDirs {
			if entries, err := os.ReadDir(dir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() && strings.HasSuffix(entry.Name(), ".jdk") {
						javaPath := filepath.Join(dir, entry.Name(), "Contents", "Home", "bin", "java")
						paths = append(paths, javaPath)
					}
				}
			}
		}

	case "linux":
		// Linux常见路径
		commonDirs := []string{
			"/usr/lib/jvm",
			"/usr/java",
			"/opt/java",
			"/opt/jdk",
			"/usr/local/java",
		}

		for _, dir := range commonDirs {
			if entries, err := os.ReadDir(dir); err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						javaPath := filepath.Join(dir, entry.Name(), "bin", "java")
						paths = append(paths, javaPath)
					}
				}
			}
		}
	}

	return paths
}

// GetRecommendedJava 获取推荐的Java配置
func GetRecommendedJava(mcVersion string) (*config.JavaInfo, error) {
	javaInfos, err := detectJavaEnvironment()
	if err != nil {
		return nil, err
	}

	// 根据MC版本推荐Java版本
	var recommendedVersion int
	switch {
	case strings.HasPrefix(mcVersion, "1.17") || strings.HasPrefix(mcVersion, "1.18") || 
		 strings.HasPrefix(mcVersion, "1.19") || strings.HasPrefix(mcVersion, "1.20"):
		recommendedVersion = 17 // MC 1.17+ 推荐Java 17
	case strings.HasPrefix(mcVersion, "1.16"):
		recommendedVersion = 11 // MC 1.16 推荐Java 11
	default:
		recommendedVersion = 8  // 其他版本推荐Java 8
	}

	// 查找最匹配的Java版本
	var bestMatch *config.JavaInfo
	for i, info := range javaInfos {
		if !info.IsValid {
			continue
		}

		version, _ := strconv.Atoi(info.Version)
		if version == recommendedVersion {
			bestMatch = &javaInfos[i]
			break
		}

		if bestMatch == nil || (version >= recommendedVersion && version < getVersionInt(bestMatch.Version)) {
			bestMatch = &javaInfos[i]
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no suitable Java version found")
	}

	return bestMatch, nil
}

// getVersionInt 获取版本的整数值
func getVersionInt(version string) int {
	if v, err := strconv.Atoi(version); err == nil {
		return v
	}
	return 0
}
