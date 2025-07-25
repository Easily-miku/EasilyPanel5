package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"easilypanel/internal/config"
	"easilypanel/internal/logger"
)

// App CLI应用结构
type App struct {
	name    string
	version string
	desc    string
	config  *config.Config
	logger  *logger.Logger
	rootCmd *cobra.Command
}

// NewApp 创建新的CLI应用
func NewApp(name, version, desc string, cfg *config.Config, log *logger.Logger) *App {
	app := &App{
		name:    name,
		version: version,
		desc:    desc,
		config:  cfg,
		logger:  log,
	}
	
	app.setupCommands()
	return app
}

// setupCommands 设置命令
func (a *App) setupCommands() {
	a.rootCmd = &cobra.Command{
		Use:     a.name,
		Short:   a.desc,
		Version: a.version,
		Long: fmt.Sprintf(`%s v%s

%s

支持的操作模式：
  1. 交互式模式：直接运行程序，通过菜单进行操作
  2. 命令行模式：使用子命令直接执行特定功能

使用 '%s help' 查看所有可用命令。`, a.name, a.version, a.desc, a.name),
	}
	
	// 添加子命令
	a.addInstanceCommands()
	a.addDownloadCommands()
	a.addJavaCommands()
	a.addBackupCommands()
	a.addConfigCommands()
}

// Execute 执行命令
func (a *App) Execute() error {
	return a.rootCmd.Execute()
}

// RunInteractiveMenu 运行交互式菜单
func (a *App) RunInteractiveMenu() error {
	a.logger.系统启动(a.version)
	
	scanner := bufio.NewScanner(os.Stdin)
	
	for {
		a.showMainMenu()
		
		fmt.Print("请选择操作 (输入数字): ")
		if !scanner.Scan() {
			break
		}
		
		choice := strings.TrimSpace(scanner.Text())
		if choice == "0" || choice == "退出" {
			break
		}
		
		if err := a.handleMenuChoice(choice); err != nil {
			a.logger.错误("菜单操作失败: %v", err)
			fmt.Printf("操作失败: %v\n", err)
		}
		
		fmt.Println("\n按回车键继续...")
		scanner.Scan()
	}
	
	a.logger.系统关闭()
	fmt.Println("感谢使用 EasilyPanel5！")
	return nil
}

// showMainMenu 显示主菜单
func (a *App) showMainMenu() {
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║                    %s v%s                    ║
║                  跨平台通用游戏服务器管理工具                  ║
╠══════════════════════════════════════════════════════════════╣
║  1. 实例管理                                                 ║
║     1.1 创建Minecraft实例    1.2 创建空白实例               ║
║     1.3 查看实例列表         1.4 启动/停止实例              ║
║     1.5 删除实例                                             ║
║                                                              ║
║  2. 服务端下载                                               ║
║     2.1 FastMirror下载源     2.2 MCSL-Sync下载源           ║
║                                                              ║
║  3. Java环境                                                 ║
║     3.1 检测Java环境         3.2 管理Java版本               ║
║                                                              ║
║  4. 内网穿透                                                 ║
║     4.1 配置OpenFRP          4.2 管理隧道                   ║
║                                                              ║
║  5. 文件管理                                                 ║
║     5.1 启动FTP服务          5.2 文件浏览器                 ║
║                                                              ║
║  6. 备份管理                                                 ║
║     6.1 创建备份             6.2 恢复备份                   ║
║     6.3 管理备份                                             ║
║                                                              ║
║  7. 系统设置                                                 ║
║     7.1 进程守护配置         7.2 日志设置                   ║
║                                                              ║
║  0. 退出程序                                                 ║
╚══════════════════════════════════════════════════════════════╝
`, a.name, a.version)
}

// handleMenuChoice 处理菜单选择
func (a *App) handleMenuChoice(choice string) error {
	switch choice {
	case "1", "1.1":
		return a.handleInstanceMenu()
	case "1.2":
		return a.createBlankInstance()
	case "1.3":
		return a.listInstances()
	case "1.4":
		return a.manageInstanceState()
	case "1.5":
		return a.deleteInstance()
	case "2", "2.1":
		return a.handleDownloadMenu("fastmirror")
	case "2.2":
		return a.handleDownloadMenu("mcsl")
	case "3", "3.1":
		return a.detectJava()
	case "3.2":
		return a.manageJava()
	case "4", "4.1":
		return a.configureFRP()
	case "4.2":
		return a.manageTunnels()
	case "5", "5.1":
		return a.startFTPServer()
	case "5.2":
		return a.fileBrowser()
	case "6", "6.1":
		return a.createBackup()
	case "6.2":
		return a.restoreBackup()
	case "6.3":
		return a.manageBackups()
	case "7", "7.1":
		return a.configureDaemon()
	case "7.2":
		return a.configureLogging()
	default:
		return fmt.Errorf("无效的选择: %s", choice)
	}
}

// addInstanceCommands 添加实例管理命令
func (a *App) addInstanceCommands() {
	instanceCmd := &cobra.Command{
		Use:   "instance",
		Short: "实例管理",
		Long:  "管理游戏服务器实例，包括创建、启动、停止、删除等操作",
	}

	// 创建实例命令
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "创建新实例",
		RunE: func(cmd *cobra.Command, args []string) error {
			name, _ := cmd.Flags().GetString("name")
			instanceType, _ := cmd.Flags().GetString("type")
			version, _ := cmd.Flags().GetString("version")

			a.logger.信息("创建实例: %s (类型: %s, 版本: %s)", name, instanceType, version)
			fmt.Printf("创建实例功能正在开发中...\n")
			return nil
		},
	}
	createCmd.Flags().StringP("name", "n", "", "实例名称")
	createCmd.Flags().StringP("type", "t", "minecraft", "实例类型 (minecraft/blank)")
	createCmd.Flags().StringP("version", "v", "", "Minecraft版本")
	createCmd.MarkFlagRequired("name")

	// 启动实例命令
	startCmd := &cobra.Command{
		Use:   "start [实例名称]",
		Short: "启动实例",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			a.logger.信息("启动实例: %s", name)
			fmt.Printf("启动实例功能正在开发中...\n")
			return nil
		},
	}

	// 停止实例命令
	stopCmd := &cobra.Command{
		Use:   "stop [实例名称]",
		Short: "停止实例",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			a.logger.信息("停止实例: %s", name)
			fmt.Printf("停止实例功能正在开发中...\n")
			return nil
		},
	}

	// 列出实例命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有实例",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.logger.信息("列出实例")
			fmt.Printf("列出实例功能正在开发中...\n")
			return nil
		},
	}

	instanceCmd.AddCommand(createCmd, startCmd, stopCmd, listCmd)
	a.rootCmd.AddCommand(instanceCmd)
}

// addDownloadCommands 添加下载命令
func (a *App) addDownloadCommands() {
	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "下载服务端",
		Long:  "从各种下载源下载Minecraft服务端",
		RunE: func(cmd *cobra.Command, args []string) error {
			source, _ := cmd.Flags().GetString("source")
			serverType, _ := cmd.Flags().GetString("type")
			version, _ := cmd.Flags().GetString("version")

			a.logger.信息("下载服务端: %s %s %s", source, serverType, version)
			fmt.Printf("下载功能正在开发中...\n")
			return nil
		},
	}

	downloadCmd.Flags().StringP("source", "s", "fastmirror", "下载源 (fastmirror/mcsl)")
	downloadCmd.Flags().StringP("type", "t", "paper", "服务端类型")
	downloadCmd.Flags().StringP("version", "v", "", "Minecraft版本")
	downloadCmd.MarkFlagRequired("version")

	a.rootCmd.AddCommand(downloadCmd)
}

// addJavaCommands 添加Java命令
func (a *App) addJavaCommands() {
	javaCmd := &cobra.Command{
		Use:   "java",
		Short: "Java环境管理",
		Long:  "检测和管理Java运行环境",
	}

	// 检测Java命令
	detectCmd := &cobra.Command{
		Use:   "detect",
		Short: "检测Java环境",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.logger.信息("检测Java环境")
			fmt.Printf("Java检测功能正在开发中...\n")
			return nil
		},
	}

	// 列出Java命令
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出已检测的Java版本",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.logger.信息("列出Java版本")
			fmt.Printf("Java列表功能正在开发中...\n")
			return nil
		},
	}

	javaCmd.AddCommand(detectCmd, listCmd)
	a.rootCmd.AddCommand(javaCmd)
}

// addBackupCommands 添加备份命令
func (a *App) addBackupCommands() {
	backupCmd := &cobra.Command{
		Use:   "backup",
		Short: "备份管理",
		Long:  "创建、恢复和管理实例备份",
	}

	// 创建备份命令
	createCmd := &cobra.Command{
		Use:   "create [实例名称]",
		Short: "创建备份",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instance := args[0]
			a.logger.信息("创建备份: %s", instance)
			fmt.Printf("备份创建功能正在开发中...\n")
			return nil
		},
	}

	// 恢复备份命令
	restoreCmd := &cobra.Command{
		Use:   "restore",
		Short: "恢复备份",
		RunE: func(cmd *cobra.Command, args []string) error {
			instance, _ := cmd.Flags().GetString("instance")
			backup, _ := cmd.Flags().GetString("backup")

			a.logger.信息("恢复备份: %s -> %s", backup, instance)
			fmt.Printf("备份恢复功能正在开发中...\n")
			return nil
		},
	}
	restoreCmd.Flags().StringP("instance", "i", "", "实例名称")
	restoreCmd.Flags().StringP("backup", "b", "", "备份文件")
	restoreCmd.MarkFlagRequired("instance")
	restoreCmd.MarkFlagRequired("backup")

	backupCmd.AddCommand(createCmd, restoreCmd)
	a.rootCmd.AddCommand(backupCmd)
}

// addConfigCommands 添加配置命令
func (a *App) addConfigCommands() {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "配置管理",
		Long:  "查看和修改应用配置",
	}

	// 显示配置命令
	showCmd := &cobra.Command{
		Use:   "show",
		Short: "显示当前配置",
		RunE: func(cmd *cobra.Command, args []string) error {
			a.logger.信息("显示配置")
			fmt.Printf("配置显示功能正在开发中...\n")
			return nil
		},
	}

	configCmd.AddCommand(showCmd)
	a.rootCmd.AddCommand(configCmd)
}

// 占位符方法，后续实现具体功能

func (a *App) handleInstanceMenu() error {
	fmt.Println("实例管理功能正在开发中...")
	return nil
}

func (a *App) createBlankInstance() error {
	fmt.Println("创建空白实例功能正在开发中...")
	return nil
}

func (a *App) listInstances() error {
	fmt.Println("查看实例列表功能正在开发中...")
	return nil
}

func (a *App) manageInstanceState() error {
	fmt.Println("启动/停止实例功能正在开发中...")
	return nil
}

func (a *App) deleteInstance() error {
	fmt.Println("删除实例功能正在开发中...")
	return nil
}

func (a *App) handleDownloadMenu(source string) error {
	fmt.Printf("%s下载功能正在开发中...\n", source)
	return nil
}

func (a *App) detectJava() error {
	fmt.Println("Java环境检测功能正在开发中...")
	return nil
}

func (a *App) manageJava() error {
	fmt.Println("Java版本管理功能正在开发中...")
	return nil
}

func (a *App) configureFRP() error {
	fmt.Println("OpenFRP配置功能正在开发中...")
	return nil
}

func (a *App) manageTunnels() error {
	fmt.Println("隧道管理功能正在开发中...")
	return nil
}

func (a *App) startFTPServer() error {
	fmt.Println("FTP服务功能正在开发中...")
	return nil
}

func (a *App) fileBrowser() error {
	fmt.Println("文件浏览器功能正在开发中...")
	return nil
}

func (a *App) createBackup() error {
	fmt.Println("创建备份功能正在开发中...")
	return nil
}

func (a *App) restoreBackup() error {
	fmt.Println("恢复备份功能正在开发中...")
	return nil
}

func (a *App) manageBackups() error {
	fmt.Println("备份管理功能正在开发中...")
	return nil
}

func (a *App) configureDaemon() error {
	fmt.Println("进程守护配置功能正在开发中...")
	return nil
}

func (a *App) configureLogging() error {
	fmt.Println("日志设置功能正在开发中...")
	return nil
}

// 辅助函数

// readInput 读取用户输入
func (a *App) readInput(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return "", fmt.Errorf("读取输入失败")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

// readIntInput 读取整数输入
func (a *App) readIntInput(prompt string) (int, error) {
	input, err := a.readInput(prompt)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(input)
}

// confirmAction 确认操作
func (a *App) confirmAction(message string) bool {
	input, err := a.readInput(fmt.Sprintf("%s (y/N): ", message))
	if err != nil {
		return false
	}
	return strings.ToLower(input) == "y" || strings.ToLower(input) == "yes"
}
