package menu

import (
	"fmt"
	"os"

	"github.com/manifoldco/promptui"
)

// MenuItem 菜单项
type MenuItem struct {
	ID          string       // 菜单项ID
	Title       string       // 显示标题
	Description string       // 描述信息
	Handler     func() error // 处理函数
	SubMenu     *Menu        // 子菜单
	Enabled     func() bool  // 是否启用
	Status      func() string // 状态显示
}

// Menu 菜单结构
type Menu struct {
	Title       string     // 菜单标题
	Description string     // 菜单描述
	Items       []MenuItem // 菜单项列表
	Parent      *Menu      // 父菜单
}

// MenuSystem 菜单系统
type MenuSystem struct {
	currentMenu *Menu
}

// NewMenuSystem 创建新的菜单系统
func NewMenuSystem() *MenuSystem {
	return &MenuSystem{}
}

// SetRootMenu 设置根菜单
func (ms *MenuSystem) SetRootMenu(menu *Menu) {
	ms.currentMenu = menu
}

// Run 运行菜单系统
func (ms *MenuSystem) Run() {
	if ms.currentMenu == nil {
		fmt.Println("错误: 未设置根菜单")
		return
	}

	ms.showMenu(ms.currentMenu)
}

// showMenu 显示菜单并处理用户选择
func (ms *MenuSystem) showMenu(menu *Menu) {
	for {
		// 准备菜单选项
		items := ms.prepareMenuItems(menu)
		if len(items) == 0 {
			fmt.Println("没有可用的菜单项")
			return
		}

		// 创建选择提示
		prompt := promptui.Select{
			Label:     menu.Title,
			Items:     items,
			Templates: ms.getSelectTemplates(),
			Size:      10,
		}

		// 显示菜单并获取用户选择
		index, _, err := prompt.Run()
		if err != nil {
			if err == promptui.ErrInterrupt {
				fmt.Println("\n感谢使用 EasilyPanel5！")
				os.Exit(0)
			}
			fmt.Printf("选择失败: %v\n", err)
			continue
		}

		// 处理用户选择
		if err := ms.handleSelection(menu, index); err != nil {
			if err.Error() == "RETURN_TO_PARENT" {
				// 返回上级菜单
				return
			}
			fmt.Printf("操作失败: %v\n", err)
			ms.waitForEnter()
		}
	}
}

// prepareMenuItems 准备菜单项列表
func (ms *MenuSystem) prepareMenuItems(menu *Menu) []string {
	var items []string

	// 添加启用的菜单项
	for _, item := range menu.Items {
		if item.Enabled != nil && !item.Enabled() {
			continue // 跳过禁用的菜单项
		}

		display := item.Title
		if item.Description != "" {
			display += " - " + item.Description
		}
		if item.Status != nil {
			status := item.Status()
			if status != "" {
				display += " [" + status + "]"
			}
		}
		items = append(items, display)
	}

	// 添加返回选项（如果有父菜单）
	if menu.Parent != nil {
		items = append(items, "← 返回上级菜单")
	}

	// 添加退出选项
	items = append(items, "✗ 退出程序")

	return items
}

// getSelectTemplates 获取选择模板
func (ms *MenuSystem) getSelectTemplates() *promptui.SelectTemplates {
	return &promptui.SelectTemplates{
		Label:    "{{ . }}:",
		Active:   "▶ {{ . | cyan }}",
		Inactive: "  {{ . }}",
		Selected: "✓ {{ . | green }}",
	}
}

// handleSelection 处理用户选择
func (ms *MenuSystem) handleSelection(menu *Menu, index int) error {
	// 计算实际的菜单项索引（排除禁用项）
	enabledItems := ms.getEnabledItems(menu)
	totalEnabledItems := len(enabledItems)

	// 检查是否是返回选项
	hasParent := menu.Parent != nil
	returnIndex := totalEnabledItems
	if hasParent && index == returnIndex {
		// 返回上级菜单，结束当前菜单循环
		return fmt.Errorf("RETURN_TO_PARENT")
	}

	// 检查是否是退出选项
	exitIndex := totalEnabledItems
	if hasParent {
		exitIndex++
	}
	if index == exitIndex {
		fmt.Println("\n感谢使用 EasilyPanel5！")
		os.Exit(0)
	}

	// 处理普通菜单项
	if index < totalEnabledItems {
		selectedItem := enabledItems[index]

		if selectedItem.SubMenu != nil {
			// 进入子菜单
			selectedItem.SubMenu.Parent = menu
			ms.showMenu(selectedItem.SubMenu)
			return nil
		} else if selectedItem.Handler != nil {
			// 执行处理函数
			return selectedItem.Handler()
		} else {
			return fmt.Errorf("该功能尚未实现")
		}
	}

	return fmt.Errorf("无效选择")
}

// getEnabledItems 获取启用的菜单项
func (ms *MenuSystem) getEnabledItems(menu *Menu) []MenuItem {
	var enabledItems []MenuItem
	for _, item := range menu.Items {
		if item.Enabled == nil || item.Enabled() {
			enabledItems = append(enabledItems, item)
		}
	}
	return enabledItems
}

// waitForEnter 等待用户按回车
func (ms *MenuSystem) waitForEnter() {
	fmt.Print("\n按回车键继续...")
	fmt.Scanln()
}

// NewMenuItem 创建新菜单项
func NewMenuItem(id, title, description string) *MenuItem {
	return &MenuItem{
		ID:          id,
		Title:       title,
		Description: description,
	}
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
