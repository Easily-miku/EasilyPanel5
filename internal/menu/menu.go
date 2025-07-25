package menu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ANSI颜色代码
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
	ColorDim    = "\033[2m"
)

// 菜单配置
const (
	MenuWidth = 70
	MenuPadding = 2
)

// MenuItem 菜单项
type MenuItem struct {
	ID          string                    // 菜单项ID
	Title       string                    // 显示标题
	Description string                    // 描述信息
	Icon        string                    // 图标
	Handler     func() error              // 处理函数
	SubMenu     *Menu                     // 子菜单
	Enabled     func() bool               // 是否启用
	Status      func() string             // 状态显示
}

// Menu 菜单结构
type Menu struct {
	Title       string      // 菜单标题
	Description string      // 菜单描述
	Items       []MenuItem  // 菜单项列表
	Parent      *Menu       // 父菜单
	Breadcrumb  []string    // 面包屑导航
}

// MenuSystem 菜单系统
type MenuSystem struct {
	currentMenu *Menu
	scanner     *bufio.Scanner
}

// NewMenuSystem 创建新的菜单系统
func NewMenuSystem() *MenuSystem {
	return &MenuSystem{
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// SetRootMenu 设置根菜单
func (ms *MenuSystem) SetRootMenu(menu *Menu) {
	ms.currentMenu = menu
	menu.Breadcrumb = []string{menu.Title}
}

// Show 显示当前菜单
func (ms *MenuSystem) Show() {
	if ms.currentMenu == nil {
		fmt.Println("错误: 未设置菜单")
		return
	}
	
	ms.displayMenu()
}

// displayMenu 显示菜单
func (ms *MenuSystem) displayMenu() {
	// 清屏
	fmt.Print("\033[2J\033[H")
	
	// 显示标题
	ms.displayHeader()
	
	// 显示面包屑导航
	ms.displayBreadcrumb()
	
	// 显示菜单项
	ms.displayMenuItems()
	
	// 显示底部信息
	ms.displayFooter()
	
	// 处理用户输入
	ms.handleInput()
}

// displayHeader 显示标题
func (ms *MenuSystem) displayHeader() {
	title := ms.currentMenu.Title

	fmt.Println(ColorCyan + "+" + strings.Repeat("-", MenuWidth-2) + "+" + ColorReset)
	fmt.Printf(ColorCyan + "|" + ColorBold + "%s" + ColorReset + ColorCyan + "|\n" + ColorReset, ms.centerText(title, MenuWidth-2))

	if ms.currentMenu.Description != "" {
		fmt.Printf(ColorCyan + "|" + ColorDim + "%s" + ColorReset + ColorCyan + "|\n" + ColorReset, ms.centerText(ms.currentMenu.Description, MenuWidth-2))
	}

	fmt.Println(ColorCyan + "+" + strings.Repeat("-", MenuWidth-2) + "+" + ColorReset)
}

// displayBreadcrumb 显示面包屑导航
func (ms *MenuSystem) displayBreadcrumb() {
	if len(ms.currentMenu.Breadcrumb) > 1 {
		breadcrumbText := "位置: " + strings.Join(ms.currentMenu.Breadcrumb, " > ")
		fmt.Printf(ColorCyan + "| " + ColorYellow + "%-*s" + ColorCyan + " |\n" + ColorReset, MenuWidth-4, breadcrumbText)
		fmt.Println(ColorCyan + "+" + strings.Repeat("-", MenuWidth-2) + "+" + ColorReset)
	}
}

// displayMenuItems 显示菜单项
func (ms *MenuSystem) displayMenuItems() {
	for i, item := range ms.currentMenu.Items {
		if item.Enabled != nil && !item.Enabled() {
			continue // 跳过禁用的菜单项
		}

		// 菜单项编号
		num := fmt.Sprintf("%d", i+1)

		// 标题
		title := item.Title

		// 状态
		status := ""
		if item.Status != nil {
			status = item.Status()
		}

		// 格式化显示
		line := fmt.Sprintf(" %s%s%s. %s%s%s", ColorGreen, num, ColorReset, ColorBold, title, ColorReset)
		if status != "" {
			line += fmt.Sprintf(" %s[%s]%s", ColorYellow, status, ColorReset)
		}

		// 计算实际显示长度（去除颜色代码）
		plainLine := fmt.Sprintf(" %s. %s", num, title)
		if status != "" {
			plainLine += fmt.Sprintf(" [%s]", status)
		}

		// 使用plainLine计算长度，但显示带颜色的line
		padding := MenuWidth - 4 - len(plainLine)
		if padding < 0 {
			padding = 0
		}
		fmt.Printf(ColorCyan + "|" + ColorReset + "%s%s" + ColorCyan + "|\n" + ColorReset, line, strings.Repeat(" ", padding))

		// 描述信息
		if item.Description != "" {
			desc := fmt.Sprintf("    %s", item.Description)
			fmt.Printf(ColorCyan + "|" + ColorDim + " %-*s " + ColorReset + ColorCyan + "|\n" + ColorReset, MenuWidth-4, desc)
		}

		fmt.Printf(ColorCyan + "|%-*s|\n" + ColorReset, MenuWidth-2, "")
	}

	// 返回选项
	if ms.currentMenu.Parent != nil {
		returnText := fmt.Sprintf(" %s0%s. %s返回上级菜单%s", ColorGreen, ColorReset, ColorBold, ColorReset)
		returnPlain := " 0. 返回上级菜单"
		padding := MenuWidth - 4 - len(returnPlain)
		if padding < 0 {
			padding = 0
		}
		fmt.Printf(ColorCyan + "|" + ColorReset + "%s%s" + ColorCyan + "|\n" + ColorReset, returnText, strings.Repeat(" ", padding))
		fmt.Printf(ColorCyan + "|%-*s|\n" + ColorReset, MenuWidth-2, "")
	}

	// 退出选项
	exitText := fmt.Sprintf(" %sq%s. %s退出程序%s", ColorRed, ColorReset, ColorBold, ColorReset)
	exitPlain := " q. 退出程序"
	padding := MenuWidth - 4 - len(exitPlain)
	if padding < 0 {
		padding = 0
	}
	fmt.Printf(ColorCyan + "|" + ColorReset + "%s%s" + ColorCyan + "|\n" + ColorReset, exitText, strings.Repeat(" ", padding))
}

// displayFooter 显示底部信息
func (ms *MenuSystem) displayFooter() {
	fmt.Println(ColorCyan + "+" + strings.Repeat("-", MenuWidth-2) + "+" + ColorReset)
	fmt.Print(ColorBold + "请选择操作 (输入数字或字母): " + ColorReset)
}

// handleInput 处理用户输入
func (ms *MenuSystem) handleInput() {
	if !ms.scanner.Scan() {
		return
	}
	
	input := strings.TrimSpace(ms.scanner.Text())
	
	switch input {
	case "q", "Q", "quit", "exit":
		fmt.Println("感谢使用 EasilyPanel5！")
		os.Exit(0)
	case "0":
		if ms.currentMenu.Parent != nil {
			ms.navigateToParent()
		} else {
			fmt.Println("已在根菜单")
			ms.waitForEnter()
		}
	default:
		ms.handleMenuSelection(input)
	}
}

// handleMenuSelection 处理菜单选择
func (ms *MenuSystem) handleMenuSelection(input string) {
	// 尝试解析为数字
	num, err := strconv.Atoi(input)
	if err != nil {
		fmt.Printf("无效输入: %s\n", input)
		ms.waitForEnter()
		ms.displayMenu()
		return
	}
	
	// 检查范围
	if num < 1 || num > len(ms.currentMenu.Items) {
		fmt.Printf("无效选择: %d\n", num)
		ms.waitForEnter()
		ms.displayMenu()
		return
	}
	
	// 获取选中的菜单项
	selectedItem := ms.currentMenu.Items[num-1]
	
	// 检查是否启用
	if selectedItem.Enabled != nil && !selectedItem.Enabled() {
		fmt.Println("该功能当前不可用")
		ms.waitForEnter()
		ms.displayMenu()
		return
	}
	
	// 执行操作
	if selectedItem.SubMenu != nil {
		// 进入子菜单
		ms.navigateToSubMenu(&selectedItem)
	} else if selectedItem.Handler != nil {
		// 执行处理函数
		err := selectedItem.Handler()
		if err != nil {
			fmt.Printf("操作失败: %v\n", err)
		}
		ms.waitForEnter()
		ms.displayMenu()
	} else {
		fmt.Println("该功能尚未实现")
		ms.waitForEnter()
		ms.displayMenu()
	}
}

// navigateToSubMenu 导航到子菜单
func (ms *MenuSystem) navigateToSubMenu(item *MenuItem) {
	subMenu := item.SubMenu
	subMenu.Parent = ms.currentMenu
	
	// 更新面包屑
	subMenu.Breadcrumb = make([]string, len(ms.currentMenu.Breadcrumb))
	copy(subMenu.Breadcrumb, ms.currentMenu.Breadcrumb)
	subMenu.Breadcrumb = append(subMenu.Breadcrumb, item.Title)
	
	ms.currentMenu = subMenu
	ms.displayMenu()
}

// navigateToParent 导航到父菜单
func (ms *MenuSystem) navigateToParent() {
	if ms.currentMenu.Parent != nil {
		ms.currentMenu = ms.currentMenu.Parent
		ms.displayMenu()
	}
}

// waitForEnter 等待用户按回车
func (ms *MenuSystem) waitForEnter() {
	fmt.Print("\n按回车键继续...")
	ms.scanner.Scan()
}

// centerText 居中文本
func (ms *MenuSystem) centerText(text string, width int) string {
	if len(text) >= width {
		return text[:width]
	}

	padding := (width - len(text)) / 2
	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", width-len(text)-padding)
}

// stripColors 移除ANSI颜色代码
func (ms *MenuSystem) stripColors(text string) string {
	// 简单的颜色代码移除
	result := text
	colorCodes := []string{
		ColorReset, ColorRed, ColorGreen, ColorYellow,
		ColorBlue, ColorPurple, ColorCyan, ColorWhite,
		ColorBold, ColorDim,
	}

	for _, code := range colorCodes {
		result = strings.ReplaceAll(result, code, "")
	}

	return result
}

// Run 运行菜单系统
func (ms *MenuSystem) Run() {
	if ms.currentMenu == nil {
		fmt.Println("错误: 未设置根菜单")
		return
	}
	
	ms.displayMenu()
}

// NewMenuItem 创建新菜单项
func NewMenuItem(id, title, description string) *MenuItem {
	return &MenuItem{
		ID:          id,
		Title:       title,
		Description: description,
	}
}

// WithIcon 设置图标
func (item *MenuItem) WithIcon(icon string) *MenuItem {
	item.Icon = icon
	return item
}

// WithHandler 设置处理函数
func (item *MenuItem) WithHandler(handler func() error) *MenuItem {
	item.Handler = handler
	return item
}

// WithSubMenu 设置子菜单
func (item *MenuItem) WithSubMenu(subMenu *Menu) *MenuItem {
	item.SubMenu = subMenu
	return item
}

// WithEnabled 设置启用条件
func (item *MenuItem) WithEnabled(enabled func() bool) *MenuItem {
	item.Enabled = enabled
	return item
}

// WithStatus 设置状态显示
func (item *MenuItem) WithStatus(status func() string) *MenuItem {
	item.Status = status
	return item
}

// NewMenu 创建新菜单
func NewMenu(title, description string) *Menu {
	return &Menu{
		Title:       title,
		Description: description,
		Items:       make([]MenuItem, 0),
	}
}

// AddItem 添加菜单项
func (menu *Menu) AddItem(item *MenuItem) *Menu {
	menu.Items = append(menu.Items, *item)
	return menu
}

// AddItems 批量添加菜单项
func (menu *Menu) AddItems(items ...*MenuItem) *Menu {
	for _, item := range items {
		menu.Items = append(menu.Items, *item)
	}
	return menu
}
