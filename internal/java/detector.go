package java

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// Java Java安装信息结构
type Java struct {
	Path    string `json:"path"`
	Version string `json:"version"`
}

// String 返回Java信息的字符串表示
func (j *Java) String() string {
	data, _ := json.Marshal(j)
	return string(data)
}

// Equal 比较两个Java实例是否相等
func (j *Java) Equal(other *Java) bool {
	return j.Path == other.Path && j.Version == other.Version
}

// Detector Java检测器
type Detector struct {
	matchKeywords   []string
	excludeKeywords []string
	foundJava       []*Java
}

// NewDetector 创建新的Java检测器
func NewDetector() *Detector {
	currentUser, _ := user.Current()
	username := currentUser.Username
	
	matchKeywords := []string{
		"1.", "bin", "cache", "client", "craft", "data", "download", "eclipse", "mine", "mc", "launch",
		"hotspot", "java", "jdk", "jre", "zulu", "dragonwell", "jvm", "microsoft", "corretto", "sigma",
		"mod", "mojang", "net", "netease", "forge", "liteloader", "fabric", "game", "vanilla", "server",
		"optifine", "oracle", "path", "program", "roaming", "local", "run", "runtime", "software", "daemon",
		"temp", "users", "x64", "x86", "lib", "usr", "env", "ext", "file", "data", "green", "vape",
		"我的", "世界", "前置", "原版", "启动", "国服", "官启", "官方", "客户", "应用", "整合",
		username, "新建文件夹", "服务", "游戏", "环境", "程序", "网易", "软件", "运行", "高清", "组件",
		"badlion", "blc", "lunar", "tlauncher", "soar", "cheatbreaker", "hmcl", "pcl", "bakaxl", "fsm",
		"jetbrains", "intellij", "idea", "pycharm", "webstorm", "clion", "goland", "rider", "datagrip",
		"appcode", "phpstorm", "rubymine", "jbr", "android", "mcsm", "msl", "mcsl", "3dmark", "arctime",
	}
	
	excludeKeywords := []string{"$", "{", "}", "__"}
	
	return &Detector{
		matchKeywords:   matchKeywords,
		excludeKeywords: excludeKeywords,
		foundJava:       make([]*Java, 0),
	}
}

// GetJavaVersion 获取Java版本信息
func (d *Detector) GetJavaVersion(javaPath string) string {
	cmd := exec.Command(javaPath, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return ""
	}
	
	// 从输出中提取版本信息
	versionPattern := regexp.MustCompile(`(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:[._](\d+))?(?:-(.+))?`)
	matches := versionPattern.FindStringSubmatch(string(output))
	
	if len(matches) > 1 {
		// 过滤空字符串并连接版本号
		var versionParts []string
		for i := 1; i < len(matches); i++ {
			if matches[i] != "" {
				versionParts = append(versionParts, matches[i])
			}
		}
		return strings.Join(versionParts, ".")
	}
	
	return ""
}

// findStr 检查字符串是否匹配关键词
func (d *Detector) findStr(s string) bool {
	s = strings.ToLower(s)
	
	// 检查排除关键词
	for _, keyword := range d.excludeKeywords {
		if strings.Contains(s, keyword) {
			return false
		}
	}
	
	// 检查匹配关键词
	for _, keyword := range d.matchKeywords {
		if strings.Contains(s, keyword) {
			return true
		}
	}
	
	return false
}

// searchInDirectory 在目录中搜索Java可执行文件
func (d *Detector) searchInDirectory(dir string, fullSearch bool) []*Java {
	var javaList []*Java
	
	if !fullSearch {
		return javaList
	}
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续搜索
		}
		
		// 检查是否是Java可执行文件
		if d.isJavaExecutable(path) {
			// 检查路径是否匹配关键词
			if d.findStr(filepath.Dir(path)) {
				version := d.GetJavaVersion(path)
				if version != "" {
					java := &Java{
						Path:    path,
						Version: version,
					}
					
					// 检查是否已存在
					exists := false
					for _, existing := range javaList {
						if existing.Equal(java) {
							exists = true
							break
						}
					}
					
					if !exists {
						javaList = append(javaList, java)
					}
				}
			}
		}
		
		return nil
	})
	
	if err != nil {
		// 记录错误但继续执行
		fmt.Printf("搜索目录 %s 时出错: %v\n", dir, err)
	}
	
	return javaList
}

// isJavaExecutable 检查文件是否是Java可执行文件
func (d *Detector) isJavaExecutable(path string) bool {
	if runtime.GOOS == "windows" {
		return strings.HasSuffix(path, "bin\\java.exe") || strings.HasSuffix(path, "bin/java.exe")
	}
	return strings.HasSuffix(path, "bin/java")
}

// DetectJava 检测系统中的Java安装
func (d *Detector) DetectJava(fullSearch bool) ([]*Java, error) {
	d.foundJava = make([]*Java, 0)

	switch runtime.GOOS {
	case "windows":
		return d.detectWindowsJava(fullSearch)
	case "darwin":
		return d.detectMacOSJava(fullSearch)
	case "linux":
		return d.detectLinuxJava(fullSearch)
	default:
		return d.detectLinuxJava(fullSearch) // 默认使用Linux逻辑
	}
}

// detectWindowsJava 检测Windows系统中的Java
func (d *Detector) detectWindowsJava(fullSearch bool) ([]*Java, error) {
	var javaList []*Java
	
	// 搜索所有驱动器
	for drive := 'C'; drive <= 'Z'; drive++ {
		drivePath := fmt.Sprintf("%c:\\", drive)
		if _, err := os.Stat(drivePath); err == nil {
			found := d.searchInDirectory(drivePath, fullSearch)
			javaList = append(javaList, found...)
		}
	}
	
	return javaList, nil
}

// detectMacOSJava 检测macOS系统中的Java
func (d *Detector) detectMacOSJava(fullSearch bool) ([]*Java, error) {
	var javaList []*Java
	
	// 检查环境变量PATH
	javaList = append(javaList, d.checkPathEnvironment()...)
	
	// 检查macOS特定路径
	macOSPaths := []string{
		"/Applications/Xcode.app/Contents/Applications/Application Loader.app/Contents/MacOS/itms/java",
		"/Library/Internet Plug-Ins/JavaAppletPlugin.plugin/Contents/Home/bin/java",
		"/System/Library/Frameworks/JavaVM.framework/Versions/Current/Commands/java",
	}
	
	for _, path := range macOSPaths {
		if _, err := os.Stat(path); err == nil {
			version := d.GetJavaVersion(path)
			if version != "" {
				java := &Java{Path: path, Version: version}
				if !d.containsJava(javaList, java) {
					javaList = append(javaList, java)
				}
			}
		}
	}
	
	// 检查默认安装路径
	basePath := "/Library/Java/JavaVirtualMachines/"
	if entries, err := os.ReadDir(basePath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				javaPath := filepath.Join(basePath, entry.Name(), "Contents/Home/bin/java")
				if _, err := os.Stat(javaPath); err == nil {
					version := d.GetJavaVersion(javaPath)
					if version != "" {
						java := &Java{Path: javaPath, Version: version}
						if !d.containsJava(javaList, java) {
							javaList = append(javaList, java)
						}
					}
				}
			}
		}
	}
	
	return javaList, nil
}

// detectLinuxJava 检测Linux系统中的Java
func (d *Detector) detectLinuxJava(fullSearch bool) ([]*Java, error) {
	var javaList []*Java
	
	// 检查环境变量PATH
	javaList = append(javaList, d.checkPathEnvironment()...)
	
	// 检查Linux默认安装路径
	linuxPaths := []string{
		"/usr",
		"/usr/java",
		"/usr/lib/jvm",
		"/usr/lib64/jvm",
		"/opt/jdk",
		"/opt/jdks",
	}
	
	for _, basePath := range linuxPaths {
		// 检查直接路径
		javaPath := filepath.Join(basePath, "bin/java")
		if _, err := os.Stat(javaPath); err == nil {
			version := d.GetJavaVersion(javaPath)
			if version != "" {
				java := &Java{Path: javaPath, Version: version}
				if !d.containsJava(javaList, java) {
					javaList = append(javaList, java)
				}
			}
		}
		
		// 检查JRE路径
		jrePath := filepath.Join(basePath, "jre/bin/java")
		if _, err := os.Stat(jrePath); err == nil {
			version := d.GetJavaVersion(jrePath)
			if version != "" {
				java := &Java{Path: jrePath, Version: version}
				if !d.containsJava(javaList, java) {
					javaList = append(javaList, java)
				}
			}
		}
		
		// 遍历子目录
		if entries, err := os.ReadDir(basePath); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					subPath := filepath.Join(basePath, entry.Name())
					
					// 检查子目录的bin/java
					subJavaPath := filepath.Join(subPath, "bin/java")
					if _, err := os.Stat(subJavaPath); err == nil {
						version := d.GetJavaVersion(subJavaPath)
						if version != "" {
							java := &Java{Path: subJavaPath, Version: version}
							if !d.containsJava(javaList, java) {
								javaList = append(javaList, java)
							}
						}
					}
					
					// 检查子目录的jre/bin/java
					subJrePath := filepath.Join(subPath, "jre/bin/java")
					if _, err := os.Stat(subJrePath); err == nil {
						version := d.GetJavaVersion(subJrePath)
						if version != "" {
							java := &Java{Path: subJrePath, Version: version}
							if !d.containsJava(javaList, java) {
								javaList = append(javaList, java)
							}
						}
					}
				}
			}
		}
	}
	
	return javaList, nil
}

// checkPathEnvironment 检查环境变量PATH中的Java
func (d *Detector) checkPathEnvironment() []*Java {
	var javaList []*Java

	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return javaList
	}

	separator := ":"
	if runtime.GOOS == "windows" {
		separator = ";"
	}

	paths := strings.Split(pathEnv, separator)
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path == "" {
			continue
		}

		var javaPath string
		if runtime.GOOS == "windows" {
			javaPath = filepath.Join(path, "java.exe")
		} else {
			javaPath = filepath.Join(path, "java")
		}

		if _, err := os.Stat(javaPath); err == nil {
			version := d.GetJavaVersion(javaPath)
			if version != "" {
				java := &Java{Path: javaPath, Version: version}
				if !d.containsJava(javaList, java) {
					javaList = append(javaList, java)
				}
			}
		}
	}

	return javaList
}

// containsJava 检查Java列表中是否包含指定的Java
func (d *Detector) containsJava(javaList []*Java, target *Java) bool {
	for _, java := range javaList {
		if java.Equal(target) {
			return true
		}
	}
	return false
}

// CheckJavaAvailability 检查Java是否可用
func (d *Detector) CheckJavaAvailability(java *Java) bool {
	if _, err := os.Stat(java.Path); err != nil {
		return false
	}

	version := d.GetJavaVersion(java.Path)
	return version == java.Version
}

// SaveJavaList 保存Java列表到文件
func SaveJavaList(javaList []*Java, filePath string) error {
	data := map[string]interface{}{
		"java": javaList,
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("序列化Java列表失败: %w", err)
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	return nil
}

// LoadJavaList 从文件加载Java列表
func LoadJavaList(filePath string) ([]*Java, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return []*Java{}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	var result map[string][]*Java
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}

	javaList, exists := result["java"]
	if !exists {
		return []*Java{}, nil
	}

	return javaList, nil
}

// SortJavaList 对Java列表按版本排序
func SortJavaList(javaList []*Java, reverse bool) {
	// 简单的字符串排序，可以根据需要改进为版本号排序
	for i := 0; i < len(javaList)-1; i++ {
		for j := i + 1; j < len(javaList); j++ {
			var shouldSwap bool
			if reverse {
				shouldSwap = javaList[i].Version < javaList[j].Version
			} else {
				shouldSwap = javaList[i].Version > javaList[j].Version
			}

			if shouldSwap {
				javaList[i], javaList[j] = javaList[j], javaList[i]
			}
		}
	}
}
