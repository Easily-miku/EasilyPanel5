package main

import (
	"fmt"
	"log"
	"os"

	"easilypanel/internal/frp"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run test_frp.go <your_token>")
		return
	}

	token := os.Args[1]
	
	// 创建OpenFRP客户端
	client := frp.NewOpenFRPClient("", token)
	
	fmt.Println("测试OpenFRP API连接...")
	fmt.Printf("使用Token: %s...%s\n", token[:8], token[len(token)-4:])
	
	// 测试获取用户信息
	fmt.Println("\n1. 测试获取用户信息...")
	userInfo, err := client.GetUserInfo()
	if err != nil {
		log.Printf("获取用户信息失败: %v", err)
	} else {
		fmt.Printf("✓ 用户信息获取成功:")
		fmt.Printf("  用户名: %s\n", userInfo.Username)
		fmt.Printf("  用户组: %s\n", userInfo.FriendlyGroup)
		fmt.Printf("  隧道配额: %d/%d\n", userInfo.Used, userInfo.Proxies)
		fmt.Printf("  剩余流量: %d MB\n", userInfo.Traffic)
	}
	
	// 测试获取隧道列表
	fmt.Println("\n2. 测试获取隧道列表...")
	proxies, err := client.GetProxies()
	if err != nil {
		log.Printf("获取隧道列表失败: %v", err)
	} else {
		fmt.Printf("✓ 隧道列表获取成功: %d个隧道\n", len(proxies.List))
		for i, proxy := range proxies.List {
			if i >= 3 { // 只显示前3个
				fmt.Printf("  ... 还有 %d 个隧道\n", len(proxies.List)-3)
				break
			}
			status := "离线"
			if proxy.Status {
				status = "在线"
			}
			fmt.Printf("  - %s (%s) - %s\n", proxy.ProxyName, proxy.ProxyType, status)
		}
	}
	
	// 测试获取节点列表
	fmt.Println("\n3. 测试获取节点列表...")
	nodes, err := client.GetNodes()
	if err != nil {
		log.Printf("获取节点列表失败: %v", err)
	} else {
		fmt.Printf("✓ 节点列表获取成功: %d个节点\n", len(nodes.List))
		onlineCount := 0
		for _, node := range nodes.List {
			if node.Status == 1 {
				onlineCount++
			}
		}
		fmt.Printf("  在线节点: %d/%d\n", onlineCount, len(nodes.List))
	}
	
	fmt.Println("\n测试完成！")
}
