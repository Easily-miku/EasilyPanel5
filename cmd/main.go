package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/manifoldco/promptui"

	"easilypanel/internal/config"
	"easilypanel/internal/download"
	"easilypanel/internal/frp"
	"easilypanel/internal/instance"
	"easilypanel/internal/java"
	"easilypanel/internal/menu"
)

func main() {
	// 定义命令行参数
	var (
		showVersion = flag.Bool("version", false, "显示版本信息")
		showHelp    = flag.Bool("help", false, "显示帮助信息")
		configFile  = flag.String("config", "", "指定配置文件路径")
		dataDir     = flag.String("data", "./data", "指定数据目录")
		logLevel    = flag.String("log", "info", "设置日志级别 (debug, info, warn, error)")
		daemon      = flag.Bool("daemon", false, "以守护进程模式运行")
	)

	flag.Parse()

	// 处理版本信息
	if *showVersion {
		fmt.Println("EasilyPanel5 v1.0.0")
		fmt.Println("跨平台通用游戏服务器管理工具")
		fmt.Println("构建时间:", time.Now().Format("2006-01-02 15:04:05"))
		return
	}

	// 处理帮助信息
	if *showHelp {
		showHelpInfo()
		return
	}

	// 处理命令行子命令
	args := flag.Args()
	if len(args) > 0 {
		handleCommandLine(args, *configFile, *dataDir, *logLevel)
		return
	}

	// 启动交互式菜单
	runInteractiveMenu(*configFile, *dataDir, *logLevel, *daemon)
}

// showHelpInfo 显示帮助信息
func showHelpInfo() {
	fmt.Println("EasilyPanel5 v1.0.0 - 跨平台通用游戏服务器管理工具")
	fmt.Println()
	fmt.Println("用法:")
	fmt.Println("  easilypanel [选项] [命令]")
	fmt.Println()
	fmt.Println("选项:")
	fmt.Println("  -version        显示版本信息")
	fmt.Println("  -help           显示帮助信息")
	fmt.Println("  -config FILE    指定配置文件路径")
	fmt.Println("  -data DIR       指定数据目录 (默认: ./data)")
	fmt.Println("  -log LEVEL      设置日志级别 (debug, info, warn, error)")
	fmt.Println("  -daemon         以守护进程模式运行")
	fmt.Println()
	fmt.Println("命令:")
	fmt.Println("  instance        实例管理")
	fmt.Println("    list          列出所有实例")
	fmt.Println("    start NAME    启动指定实例")
	fmt.Println("    stop NAME     停止指定实例")
	fmt.Println("    status NAME   查看实例状态")
	fmt.Println()
	fmt.Println("  frp             内网穿透管理")
	fmt.Println("    status        查看frpc状态")
	fmt.Println("    start         启动frpc")
	fmt.Println("    stop          停止frpc")
	fmt.Println("    restart       重启frpc")
	fmt.Println()
	fmt.Println("  java            Java环境管理")
	fmt.Println("    detect        检测Java版本")
	fmt.Println("    list          列出Java版本")
	fmt.Println()
	fmt.Println("示例:")
	fmt.Println("  easilypanel                    # 启动交互式界面")
	fmt.Println("  easilypanel -version           # 显示版本信息")
	fmt.Println("  easilypanel instance list      # 列出所有实例")
	fmt.Println("  easilypanel frp status         # 查看frpc状态")
	fmt.Println("  easilypanel java detect        # 检测Java版本")
}

// runInteractiveMenu 运行交互式菜单
func runInteractiveMenu(configFile, dataDir, logLevel string, daemon bool) {
	// 显示启动信息
	if !daemon {
		fmt.Println("正在启动 EasilyPanel5 v1.0.0...")
		fmt.Println("欢迎使用 跨平台通用游戏服务器管理工具！")
		fmt.Printf("数据目录: %s\n", dataDir)
		fmt.Printf("日志级别: %s\n", logLevel)
		fmt.Println()
	}

	// 初始化配置
	configManager := config.NewManager(configFile)
	if err := configManager.Initialize(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		return
	}

	// 设置数据目录
	if dataDir != "./data" {
		config.Set("app.data_dir", dataDir)
	}

	// 设置日志级别
	if logLevel != "info" {
		config.Set("app.log_level", logLevel)
	}

	// 守护进程模式
	if daemon {
		fmt.Println("守护进程模式启动...")
		// 这里可以添加守护进程逻辑
		return
	}

	// 创建菜单系统
	menuSystem := menu.NewMenuSystem()

	// 创建主菜单
	mainMenu := createMainMenu()

	// 设置根菜单并运行
	menuSystem.SetRootMenu(mainMenu)
	menuSystem.Run()
}

// createMainMenu 创建主菜单
func createMainMenu() *menu.Menu {
	mainMenu := menu.NewMenu("EasilyPanel5 v1.0.0", "跨平台通用游戏服务器管理工具")
	
	// 实例管理
	instanceMenu := createInstanceMenu()
	mainMenu.AddItem(
		menu.NewMenuItem("instance", "实例管理", "创建、管理和监控游戏服务器实例").
			WithSubMenu(instanceMenu).
			WithStatus(func() string {
				// 显示实例数量
				manager := instance.NewManager("./data/instances")
				instances, err := manager.ListInstances()
				if err != nil {
					return "错误"
				}
				return fmt.Sprintf("%d个实例", len(instances))
			}),
	)
	
	// 服务端下载
	downloadMenu := createDownloadMenu()
	mainMenu.AddItem(
		menu.NewMenuItem("download", "服务端下载", "下载各种游戏服务端文件").
			WithSubMenu(downloadMenu),
	)

	// Java环境
	javaMenu := createJavaMenu()
	mainMenu.AddItem(
		menu.NewMenuItem("java", "Java环境", "检测和管理Java运行环境").
			WithSubMenu(javaMenu).
			WithStatus(func() string {
				// 显示Java状态
				detector := java.NewDetector()
				versions, _ := detector.DetectJava(false)
				if len(versions) == 0 {
					return "未检测到"
				}
				return fmt.Sprintf("%d个版本", len(versions))
			}),
	)

	// 内网穿透
	frpMenu := createFRPMenu()
	mainMenu.AddItem(
		menu.NewMenuItem("frp", "内网穿透", "OpenFRP内网穿透服务配置").
			WithSubMenu(frpMenu).
			WithStatus(func() string {
				// 显示FRP状态
				manager := frp.NewManager("./data")
				if manager.IsFRPCRunning() {
					return "运行中"
				}
				return "已停止"
			}),
	)

	// 系统设置
	settingsMenu := createSettingsMenu()
	mainMenu.AddItem(
		menu.NewMenuItem("settings", "系统设置", "配置管理、备份等系统功能").
			WithSubMenu(settingsMenu),
	)
	
	return mainMenu
}

// createInstanceMenu 创建实例管理菜单
func createInstanceMenu() *menu.Menu {
	instanceMenu := menu.NewMenu("实例管理", "管理游戏服务器实例")
	
	instanceMenu.AddItems(
		menu.NewMenuItem("create", "创建实例", "创建新的游戏服务器实例").
			WithHandler(func() error {
				return handleCreateInstance()
			}),

		menu.NewMenuItem("list", "实例列表", "查看所有实例的状态和信息").
			WithHandler(func() error {
				return handleInstanceList()
			}),

		menu.NewMenuItem("manage", "管理实例", "启动、停止、删除实例").
			WithHandler(func() error {
				return handleManageInstance()
			}),

		menu.NewMenuItem("monitor", "实例监控", "监控实例运行状态和性能").
			WithHandler(func() error {
				fmt.Println("实例监控功能正在开发中...")
				return nil
			}).
			WithEnabled(func() bool { return false }),
	)
	
	return instanceMenu
}

// createDownloadMenu 创建下载菜单
func createDownloadMenu() *menu.Menu {
	downloadMenu := menu.NewMenu("服务端下载", "下载各种游戏服务端")
	
	downloadMenu.AddItems(
		menu.NewMenuItem("fastmirror", "FastMirror下载", "从FastMirror下载服务端").
			WithHandler(func() error {
				return handleFastMirrorDownload()
			}),

		menu.NewMenuItem("files", "已下载文件", "查看和管理已下载的文件").
			WithHandler(func() error {
				return handleDownloadedFiles()
			}),

		menu.NewMenuItem("cleanup", "清理下载", "清理临时文件和无用下载").
			WithHandler(func() error {
				return handleCleanupDownloads()
			}),
	)
	
	return downloadMenu
}

// createJavaMenu 创建Java菜单
func createJavaMenu() *menu.Menu {
	javaMenu := menu.NewMenu("Java环境", "管理Java运行环境")
	
	javaMenu.AddItems(
		menu.NewMenuItem("detect", "检测Java", "自动检测系统中的Java版本").
			WithHandler(func() error {
				return handleJavaDetect()
			}),

		menu.NewMenuItem("list", "Java列表", "查看所有检测到的Java版本").
			WithHandler(func() error {
				return handleJavaList()
			}),

		menu.NewMenuItem("add", "手动添加Java", "手动添加Java环境路径").
			WithHandler(func() error {
				return handleJavaAdd()
			}),

		menu.NewMenuItem("install", "安装Java", "下载并安装Java运行环境").
			WithHandler(func() error {
				fmt.Println("Java安装功能正在开发中...")
				return nil
			}).
			WithEnabled(func() bool { return false }),
	)
	
	return javaMenu
}

// createFRPMenu 创建FRP菜单
func createFRPMenu() *menu.Menu {
	frpMenu := menu.NewMenu("内网穿透", "OpenFRP内网穿透服务")
	
	frpMenu.AddItems(
		menu.NewMenuItem("setup", "配置OpenFRP", "设置OpenFRP认证和客户端").
			WithHandler(func() error {
				return handleFRPSetup()
			}),

		menu.NewMenuItem("tunnels", "管理隧道", "创建、编辑、删除隧道").
			WithHandler(func() error {
				return handleFRPTunnels()
			}),

		menu.NewMenuItem("client", "frpc客户端", "管理frpc客户端进程").
			WithHandler(func() error {
				return handleFRPClient()
			}),

		menu.NewMenuItem("status", "状态监控", "查看隧道和客户端状态").
			WithHandler(func() error {
				return handleFRPStatus()
			}),
	)
	
	return frpMenu
}

// createSettingsMenu 创建设置菜单
func createSettingsMenu() *menu.Menu {
	settingsMenu := menu.NewMenu("系统设置", "配置和系统管理")
	
	settingsMenu.AddItems(
		menu.NewMenuItem("config", "配置管理", "查看和修改系统配置").
			WithSubMenu(createConfigMenu()),

		menu.NewMenuItem("backup", "备份管理", "创建和恢复系统备份").
			WithHandler(func() error {
				fmt.Println("备份管理功能正在开发中...")
				return nil
			}).
			WithEnabled(func() bool { return false }),

		menu.NewMenuItem("logs", "日志查看", "查看系统运行日志").
			WithHandler(func() error {
				fmt.Println("日志查看功能正在开发中...")
				return nil
			}).
			WithEnabled(func() bool { return false }),

		menu.NewMenuItem("about", "关于程序", "查看版本信息和帮助").
			WithHandler(func() error {
				return handleAbout()
			}),
	)
	
	return settingsMenu
}

// 处理函数占位符
func handleCreateInstance() error {
	fmt.Println("=== 创建实例 ===")

	scanner := bufio.NewScanner(os.Stdin)

	// 输入实例名称
	fmt.Print("请输入实例名称: ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}
	instanceName := strings.TrimSpace(scanner.Text())
	if instanceName == "" {
		return fmt.Errorf("实例名称不能为空")
	}

	// 选择服务端类型
	serverTypes := []string{"Minecraft Java版", "Minecraft 基岩版"}

	prompt := promptui.Select{
		Label: "请选择服务端类型",
		Items: serverTypes,
	}

	index, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("选择服务端类型失败: %w", err)
	}

	var serverType string
	switch index {
	case 0:
		serverType = "minecraft"
	case 1:
		serverType = "bedrock"
	default:
		return fmt.Errorf("无效的服务端类型")
	}

	// 输入端口
	fmt.Print("请输入服务器端口 (默认: 25565): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	port := strings.TrimSpace(scanner.Text())
	if port == "" {
		port = "25565"
	}

	// 创建实例
	manager := instance.NewManager("./data/instances")

	fmt.Printf("\n正在创建实例 '%s'...\n", instanceName)

	if serverType == "minecraft" {
		// 检测Java路径
		detector := java.NewDetector()
		javaVersions, _ := detector.DetectJava(false)
		javaPath := "java"
		if len(javaVersions) > 0 {
			javaPath = javaVersions[0].Path
		}

		_, err := manager.CreateMinecraftInstance(instanceName, "latest", "vanilla", javaPath)
		if err != nil {
			return fmt.Errorf("创建实例失败: %w", err)
		}
	} else {
		_, err := manager.CreateBlankInstance(instanceName, "基岩版服务器", "")
		if err != nil {
			return fmt.Errorf("创建实例失败: %w", err)
		}
	}

	fmt.Printf("✓ 实例 '%s' 创建成功\n", instanceName)
	fmt.Printf("类型: %s\n", serverType)
	fmt.Printf("端口: %s\n", port)

	return nil
}

func handleInstanceList() error {
	fmt.Println("=== 实例列表 ===")
	manager := instance.NewManager("./data/instances")
	instances, err := manager.ListInstances()
	if err != nil {
		return err
	}

	if len(instances) == 0 {
		fmt.Println("暂无实例")
		return nil
	}

	for _, inst := range instances {
		fmt.Printf("- %s (%s)\n", inst.Name, inst.Type)
	}
	return nil
}

func handleManageInstance() error {
	fmt.Println("=== 管理实例 ===")

	manager := instance.NewManager("./data/instances")
	instances, err := manager.ListInstances()
	if err != nil {
		return fmt.Errorf("获取实例列表失败: %w", err)
	}

	if len(instances) == 0 {
		fmt.Println("暂无实例，请先创建实例")
		return nil
	}

	// 显示实例列表
	fmt.Println("现有实例:")
	for i, inst := range instances {
		status := inst.Status
		if status == "" {
			status = "未知"
		}
		fmt.Printf("%d. %s (%s) - %s\n", i+1, inst.Name, inst.Type, status)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\n请选择要管理的实例 (输入序号): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())
	instanceIndex := -1
	for i := range instances {
		if fmt.Sprintf("%d", i+1) == choice {
			instanceIndex = i
			break
		}
	}

	if instanceIndex == -1 {
		return fmt.Errorf("无效的实例选择")
	}

	selectedInstance := instances[instanceIndex]

	// 显示管理选项
	fmt.Printf("\n管理实例: %s\n", selectedInstance.Name)

	actions := []string{
		"启动实例",
		"停止实例",
		"重启实例",
		"删除实例",
		"查看配置",
		"编辑配置",
		"查看日志",
	}

	prompt := promptui.Select{
		Label: "请选择操作",
		Items: actions,
	}

	actionIndex, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("选择操作失败: %w", err)
	}

	// 创建进程管理器
	processManager := instance.NewProcessManager("./data/instances")

	switch actionIndex {
	case 0:
		fmt.Printf("正在启动实例 '%s'...\n", selectedInstance.Name)
		if err := processManager.StartInstance(selectedInstance.Name); err != nil {
			return fmt.Errorf("启动实例失败: %w", err)
		}
		fmt.Println("✓ 实例启动成功")

	case 1:
		fmt.Printf("正在停止实例 '%s'...\n", selectedInstance.Name)
		if err := processManager.StopInstance(selectedInstance.Name); err != nil {
			return fmt.Errorf("停止实例失败: %w", err)
		}
		fmt.Println("✓ 实例停止成功")

	case 2:
		fmt.Printf("正在重启实例 '%s'...\n", selectedInstance.Name)
		if err := processManager.RestartInstance(selectedInstance.Name); err != nil {
			return fmt.Errorf("重启实例失败: %w", err)
		}
		fmt.Println("✓ 实例重启成功")

	case 3:
		fmt.Printf("确定要删除实例 '%s' 吗? (y/N): ", selectedInstance.Name)
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if confirm == "y" || confirm == "yes" {
			if err := manager.DeleteInstance(selectedInstance.Name, true); err != nil {
				return fmt.Errorf("删除实例失败: %w", err)
			}
			fmt.Printf("✓ 实例 '%s' 已删除\n", selectedInstance.Name)
		} else {
			fmt.Println("取消删除")
		}

	case 4:
		fmt.Printf("\n实例配置: %s\n", selectedInstance.Name)
		fmt.Printf("类型: %s\n", selectedInstance.Type)
		fmt.Printf("端口: %d\n", selectedInstance.Port)
		fmt.Printf("状态: %s\n", selectedInstance.Status)
		fmt.Printf("工作目录: %s\n", selectedInstance.WorkDir)
		if selectedInstance.ServerJar != "" {
			fmt.Printf("服务端文件: %s\n", selectedInstance.ServerJar)
		}
		if selectedInstance.JavaPath != "" {
			fmt.Printf("Java路径: %s\n", selectedInstance.JavaPath)
		}
		if selectedInstance.MaxMemory != "" {
			fmt.Printf("最大内存: %s\n", selectedInstance.MaxMemory)
		}
		fmt.Printf("创建时间: %s\n", selectedInstance.CreatedAt.Format("2006-01-02 15:04:05"))
		if selectedInstance.LastStarted != nil {
			fmt.Printf("最后启动: %s\n", selectedInstance.LastStarted.Format("2006-01-02 15:04:05"))
		}

	case 5:
		return handleEditInstanceConfig(manager, selectedInstance, scanner)

	case 6:
		return handleViewInstanceLogs(selectedInstance)

	default:
		return fmt.Errorf("无效的操作选择")
	}

	return nil
}

func handleFastMirrorDownload() error {
	fmt.Println("=== FastMirror下载 ===")
	dm := download.NewDownloadManager("./data")

	// 获取服务端列表
	servers, err := dm.ListAvailableServers()
	if err != nil {
		return fmt.Errorf("获取服务端列表失败: %w", err)
	}

	if len(servers) == 0 {
		fmt.Println("未找到可用的服务端")
		return nil
	}

	// 显示服务端列表
	fmt.Printf("可用服务端 (%d 个):\n", len(servers))
	for i, server := range servers {
		fmt.Printf("%d. %s\n", i+1, server.Name)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("\n请选择要下载的服务端 (输入序号): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())
	serverIndex := -1
	for i := range servers {
		if fmt.Sprintf("%d", i+1) == choice {
			serverIndex = i
			break
		}
	}

	if serverIndex == -1 {
		return fmt.Errorf("无效的服务端选择")
	}

	selectedServer := servers[serverIndex]
	fmt.Printf("选择的服务端: %s\n", selectedServer.Name)

	// 获取版本列表
	fmt.Println("正在获取版本列表...")
	versions, err := dm.ListVersions(selectedServer.Name)
	if err != nil {
		return fmt.Errorf("获取版本列表失败: %w", err)
	}

	if len(versions) == 0 {
		fmt.Println("未找到可用版本")
		return nil
	}

	// 显示版本列表（只显示前10个）
	fmt.Printf("\n可用版本 (%d 个):\n", len(versions))
	displayCount := len(versions)
	if displayCount > 10 {
		displayCount = 10
	}

	for i := 0; i < displayCount; i++ {
		fmt.Printf("%d. %s\n", i+1, versions[i])
	}

	if len(versions) > 10 {
		fmt.Printf("... 还有 %d 个版本\n", len(versions)-10)
	}

	fmt.Print("\n请选择版本 (输入序号或直接输入版本号): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	versionChoice := strings.TrimSpace(scanner.Text())
	var selectedVersion string

	// 尝试解析为序号
	for i := 0; i < displayCount; i++ {
		if fmt.Sprintf("%d", i+1) == versionChoice {
			selectedVersion = versions[i]
			break
		}
	}

	// 如果不是序号，直接使用输入的版本号
	if selectedVersion == "" {
		selectedVersion = versionChoice
	}

	fmt.Printf("选择的版本: %s\n", selectedVersion)

	// 询问是否下载最新构建
	fmt.Print("是否下载最新构建? (Y/n): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	downloadLatest := true
	latestChoice := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if latestChoice == "n" || latestChoice == "no" {
		downloadLatest = false
	}

	// 开始下载
	fmt.Printf("\n开始下载 %s %s...\n", selectedServer.Name, selectedVersion)

	var filePath string
	if downloadLatest {
		// 获取最新构建
		latestBuild, err := dm.GetLatestBuild(selectedServer.Name, selectedVersion)
		if err != nil {
			return fmt.Errorf("获取最新构建失败: %w", err)
		}
		filePath, err = dm.DownloadServer(selectedServer.Name, selectedVersion, latestBuild.CoreVersion, true)
	} else {
		// 获取构建列表
		builds, err := dm.ListBuilds(selectedServer.Name, selectedVersion, 10)
		if err != nil {
			return fmt.Errorf("获取构建列表失败: %w", err)
		}

		if len(builds) == 0 {
			return fmt.Errorf("未找到可用构建")
		}

		// 显示构建列表
		fmt.Printf("可用构建 (%d 个):\n", len(builds))
		for i, build := range builds {
			fmt.Printf("%d. %s (更新时间: %s)\n", i+1, build.CoreVersion, build.UpdateTime)
		}

		fmt.Print("请选择构建 (输入序号): ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		buildChoice := strings.TrimSpace(scanner.Text())
		buildIndex := -1
		for i := range builds {
			if fmt.Sprintf("%d", i+1) == buildChoice {
				buildIndex = i
				break
			}
		}

		if buildIndex == -1 {
			return fmt.Errorf("无效的构建选择")
		}

		selectedBuild := builds[buildIndex]
		filePath, err = dm.DownloadServer(selectedServer.Name, selectedVersion, selectedBuild.CoreVersion, true)
	}

	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	fmt.Printf("\n✓ 下载完成!\n")
	fmt.Printf("文件路径: %s\n", filePath)

	// 询问是否创建实例
	fmt.Print("\n是否使用此文件创建Minecraft实例? (y/N): ")
	if !scanner.Scan() {
		return nil
	}

	createInstance := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if createInstance == "y" || createInstance == "yes" {
		return handleCreateInstanceFromDownload(filePath, selectedServer.Name, selectedVersion)
	}

	return nil
}

func handleDownloadedFiles() error {
	fmt.Println("=== 已下载文件 ===")
	dm := download.NewDownloadManager("./data")

	// 获取下载目录
	downloadDir := dm.GetDownloadDir()

	// 检查目录是否存在
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		fmt.Println("下载目录不存在，暂无已下载文件")
		return nil
	}

	// 读取目录内容
	files, err := os.ReadDir(downloadDir)
	if err != nil {
		return fmt.Errorf("读取下载目录失败: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("暂无已下载文件")
		return nil
	}

	fmt.Printf("下载目录: %s\n", downloadDir)
	fmt.Printf("已下载文件 (%d 个):\n\n", len(files))

	for i, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// 格式化文件大小
		size := info.Size()
		var sizeStr string
		if size < 1024 {
			sizeStr = fmt.Sprintf("%d B", size)
		} else if size < 1024*1024 {
			sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
		} else if size < 1024*1024*1024 {
			sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
		} else {
			sizeStr = fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
		}

		fmt.Printf("%d. %s\n", i+1, file.Name())
		fmt.Printf("   大小: %s\n", sizeStr)
		fmt.Printf("   修改时间: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
		fmt.Println()
	}

	return nil
}

func handleCleanupDownloads() error {
	fmt.Println("=== 清理下载 ===")

	dm := download.NewDownloadManager("./data")
	downloadDir := dm.GetDownloadDir()

	// 检查目录是否存在
	if _, err := os.Stat(downloadDir); os.IsNotExist(err) {
		fmt.Println("下载目录不存在，无需清理")
		return nil
	}

	// 读取目录内容
	files, err := os.ReadDir(downloadDir)
	if err != nil {
		return fmt.Errorf("读取下载目录失败: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("下载目录为空，无需清理")
		return nil
	}

	// 计算总大小
	var totalSize int64
	var fileCount int
	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err == nil {
				totalSize += info.Size()
				fileCount++
			}
		}
	}

	if fileCount == 0 {
		fmt.Println("没有找到可清理的文件")
		return nil
	}

	// 格式化总大小
	var totalSizeStr string
	if totalSize < 1024*1024 {
		totalSizeStr = fmt.Sprintf("%.1f KB", float64(totalSize)/1024)
	} else if totalSize < 1024*1024*1024 {
		totalSizeStr = fmt.Sprintf("%.1f MB", float64(totalSize)/(1024*1024))
	} else {
		totalSizeStr = fmt.Sprintf("%.1f GB", float64(totalSize)/(1024*1024*1024))
	}

	fmt.Printf("找到 %d 个文件，总大小: %s\n", fileCount, totalSizeStr)

	cleanupOptions := []string{
		"清理所有下载文件",
		"清理7天前的文件",
		"清理30天前的文件",
		"取消清理",
	}

	prompt := promptui.Select{
		Label: "请选择清理方式",
		Items: cleanupOptions,
	}

	cleanupIndex, _, err := prompt.Run()
	if err != nil {
		return fmt.Errorf("选择清理方式失败: %w", err)
	}

	switch cleanupIndex {
	case 0:
		confirmPrompt := promptui.Prompt{
			Label:     "确定要删除所有下载文件吗",
			IsConfirm: true,
		}
		_, err := confirmPrompt.Run()
		if err == nil {
			return cleanupFiles(downloadDir, 0)
		}
		fmt.Println("取消清理")

	case 1:
		return cleanupFiles(downloadDir, 7)

	case 2:
		return cleanupFiles(downloadDir, 30)

	case 3:
		fmt.Println("取消清理")

	default:
		fmt.Println("无效选择")
	}

	return nil
}

func cleanupFiles(dir string, days int) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	now := time.Now()
	var deletedCount int
	var deletedSize int64

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// 检查文件年龄
		if days > 0 {
			age := now.Sub(info.ModTime())
			if age.Hours() < float64(days*24) {
				continue
			}
		}

		filePath := fmt.Sprintf("%s/%s", dir, file.Name())
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("删除文件失败: %s - %v\n", file.Name(), err)
			continue
		}

		deletedCount++
		deletedSize += info.Size()
		fmt.Printf("已删除: %s\n", file.Name())
	}

	// 格式化删除的大小
	var deletedSizeStr string
	if deletedSize < 1024*1024 {
		deletedSizeStr = fmt.Sprintf("%.1f KB", float64(deletedSize)/1024)
	} else if deletedSize < 1024*1024*1024 {
		deletedSizeStr = fmt.Sprintf("%.1f MB", float64(deletedSize)/(1024*1024))
	} else {
		deletedSizeStr = fmt.Sprintf("%.1f GB", float64(deletedSize)/(1024*1024*1024))
	}

	fmt.Printf("\n✓ 清理完成: 删除了 %d 个文件，释放空间 %s\n", deletedCount, deletedSizeStr)
	return nil
}

func handleJavaDetect() error {
	fmt.Println("=== 检测Java ===")
	detector := java.NewDetector()
	versions, err := detector.DetectJava(true)
	if err != nil {
		return err
	}

	if len(versions) == 0 {
		fmt.Println("未检测到Java环境")
		return nil
	}

	fmt.Printf("检测到 %d 个Java版本:\n", len(versions))
	for _, version := range versions {
		fmt.Printf("- Java %d (%s)\n", version.Version, version.Path)
	}
	return nil
}

func handleJavaList() error {
	return handleJavaDetect()
}

func handleJavaAdd() error {
	fmt.Println("=== 手动添加Java ===")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("请输入JDK的bin文件夹路径")
	fmt.Println("示例:")
	fmt.Println("  Linux/macOS: /usr/lib/jvm/java-17-openjdk/bin")
	fmt.Println("  Windows: C:\\Program Files\\Java\\jdk-17\\bin")
	fmt.Println()

	fmt.Print("JDK bin路径: ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	binPath := strings.TrimSpace(scanner.Text())
	if binPath == "" {
		fmt.Println("操作已取消")
		return nil
	}

	// 验证路径是否存在
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("路径不存在: %s", binPath)
	}

	// 构建java可执行文件路径
	var javaPath string
	if runtime.GOOS == "windows" {
		javaPath = filepath.Join(binPath, "java.exe")
	} else {
		javaPath = filepath.Join(binPath, "java")
	}

	// 验证java可执行文件是否存在
	if _, err := os.Stat(javaPath); os.IsNotExist(err) {
		return fmt.Errorf("在指定路径中未找到java可执行文件: %s", javaPath)
	}

	fmt.Printf("找到Java可执行文件: %s\n", javaPath)
	fmt.Println("正在验证Java版本...")

	// 创建Java管理器
	manager := java.NewManager("./data/configs")

	// 尝试添加Java
	addedJava, err := manager.AddJava(javaPath)
	if err != nil {
		if strings.Contains(err.Error(), "Java已存在") {
			fmt.Printf("✓ Java已存在于列表中: %s (版本 %s)\n", addedJava.Path, addedJava.Version)
			return nil
		}
		return fmt.Errorf("添加Java失败: %w", err)
	}

	fmt.Printf("✓ Java添加成功!\n")
	fmt.Printf("  路径: %s\n", addedJava.Path)
	fmt.Printf("  版本: %s\n", addedJava.Version)

	// 显示当前Java列表
	fmt.Println("\n当前Java列表:")
	manager.PrintJavaList()

	return nil
}

func handleFRPSetup() error {
	fmt.Println("=== 配置OpenFRP ===")

	manager := frp.NewManager("./data")

	// 检查frpc是否已安装
	fmt.Println("正在检查frpc客户端...")
	if err := manager.SetupFRPC(); err != nil {
		return fmt.Errorf("设置frpc失败: %w", err)
	}
	fmt.Println("✓ frpc客户端检查完成")

	// 检查认证令牌
	token := config.GetString("frp.openfrp.authorization")
	if token == "" {
		fmt.Println("\n请设置OpenFRP认证令牌:")
		fmt.Println("1. 访问 https://openfrp.net")
		fmt.Println("2. 登录账户")
		fmt.Println("3. 在个人中心获取Authorization令牌")

		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("\n请输入认证令牌: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		token = strings.TrimSpace(scanner.Text())
		if token == "" {
			return fmt.Errorf("认证令牌不能为空")
		}

		manager.SetAuthorization(token)

		// 测试连接
		fmt.Println("正在验证令牌...")
		if err := manager.TestConnection(); err != nil {
			return fmt.Errorf("认证失败: %w", err)
		}

		// 保存到配置
		config.Set("frp.openfrp.authorization", token)
		if err := config.SaveConfig(); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Println("✓ 认证令牌设置成功")
	} else {
		manager.SetAuthorization(token)
		fmt.Println("✓ 已配置认证令牌")

		// 测试连接
		if err := manager.TestConnection(); err != nil {
			fmt.Printf("警告: 令牌验证失败: %v\n", err)
		} else {
			fmt.Println("✓ 令牌验证成功")
		}
	}

	// 显示用户信息
	userInfo, err := manager.GetUserInfo()
	if err != nil {
		return fmt.Errorf("获取用户信息失败: %w", err)
	}

	fmt.Printf("\n用户信息:\n")
	fmt.Printf("用户名: %s\n", userInfo.Username)
	fmt.Printf("用户组: %s\n", userInfo.FriendlyGroup)
	fmt.Printf("隧道配额: %d/%d\n", userInfo.Used, userInfo.Proxies)
	fmt.Printf("剩余流量: %d MB\n", userInfo.Traffic)

	fmt.Println("\n✓ OpenFRP配置完成")
	return nil
}

func handleFRPTunnels() error {
	fmt.Println("=== 管理隧道 ===")

	manager := frp.NewManager("./data")

	// 检查认证
	token := config.GetString("frp.openfrp.authorization")
	if token == "" {
		return fmt.Errorf("请先配置OpenFRP认证令牌")
	}
	manager.SetAuthorization(token)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n隧道管理:")
		fmt.Println("1. 查看隧道列表")
		fmt.Println("2. 创建新隧道")
		fmt.Println("3. 删除隧道")
		fmt.Println("4. 返回上级菜单")
		fmt.Print("请选择操作 (1-4): ")

		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			if err := showTunnelList(manager); err != nil {
				fmt.Printf("错误: %v\n", err)
			}

		case "2":
			if err := createNewTunnel(manager, scanner); err != nil {
				fmt.Printf("错误: %v\n", err)
			}

		case "3":
			if err := deleteTunnel(manager, scanner); err != nil {
				fmt.Printf("错误: %v\n", err)
			}

		case "4":
			return nil

		default:
			fmt.Println("无效选择")
		}
	}
}

func handleFRPClient() error {
	fmt.Println("=== frpc客户端 ===")

	manager := frp.NewManager("./data")
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\nfrpc客户端管理:")
		fmt.Printf("当前状态: %s\n", manager.GetFRPCStatus())
		fmt.Println("\n操作选项:")
		fmt.Println("1. 启动frpc (配置文件方式)")
		fmt.Println("2. 启动frpc (命令行方式)")
		fmt.Println("3. 停止frpc")
		fmt.Println("4. 重启frpc")
		fmt.Println("5. 查看日志")
		fmt.Println("6. 清空日志")
		fmt.Println("7. 查看配置")
		fmt.Println("8. 重新生成配置")
		fmt.Println("0. 返回上级菜单")
		fmt.Print("请选择操作 (0-8): ")

		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Println("正在启动frpc (配置文件方式)...")
			if err := manager.StartFRPC(); err != nil {
				fmt.Printf("启动失败: %v\n", err)
			} else {
				fmt.Println("✓ frpc启动成功")
			}

		case "2":
			if err := handleStartFRPCWithCommand(manager, scanner); err != nil {
				fmt.Printf("启动失败: %v\n", err)
			}

		case "3":
			fmt.Println("正在停止frpc...")
			if err := manager.StopFRPC(); err != nil {
				fmt.Printf("停止失败: %v\n", err)
			} else {
				fmt.Println("✓ frpc停止成功")
			}

		case "4":
			fmt.Println("正在重启frpc...")
			if err := manager.RestartFRPC(); err != nil {
				fmt.Printf("重启失败: %v\n", err)
			} else {
				fmt.Println("✓ frpc重启成功")
			}

		case "5":
			if err := handleViewFRPCLogs(manager); err != nil {
				fmt.Printf("查看日志失败: %v\n", err)
			}

		case "6":
			fmt.Print("确定要清空frpc日志吗? (y/N): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if confirm == "y" || confirm == "yes" {
				if err := manager.ClearFRPCLogs(); err != nil {
					fmt.Printf("清空日志失败: %v\n", err)
				} else {
					fmt.Println("✓ 日志已清空")
				}
			}

		case "7":
			if err := handleViewFRPCConfig(manager); err != nil {
				fmt.Printf("查看配置失败: %v\n", err)
			}

		case "8":
			fmt.Print("确定要重新生成frpc配置吗? (y/N): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if confirm == "y" || confirm == "yes" {
				// 获取配置信息
				token := config.GetString("frp.openfrp.authorization")
				serverAddr := config.GetString("frp.openfrp.server_addr")
				if serverAddr == "" {
					serverAddr = "frp-app.top:7000"
				}

				if err := manager.GenerateConfig(serverAddr, token); err != nil {
					fmt.Printf("生成配置失败: %v\n", err)
				} else {
					fmt.Println("✓ 配置已重新生成")
				}
			}

		case "0":
			return nil

		default:
			fmt.Println("无效选择")
		}

		fmt.Print("\n按回车键继续...")
		scanner.Scan()
	}
}

// handleStartFRPCWithCommand 使用命令行方式启动frpc
func handleStartFRPCWithCommand(manager *frp.Manager, scanner *bufio.Scanner) error {
	fmt.Println("正在启动frpc (命令行方式)...")

	// 获取用户访问密钥
	userToken := config.GetString("frp.openfrp.user_token")
	if userToken == "" {
		fmt.Print("请输入用户访问密钥: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		userToken = strings.TrimSpace(scanner.Text())
		if userToken == "" {
			return fmt.Errorf("用户访问密钥不能为空")
		}

		// 保存到配置
		config.Set("frp.openfrp.user_token", userToken)
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("警告: 保存配置失败: %v\n", err)
		}
	}

	// 设置认证令牌以获取隧道列表
	authToken := config.GetString("frp.openfrp.authorization")
	if authToken == "" {
		return fmt.Errorf("未设置认证令牌，请先在设置中配置")
	}

	manager.SetAuthorization(authToken)

	// 获取隧道列表
	fmt.Println("正在获取隧道列表...")
	proxies, err := manager.GetProxies()
	if err != nil {
		return fmt.Errorf("获取隧道列表失败: %w", err)
	}

	if len(proxies) == 0 {
		return fmt.Errorf("没有可用的隧道")
	}

	// 显示隧道列表供用户选择
	fmt.Println("\n可用隧道列表:")
	var enabledProxies []frp.ProxyInfo
	for _, proxy := range proxies {
		if proxy.Status {
			fmt.Printf("%d. %s (ID: %d) - %s:%d -> %s\n",
				len(enabledProxies)+1, proxy.ProxyName, proxy.ID,
				proxy.ProxyType, proxy.LocalPort, proxy.FriendlyNode)
			enabledProxies = append(enabledProxies, proxy)
		}
	}

	if len(enabledProxies) == 0 {
		return fmt.Errorf("没有启用的隧道")
	}

	fmt.Print("\n请输入要启动的隧道编号 (多个用逗号分隔，回车启动全部): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	selection := strings.TrimSpace(scanner.Text())
	var selectedProxyIDs []string

	if selection == "" {
		// 启动全部隧道
		for _, proxy := range enabledProxies {
			selectedProxyIDs = append(selectedProxyIDs, fmt.Sprintf("%d", proxy.ID))
		}
	} else {
		// 解析用户选择
		selections := strings.Split(selection, ",")
		for _, sel := range selections {
			sel = strings.TrimSpace(sel)
			if sel == "" {
				continue
			}

			// 转换为数字索引
			index := 0
			if _, err := fmt.Sscanf(sel, "%d", &index); err != nil {
				return fmt.Errorf("无效的隧道编号: %s", sel)
			}

			if index < 1 || index > len(enabledProxies) {
				return fmt.Errorf("隧道编号超出范围: %d", index)
			}

			selectedProxyIDs = append(selectedProxyIDs, fmt.Sprintf("%d", enabledProxies[index-1].ID))
		}
	}

	if len(selectedProxyIDs) == 0 {
		return fmt.Errorf("没有选择任何隧道")
	}

	// 启动frpc
	if err := manager.StartFRPCWithCommand(userToken, selectedProxyIDs); err != nil {
		return err
	}

	fmt.Printf("✓ frpc启动成功，已启动 %d 个隧道\n", len(selectedProxyIDs))
	return nil
}

func handleFRPStatus() error {
	fmt.Println("=== 状态监控 ===")

	manager := frp.NewManager("./data")

	// 检查认证状态
	token := config.GetString("frp.openfrp.authorization")
	if token == "" {
		fmt.Println("❌ 未配置OpenFRP认证令牌")
		return nil
	}

	manager.SetAuthorization(token)
	fmt.Println("✅ OpenFRP认证已配置")

	// 检查frpc状态
	frpcStatus := manager.GetFRPCStatus()
	if frpcStatus == "运行中" {
		fmt.Println("✅ frpc客户端运行中")
	} else {
		fmt.Println("❌ frpc客户端已停止")
	}

	// 获取用户信息
	fmt.Println("\n正在获取用户信息...")
	userInfo, err := manager.GetUserInfo()
	if err != nil {
		fmt.Printf("❌ 获取用户信息失败: %v\n", err)
		return nil
	}

	fmt.Printf("✅ 用户: %s (%s)\n", userInfo.Username, userInfo.FriendlyGroup)
	fmt.Printf("   隧道配额: %d/%d\n", userInfo.Used, userInfo.Proxies)
	fmt.Printf("   剩余流量: %d MB\n", userInfo.Traffic)

	// 获取隧道状态
	fmt.Println("\n正在获取隧道状态...")
	proxies, err := manager.GetProxies()
	if err != nil {
		fmt.Printf("❌ 获取隧道列表失败: %v\n", err)
		return nil
	}

	if len(proxies) == 0 {
		fmt.Println("📋 暂无隧道")
	} else {
		onlineCount := 0
		for _, proxy := range proxies {
			if proxy.Status {
				onlineCount++
			}
		}
		fmt.Printf("📋 隧道状态: %d/%d 在线\n", onlineCount, len(proxies))

		// 显示在线隧道
		if onlineCount > 0 {
			fmt.Println("\n在线隧道:")
			for _, proxy := range proxies {
				if proxy.Status {
					fmt.Printf("  ✅ %s (%s) - %s\n", proxy.ProxyName, proxy.ProxyType, proxy.ConnectAddress)
				}
			}
		}

		// 显示离线隧道
		offlineCount := len(proxies) - onlineCount
		if offlineCount > 0 {
			fmt.Printf("\n离线隧道 (%d个):\n", offlineCount)
			for _, proxy := range proxies {
				if !proxy.Status {
					fmt.Printf("  ❌ %s (%s)\n", proxy.ProxyName, proxy.ProxyType)
				}
			}
		}
	}

	// 系统资源状态
	fmt.Println("\n=== 系统状态 ===")

	// 检查实例状态
	instanceManager := instance.NewManager("./data/instances")
	instances, err := instanceManager.ListInstances()
	if err == nil {
		runningCount := 0
		for _, inst := range instances {
			if inst.Status == instance.StatusRunning {
				runningCount++
			}
		}
		fmt.Printf("🎮 实例状态: %d/%d 运行中\n", runningCount, len(instances))
	}

	// 检查磁盘使用情况
	if err := displayDiskUsage(); err != nil {
		fmt.Printf("❌ 获取磁盘信息失败: %v\n", err)
	}

	return nil
}

func displayDiskUsage() error {
	// 检查数据目录大小
	dataDir := "./data"
	size, err := getDirSize(dataDir)
	if err != nil {
		return err
	}

	var sizeStr string
	if size < 1024*1024 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		sizeStr = fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
	} else {
		sizeStr = fmt.Sprintf("%.1f GB", float64(size)/(1024*1024*1024))
	}

	fmt.Printf("💾 数据目录大小: %s\n", sizeStr)
	return nil
}

func getDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 忽略错误，继续计算
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

func handleViewConfig() error {
	fmt.Println("\n=== 当前配置 ===")

	// OpenFRP配置
	fmt.Println("\n[OpenFRP配置]")
	token := config.GetString("frp.openfrp.authorization")
	if token != "" {
		// 只显示前8位和后4位
		maskedToken := token[:8] + "****" + token[len(token)-4:]
		fmt.Printf("认证令牌: %s\n", maskedToken)
	} else {
		fmt.Println("认证令牌: 未配置")
	}

	fmt.Printf("API地址: %s\n", config.GetString("frp.openfrp.api_url"))
	fmt.Printf("默认节点: %d\n", config.GetInt("frp.openfrp.default_node_id"))

	// 系统配置
	fmt.Println("\n[系统配置]")
	fmt.Printf("数据目录: %s\n", config.GetString("app.data_dir"))
	fmt.Printf("日志级别: %s\n", config.GetString("app.log_level"))
	fmt.Printf("自动更新: %t\n", config.GetBool("app.auto_update"))

	// Java配置
	fmt.Println("\n[Java配置]")
	fmt.Printf("默认Java路径: %s\n", config.GetString("java.default_path"))
	fmt.Printf("默认最大内存: %s\n", config.GetString("java.default_max_memory"))
	fmt.Printf("默认最小内存: %s\n", config.GetString("java.default_min_memory"))

	return nil
}

func handleEditFRPConfig(scanner *bufio.Scanner) error {
	fmt.Println("\n=== 修改OpenFRP配置 ===")

	fmt.Println("1. 修改认证令牌")
	fmt.Println("2. 修改用户访问密钥")
	fmt.Println("3. 修改API地址")
	fmt.Println("4. 修改默认节点")
	fmt.Print("请选择要修改的配置 (1-4): ")

	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		fmt.Print("请输入新的认证令牌: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newToken := strings.TrimSpace(scanner.Text())
		if newToken != "" {
			config.Set("frp.openfrp.authorization", newToken)
			if err := config.SaveConfig(); err != nil {
				return err
			}
			fmt.Println("✓ 认证令牌已更新")
		}

	case "2":
		fmt.Printf("当前用户访问密钥: %s\n", config.GetString("frp.openfrp.user_token"))
		fmt.Print("请输入新的用户访问密钥: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newUserToken := strings.TrimSpace(scanner.Text())
		if newUserToken != "" {
			config.Set("frp.openfrp.user_token", newUserToken)
			if err := config.SaveConfig(); err != nil {
				return err
			}
			fmt.Println("✓ 用户访问密钥已更新")
		}

	case "3":
		fmt.Printf("当前API地址: %s\n", config.GetString("frp.openfrp.api_url"))
		fmt.Print("请输入新的API地址: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newURL := strings.TrimSpace(scanner.Text())
		if newURL != "" {
			config.Set("frp.openfrp.api_url", newURL)
			if err := config.SaveConfig(); err != nil {
				return err
			}
			fmt.Println("✓ API地址已更新")
		}

	case "4":
		fmt.Printf("当前默认节点: %d\n", config.GetInt("frp.openfrp.default_node_id"))
		fmt.Print("请输入新的默认节点ID: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newNodeID := strings.TrimSpace(scanner.Text())
		if newNodeID != "" {
			config.Set("frp.openfrp.default_node_id", newNodeID)
			if err := config.SaveConfig(); err != nil {
				return err
			}
			fmt.Println("✓ 默认节点已更新")
		}

	default:
		fmt.Println("无效选择")
	}

	return nil
}

func handleEditSystemConfig(scanner *bufio.Scanner) error {
	fmt.Println("\n=== 修改系统配置 ===")

	fmt.Println("1. 修改数据目录")
	fmt.Println("2. 修改日志级别")
	fmt.Println("3. 修改自动更新设置")
	fmt.Print("请选择要修改的配置 (1-3): ")

	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		fmt.Printf("当前数据目录: %s\n", config.GetString("app.data_dir"))
		fmt.Print("请输入新的数据目录: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newDir := strings.TrimSpace(scanner.Text())
		if newDir != "" {
			config.Set("app.data_dir", newDir)
			if err := config.SaveConfig(); err != nil {
				return err
			}
			fmt.Println("✓ 数据目录已更新")
		}

	case "2":
		fmt.Printf("当前日志级别: %s\n", config.GetString("app.log_level"))
		fmt.Println("可选级别: debug, info, warn, error")
		fmt.Print("请输入新的日志级别: ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newLevel := strings.TrimSpace(scanner.Text())
		if newLevel != "" {
			config.Set("app.log_level", newLevel)
			if err := config.SaveConfig(); err != nil {
				return err
			}
			fmt.Println("✓ 日志级别已更新")
		}

	case "3":
		fmt.Printf("当前自动更新: %t\n", config.GetBool("app.auto_update"))
		fmt.Print("是否启用自动更新? (y/N): ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}
		newValue := strings.ToLower(strings.TrimSpace(scanner.Text()))
		autoUpdate := (newValue == "y" || newValue == "yes")
		config.Set("app.auto_update", autoUpdate)
		if err := config.SaveConfig(); err != nil {
			return err
		}
		fmt.Printf("✓ 自动更新已设置为: %t\n", autoUpdate)

	default:
		fmt.Println("无效选择")
	}

	return nil
}

func handleResetConfig() error {
	// 删除配置文件
	configFile := "./configs/config.yaml"
	if err := os.Remove(configFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除配置文件失败: %w", err)
	}

	// 重新初始化配置
	configManager := config.NewManager("")
	return configManager.Initialize()
}

func handleExportConfig(scanner *bufio.Scanner) error {
	fmt.Print("请输入导出文件路径 (默认: ./config_backup.yaml): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	exportPath := strings.TrimSpace(scanner.Text())
	if exportPath == "" {
		exportPath = "./config_backup.yaml"
	}

	// 读取当前配置文件
	configFile := "./configs/config.yaml"
	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 写入导出文件
	if err := os.WriteFile(exportPath, content, 0644); err != nil {
		return fmt.Errorf("写入导出文件失败: %w", err)
	}

	fmt.Printf("✓ 配置已导出到: %s\n", exportPath)
	return nil
}

func handleImportConfig(scanner *bufio.Scanner) error {
	fmt.Print("请输入配置文件路径: ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	importPath := strings.TrimSpace(scanner.Text())
	if importPath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 检查文件是否存在
	if _, err := os.Stat(importPath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", importPath)
	}

	// 读取导入文件
	content, err := os.ReadFile(importPath)
	if err != nil {
		return fmt.Errorf("读取导入文件失败: %w", err)
	}

	// 备份当前配置
	configFile := "./configs/config.yaml"
	backupFile := "./configs/config.yaml.backup"
	if _, err := os.Stat(configFile); err == nil {
		if err := os.Rename(configFile, backupFile); err != nil {
			return fmt.Errorf("备份当前配置失败: %w", err)
		}
	}

	// 写入新配置
	if err := os.WriteFile(configFile, content, 0644); err != nil {
		// 恢复备份
		os.Rename(backupFile, configFile)
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	fmt.Printf("✓ 配置已从 %s 导入\n", importPath)
	fmt.Println("注意: 需要重启程序以使新配置生效")
	return nil
}

// createConfigMenu 创建配置管理子菜单
func createConfigMenu() *menu.Menu {
	configMenu := menu.NewMenu("配置管理", "查看和修改系统配置")

	configMenu.AddItems(
		menu.NewMenuItem("view", "查看当前配置", "显示当前系统配置").
			WithHandler(func() error {
				return handleViewConfig()
			}),

		menu.NewMenuItem("frp", "修改OpenFRP配置", "修改OpenFRP相关配置").
			WithHandler(func() error {
				scanner := bufio.NewScanner(os.Stdin)
				return handleEditFRPConfig(scanner)
			}),

		menu.NewMenuItem("system", "修改系统配置", "修改系统相关配置").
			WithHandler(func() error {
				scanner := bufio.NewScanner(os.Stdin)
				return handleEditSystemConfig(scanner)
			}),

		menu.NewMenuItem("reset", "重置配置", "重置所有配置到默认值").
			WithHandler(func() error {
				confirmPrompt := promptui.Prompt{
					Label:     "确定要重置所有配置吗? 这将删除所有自定义设置",
					IsConfirm: true,
				}
				_, err := confirmPrompt.Run()
				if err == nil {
					if err := handleResetConfig(); err != nil {
						return fmt.Errorf("重置配置失败: %w", err)
					}
					fmt.Println("✓ 配置已重置")
				}
				return nil
			}),

		menu.NewMenuItem("export", "导出配置", "导出当前配置到文件").
			WithHandler(func() error {
				scanner := bufio.NewScanner(os.Stdin)
				return handleExportConfig(scanner)
			}),

		menu.NewMenuItem("import", "导入配置", "从文件导入配置").
			WithHandler(func() error {
				scanner := bufio.NewScanner(os.Stdin)
				return handleImportConfig(scanner)
			}),
	)

	return configMenu
}

func handleConfigManagement() error {
	// 这个函数现在只是为了兼容性，实际应该使用createConfigMenu
	fmt.Println("配置管理功能请通过主菜单 -> 系统设置 -> 配置管理 访问")
	return nil
}

func handleAbout() error {
	fmt.Println("=== 关于 EasilyPanel5 ===")
	fmt.Println("版本: v1.0.0")
	fmt.Println("作者: EasilyPanel Team")
	fmt.Println("描述: 跨平台通用游戏服务器管理工具")
	fmt.Println("支持: Minecraft、Java环境管理、内网穿透等")
	return nil
}

func handleCreateInstanceFromDownload(filePath, serverType, version string) error {
	fmt.Println("\n=== 从下载创建实例 ===")

	scanner := bufio.NewScanner(os.Stdin)

	// 输入实例名称
	defaultName := fmt.Sprintf("%s-%s", serverType, version)
	fmt.Printf("请输入实例名称 (默认: %s): ", defaultName)
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	instanceName := strings.TrimSpace(scanner.Text())
	if instanceName == "" {
		instanceName = defaultName
	}

	// 输入端口
	fmt.Print("请输入服务器端口 (默认: 25565): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	port := strings.TrimSpace(scanner.Text())
	if port == "" {
		port = "25565"
	}

	// 创建实例
	manager := instance.NewManager("./data/instances")

	fmt.Printf("正在创建实例 '%s'...\n", instanceName)

	// 检测Java路径
	detector := java.NewDetector()
	javaVersions, _ := detector.DetectJava(false)
	javaPath := "java"
	if len(javaVersions) > 0 {
		javaPath = javaVersions[0].Path
	}

	inst, err := manager.CreateMinecraftInstance(instanceName, version, serverType, javaPath)
	if err != nil {
		return fmt.Errorf("创建实例失败: %w", err)
	}

	// 复制服务端文件到实例目录
	instanceDir := inst.GetWorkDir("./data/instances")
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		return fmt.Errorf("创建实例目录失败: %w", err)
	}

	// 获取原文件名
	originalFileName := filepath.Base(filePath)
	targetFilePath := filepath.Join(instanceDir, originalFileName)

	fmt.Printf("正在复制服务端文件到实例目录...\n")
	if err := copyFile(filePath, targetFilePath); err != nil {
		return fmt.Errorf("复制服务端文件失败: %w", err)
	}

	// 设置服务端文件路径为实例目录中的文件
	inst.ServerJar = originalFileName // 只保存文件名，因为工作目录已经设置
	if err := manager.UpdateInstance(inst); err != nil {
		return fmt.Errorf("保存实例配置失败: %w", err)
	}

	fmt.Printf("✓ 实例 '%s' 创建成功\n", instanceName)
	fmt.Printf("服务端: %s %s\n", serverType, version)
	fmt.Printf("端口: %s\n", port)
	fmt.Printf("服务端文件: %s\n", filePath)

	return nil
}

// handleCommandLine 处理命令行模式
func handleCommandLine(args []string, configFile, dataDir, logLevel string) {
	if len(args) == 0 {
		fmt.Println("错误: 缺少命令参数")
		fmt.Println("使用 'easilypanel -help' 查看帮助信息")
		return
	}

	// 初始化配置
	configManager := config.NewManager(configFile)
	if err := configManager.Initialize(); err != nil {
		fmt.Printf("初始化配置失败: %v\n", err)
		return
	}

	// 设置配置
	if dataDir != "./data" {
		config.Set("app.data_dir", dataDir)
	}
	if logLevel != "info" {
		config.Set("app.log_level", logLevel)
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "instance":
		handleInstanceCommand(subArgs, dataDir)
	case "frp":
		handleFRPCommand(subArgs, dataDir)
	case "java":
		handleJavaCommand(subArgs)
	case "download":
		handleDownloadCommand(subArgs, dataDir)
	case "config":
		handleConfigCommand(subArgs)
	default:
		fmt.Printf("未知命令: %s\n", command)
		fmt.Println("使用 'easilypanel -help' 查看可用命令")
	}
}

func handleInstanceCommand(args []string, dataDir string) {
	if len(args) == 0 {
		fmt.Println("实例管理命令:")
		fmt.Println("  list          列出所有实例")
		fmt.Println("  start NAME    启动指定实例")
		fmt.Println("  stop NAME     停止指定实例")
		fmt.Println("  status NAME   查看实例状态")
		return
	}

	manager := instance.NewManager(filepath.Join(dataDir, "instances"))
	processManager := instance.NewProcessManager(filepath.Join(dataDir, "instances"))

	switch args[0] {
	case "list":
		instances, err := manager.ListInstances()
		if err != nil {
			fmt.Printf("获取实例列表失败: %v\n", err)
			return
		}

		if len(instances) == 0 {
			fmt.Println("暂无实例")
			return
		}

		fmt.Printf("实例列表 (%d个):\n", len(instances))
		for _, inst := range instances {
			fmt.Printf("  %s (%s) - %s\n", inst.Name, inst.Type, inst.Status)
		}

	case "start":
		if len(args) < 2 {
			fmt.Println("错误: 缺少实例名称")
			return
		}
		instanceName := args[1]
		if err := processManager.StartInstance(instanceName); err != nil {
			fmt.Printf("启动实例失败: %v\n", err)
		} else {
			fmt.Printf("实例 '%s' 启动成功\n", instanceName)
		}

	case "stop":
		if len(args) < 2 {
			fmt.Println("错误: 缺少实例名称")
			return
		}
		instanceName := args[1]
		if err := processManager.StopInstance(instanceName); err != nil {
			fmt.Printf("停止实例失败: %v\n", err)
		} else {
			fmt.Printf("实例 '%s' 停止成功\n", instanceName)
		}

	case "status":
		if len(args) < 2 {
			fmt.Println("错误: 缺少实例名称")
			return
		}
		instanceName := args[1]
		instances, err := manager.ListInstances()
		if err != nil {
			fmt.Printf("获取实例信息失败: %v\n", err)
			return
		}

		for _, inst := range instances {
			if inst.Name == instanceName {
				fmt.Printf("实例: %s\n", inst.Name)
				fmt.Printf("类型: %s\n", inst.Type)
				fmt.Printf("端口: %d\n", inst.Port)
				fmt.Printf("状态: %s\n", inst.Status)
				return
			}
		}
		fmt.Printf("未找到实例: %s\n", instanceName)

	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

func handleFRPCommand(args []string, dataDir string) {
	if len(args) == 0 {
		fmt.Println("FRP管理命令:")
		fmt.Println("  status        查看frpc状态")
		fmt.Println("  start         启动frpc")
		fmt.Println("  stop          停止frpc")
		fmt.Println("  restart       重启frpc")
		return
	}

	manager := frp.NewManager(dataDir)

	switch args[0] {
	case "status":
		status := manager.GetFRPCStatus()
		fmt.Printf("frpc状态: %s\n", status)

	case "start":
		if err := manager.StartFRPC(); err != nil {
			fmt.Printf("启动frpc失败: %v\n", err)
		} else {
			fmt.Println("frpc启动成功")
		}

	case "stop":
		if err := manager.StopFRPC(); err != nil {
			fmt.Printf("停止frpc失败: %v\n", err)
		} else {
			fmt.Println("frpc停止成功")
		}

	case "restart":
		if err := manager.RestartFRPC(); err != nil {
			fmt.Printf("重启frpc失败: %v\n", err)
		} else {
			fmt.Println("frpc重启成功")
		}

	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

func handleJavaCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Java管理命令:")
		fmt.Println("  detect        检测Java版本")
		fmt.Println("  list          列出Java版本")
		return
	}

	detector := java.NewDetector()

	switch args[0] {
	case "detect", "list":
		versions, err := detector.DetectJava(true)
		if err != nil {
			fmt.Printf("检测Java失败: %v\n", err)
			return
		}

		if len(versions) == 0 {
			fmt.Println("未检测到Java环境")
			return
		}

		fmt.Printf("检测到 %d 个Java版本:\n", len(versions))
		for _, version := range versions {
			fmt.Printf("  Java %d (%s)\n", version.Version, version.Path)
		}

	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

func handleDownloadCommand(args []string, dataDir string) {
	if len(args) == 0 {
		fmt.Println("下载管理命令:")
		fmt.Println("  list          列出可用服务端")
		fmt.Println("  files         查看已下载文件")
		return
	}

	dm := download.NewDownloadManager(dataDir)

	switch args[0] {
	case "list":
		servers, err := dm.ListAvailableServers()
		if err != nil {
			fmt.Printf("获取服务端列表失败: %v\n", err)
			return
		}

		fmt.Printf("可用服务端 (%d个):\n", len(servers))
		for _, server := range servers {
			fmt.Printf("  %s\n", server.Name)
		}

	case "files":
		dm.PrintDownloadedFiles()

	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

func handleConfigCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("配置管理命令:")
		fmt.Println("  show          显示当前配置")
		fmt.Println("  set KEY VALUE 设置配置项")
		fmt.Println("  get KEY       获取配置项")
		return
	}

	switch args[0] {
	case "show":
		fmt.Println("当前配置:")
		fmt.Printf("  数据目录: %s\n", config.GetString("app.data_dir"))
		fmt.Printf("  日志级别: %s\n", config.GetString("app.log_level"))
		fmt.Printf("  自动更新: %t\n", config.GetBool("app.auto_update"))

	case "set":
		if len(args) < 3 {
			fmt.Println("错误: 缺少参数")
			fmt.Println("用法: config set KEY VALUE")
			return
		}
		key, value := args[1], args[2]
		config.Set(key, value)
		if err := config.SaveConfig(); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
		} else {
			fmt.Printf("配置已设置: %s = %s\n", key, value)
		}

	case "get":
		if len(args) < 2 {
			fmt.Println("错误: 缺少参数")
			fmt.Println("用法: config get KEY")
			return
		}
		key := args[1]
		value := config.GetString(key)
		fmt.Printf("%s = %s\n", key, value)

	default:
		fmt.Printf("未知子命令: %s\n", args[0])
	}
}

func handleEditInstanceConfig(manager *instance.Manager, inst *instance.Instance, scanner *bufio.Scanner) error {
	fmt.Printf("\n=== 编辑实例配置: %s ===\n", inst.Name)

	for {
		fmt.Println("\n可编辑的配置项:")
		fmt.Println("1. 最大内存")
		fmt.Println("2. 最小内存")
		fmt.Println("3. Java参数")
		fmt.Println("4. 服务器参数")
		fmt.Println("5. 启动命令")
		fmt.Println("6. 自动启动")
		fmt.Println("7. 自动重启")
		fmt.Println("0. 保存并返回")
		fmt.Print("请选择要编辑的配置 (0-7): ")

		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		choice := strings.TrimSpace(scanner.Text())

		switch choice {
		case "1":
			fmt.Printf("当前最大内存: %s\n", inst.MaxMemory)
			fmt.Print("请输入新的最大内存 (如: 2G, 1024M): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			newValue := strings.TrimSpace(scanner.Text())
			if newValue != "" {
				inst.MaxMemory = newValue
				fmt.Println("✓ 最大内存已更新")
			}

		case "2":
			fmt.Printf("当前最小内存: %s\n", inst.MinMemory)
			fmt.Print("请输入新的最小内存 (如: 1G, 512M): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			newValue := strings.TrimSpace(scanner.Text())
			if newValue != "" {
				inst.MinMemory = newValue
				fmt.Println("✓ 最小内存已更新")
			}

		case "3":
			fmt.Printf("当前Java参数: %v\n", inst.JavaArgs)
			fmt.Print("请输入新的Java参数 (用空格分隔): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			newValue := strings.TrimSpace(scanner.Text())
			if newValue != "" {
				inst.JavaArgs = strings.Fields(newValue)
				fmt.Println("✓ Java参数已更新")
			}

		case "4":
			fmt.Printf("当前服务器参数: %v\n", inst.ServerArgs)
			fmt.Print("请输入新的服务器参数 (用空格分隔): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			newValue := strings.TrimSpace(scanner.Text())
			if newValue != "" {
				inst.ServerArgs = strings.Fields(newValue)
				fmt.Println("✓ 服务器参数已更新")
			}

		case "5":
			return handleEditStartCommand(inst, scanner)

		case "6":
			fmt.Printf("当前自动启动: %t\n", inst.AutoStart)
			fmt.Print("是否启用自动启动? (y/N): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			newValue := strings.ToLower(strings.TrimSpace(scanner.Text()))
			inst.AutoStart = (newValue == "y" || newValue == "yes")
			fmt.Printf("✓ 自动启动已设置为: %t\n", inst.AutoStart)

		case "7":
			fmt.Printf("当前自动重启: %t\n", inst.AutoRestart)
			fmt.Print("是否启用自动重启? (y/N): ")
			if !scanner.Scan() {
				return fmt.Errorf("读取输入失败")
			}
			newValue := strings.ToLower(strings.TrimSpace(scanner.Text()))
			inst.AutoRestart = (newValue == "y" || newValue == "yes")
			fmt.Printf("✓ 自动重启已设置为: %t\n", inst.AutoRestart)

		case "0":
			// 保存配置
			if err := manager.UpdateInstance(inst); err != nil {
				return fmt.Errorf("保存配置失败: %w", err)
			}
			fmt.Println("✓ 配置已保存")
			return nil

		default:
			fmt.Println("无效选择")
		}
	}
}

// handleEditStartCommand 编辑启动命令
func handleEditStartCommand(inst *instance.Instance, scanner *bufio.Scanner) error {
	fmt.Println("\n=== 编辑启动命令 ===")

	// 显示当前状态
	if inst.UseCustomCmd && inst.StartCmd != "" {
		fmt.Printf("当前使用自定义启动命令: %s\n", inst.StartCmd)
	} else {
		fmt.Println("当前使用默认启动命令 (java -jar)")
	}

	fmt.Println("\n启动命令选项:")
	fmt.Println("1. 使用默认启动命令 (java -jar)")
	fmt.Println("2. 设置自定义启动命令")
	fmt.Println("3. 查看当前完整启动命令")
	fmt.Println("0. 返回")
	fmt.Print("请选择 (0-3): ")

	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())

	switch choice {
	case "1":
		// 使用默认启动命令
		inst.UseCustomCmd = false
		inst.StartCmd = ""
		fmt.Println("✓ 已设置为使用默认启动命令")

	case "2":
		// 设置自定义启动命令
		fmt.Println("\n自定义启动命令说明:")
		fmt.Println("- 可以使用任意命令替代默认的 java -jar")
		fmt.Println("- 支持完整的命令行参数")
		fmt.Println("- 工作目录会自动设置为实例目录")
		fmt.Println("- 示例: python3 server.py")
		fmt.Println("- 示例: ./bedrock_server")
		fmt.Println("- 示例: java -Xmx2G -Xms1G -jar server.jar nogui")
		fmt.Println()

		if inst.StartCmd != "" {
			fmt.Printf("当前自定义命令: %s\n", inst.StartCmd)
		}

		fmt.Print("请输入新的启动命令 (留空取消): ")
		if !scanner.Scan() {
			return fmt.Errorf("读取输入失败")
		}

		newCmd := strings.TrimSpace(scanner.Text())
		if newCmd != "" {
			inst.StartCmd = newCmd
			inst.UseCustomCmd = true
			fmt.Printf("✓ 自定义启动命令已设置为: %s\n", newCmd)
		} else {
			fmt.Println("操作已取消")
		}

	case "3":
		// 查看当前完整启动命令
		fmt.Println("\n当前完整启动命令:")
		if inst.UseCustomCmd && inst.StartCmd != "" {
			fmt.Printf("自定义命令: %s\n", inst.StartCmd)
		} else {
			// 显示默认命令（需要模拟生成）
			defaultCmd := fmt.Sprintf("java -Xmx%s -Xms%s",
				getMemoryOrDefault(inst.MaxMemory, "1G"),
				getMemoryOrDefault(inst.MinMemory, "512M"))

			if len(inst.JavaArgs) > 0 {
				defaultCmd += " " + strings.Join(inst.JavaArgs, " ")
			}

			defaultCmd += " -jar " + inst.ServerJar

			if len(inst.ServerArgs) > 0 {
				defaultCmd += " " + strings.Join(inst.ServerArgs, " ")
			}

			fmt.Printf("默认命令: %s\n", defaultCmd)
		}

	case "0":
		return nil

	default:
		fmt.Println("无效选择")
	}

	return nil
}

// getMemoryOrDefault 获取内存设置或默认值
func getMemoryOrDefault(memory, defaultValue string) string {
	if memory != "" {
		return memory
	}
	return defaultValue
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 确保数据写入磁盘
	err = destFile.Sync()
	if err != nil {
		return fmt.Errorf("同步文件失败: %w", err)
	}

	return nil
}

func handleViewInstanceLogs(inst *instance.Instance) error {
	fmt.Printf("\n=== 查看实例日志: %s ===\n", inst.Name)

	logFile := fmt.Sprintf("./data/instances/%s/logs/latest.log", inst.Name)

	// 检查日志文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("日志文件不存在，实例可能尚未启动过")
		return nil
	}

	// 读取日志文件的最后50行
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("读取日志文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	startLine := 0
	if len(lines) > 50 {
		startLine = len(lines) - 50
	}

	fmt.Println("最近的日志内容 (最后50行):")
	fmt.Println(strings.Repeat("-", 60))
	for i := startLine; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			fmt.Println(lines[i])
		}
	}
	fmt.Println(strings.Repeat("-", 60))

	return nil
}

func showTunnelList(manager *frp.Manager) error {
	fmt.Println("\n正在获取隧道列表...")

	proxies, err := manager.GetProxies()
	if err != nil {
		return fmt.Errorf("获取隧道列表失败: %w", err)
	}

	if len(proxies) == 0 {
		fmt.Println("暂无隧道")
		return nil
	}

	fmt.Printf("\n隧道列表 (%d 个):\n", len(proxies))
	for _, proxy := range proxies {
		status := "离线"
		if proxy.Status {
			status = "在线"
		}

		fmt.Printf("ID: %d | 名称: %s | 类型: %s | 本地端口: %d | 状态: %s\n",
			proxy.ID, proxy.ProxyName, proxy.ProxyType, proxy.LocalPort, status)

		if proxy.Status && proxy.ConnectAddress != "" {
			fmt.Printf("  连接地址: %s\n", proxy.ConnectAddress)
		}
		fmt.Println()
	}

	return nil
}

func createNewTunnel(manager *frp.Manager, scanner *bufio.Scanner) error {
	fmt.Println("\n=== 创建新隧道 ===")

	// 输入隧道名称
	fmt.Print("请输入隧道名称: ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}
	tunnelName := strings.TrimSpace(scanner.Text())
	if tunnelName == "" {
		return fmt.Errorf("隧道名称不能为空")
	}

	// 选择隧道类型
	fmt.Println("\n隧道类型:")
	fmt.Println("1. TCP")
	fmt.Println("2. UDP")
	fmt.Println("3. HTTP")
	fmt.Println("4. HTTPS")
	fmt.Print("请选择类型 (1-4): ")

	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	var tunnelType string
	switch strings.TrimSpace(scanner.Text()) {
	case "1":
		tunnelType = "tcp"
	case "2":
		tunnelType = "udp"
	case "3":
		tunnelType = "http"
	case "4":
		tunnelType = "https"
	default:
		return fmt.Errorf("无效的隧道类型")
	}

	// 输入本地端口
	fmt.Print("请输入本地端口: ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}
	localPortStr := strings.TrimSpace(scanner.Text())
	if localPortStr == "" {
		return fmt.Errorf("本地端口不能为空")
	}

	// 获取节点列表
	fmt.Println("正在获取节点列表...")
	nodes, err := manager.GetNodes()
	if err != nil {
		return fmt.Errorf("获取节点列表失败: %w", err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("未找到可用节点")
	}

	// 显示节点列表
	fmt.Println("\n可用节点:")
	for i, node := range nodes {
		status := "离线"
		if node.Status == 1 {
			status = "在线"
		}
		fmt.Printf("%d. %s (%s) - %s\n", i+1, node.Name, node.Hostname, status)
	}

	fmt.Print("请选择节点 (输入序号): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	nodeChoice := strings.TrimSpace(scanner.Text())
	nodeIndex := -1
	for i := range nodes {
		if fmt.Sprintf("%d", i+1) == nodeChoice {
			nodeIndex = i
			break
		}
	}

	if nodeIndex == -1 {
		return fmt.Errorf("无效的节点选择")
	}

	selectedNode := nodes[nodeIndex]

	// 创建隧道
	fmt.Printf("\n正在创建隧道 '%s'...\n", tunnelName)

	tunnelConfig := &frp.CreateProxyRequest{
		Name:      tunnelName,
		Type:      tunnelType,
		LocalAddr: "127.0.0.1",
		LocalPort: localPortStr,
		NodeID:    selectedNode.ID,
	}

	err = manager.CreateProxy(tunnelConfig)
	if err != nil {
		return fmt.Errorf("创建隧道失败: %w", err)
	}

	fmt.Printf("✓ 隧道创建成功\n")
	fmt.Printf("隧道名称: %s\n", tunnelName)
	fmt.Printf("类型: %s\n", tunnelType)
	fmt.Printf("本地端口: %s\n", localPortStr)
	fmt.Printf("节点: %s\n", selectedNode.Name)

	return nil
}

func deleteTunnel(manager *frp.Manager, scanner *bufio.Scanner) error {
	fmt.Println("\n=== 删除隧道 ===")

	// 获取隧道列表
	proxies, err := manager.GetProxies()
	if err != nil {
		return fmt.Errorf("获取隧道列表失败: %w", err)
	}

	if len(proxies) == 0 {
		fmt.Println("暂无隧道可删除")
		return nil
	}

	// 显示隧道列表
	fmt.Println("现有隧道:")
	for i, proxy := range proxies {
		fmt.Printf("%d. %s (ID: %d) - %s\n", i+1, proxy.ProxyName, proxy.ID, proxy.ProxyType)
	}

	fmt.Print("\n请选择要删除的隧道 (输入序号): ")
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	choice := strings.TrimSpace(scanner.Text())
	proxyIndex := -1
	for i := range proxies {
		if fmt.Sprintf("%d", i+1) == choice {
			proxyIndex = i
			break
		}
	}

	if proxyIndex == -1 {
		return fmt.Errorf("无效的隧道选择")
	}

	selectedProxy := proxies[proxyIndex]

	// 确认删除
	fmt.Printf("确定要删除隧道 '%s' (ID: %d) 吗? (y/N): ", selectedProxy.ProxyName, selectedProxy.ID)
	if !scanner.Scan() {
		return fmt.Errorf("读取输入失败")
	}

	confirm := strings.ToLower(strings.TrimSpace(scanner.Text()))
	if confirm != "y" && confirm != "yes" {
		fmt.Println("取消删除")
		return nil
	}

	// 删除隧道
	fmt.Printf("正在删除隧道 '%s'...\n", selectedProxy.ProxyName)
	if err := manager.DeleteProxy(selectedProxy.ID); err != nil {
		return fmt.Errorf("删除隧道失败: %w", err)
	}

	fmt.Printf("✓ 隧道 '%s' 已删除\n", selectedProxy.ProxyName)
	return nil
}

func handleViewFRPCLogs(manager *frp.Manager) error {
	fmt.Println("\n=== frpc日志 ===")

	logFile := "./data/logs/frpc.log"

	// 检查日志文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Println("日志文件不存在，frpc可能尚未启动过")
		return nil
	}

	// 读取日志文件的最后100行
	content, err := os.ReadFile(logFile)
	if err != nil {
		return fmt.Errorf("读取日志文件失败: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	startLine := 0
	if len(lines) > 100 {
		startLine = len(lines) - 100
	}

	fmt.Println("最近的日志内容 (最后100行):")
	fmt.Println(strings.Repeat("-", 80))
	for i := startLine; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "" {
			fmt.Println(lines[i])
		}
	}
	fmt.Println(strings.Repeat("-", 80))

	return nil
}

func handleViewFRPCConfig(manager *frp.Manager) error {
	fmt.Println("\n=== frpc配置 ===")

	configFile := "./data/configs/frpc.ini"

	// 检查配置文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("配置文件不存在，请先生成配置")
		return nil
	}

	// 读取配置文件
	content, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	fmt.Printf("配置文件路径: %s\n", configFile)
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println(string(content))
	fmt.Println(strings.Repeat("-", 60))

	return nil
}
