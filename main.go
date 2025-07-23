package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"easilypanel5/api"
	"easilypanel5/auth"
	"easilypanel5/config"
	"easilypanel5/core"
	"easilypanel5/frp"
	"easilypanel5/server"
)

const (
	DefaultPort = "8080"
	AppName     = "EasilyPanel5"
	Version     = "1.1.0"
)

func main() {
	fmt.Printf("%s v%s - Minecraft Server Manager\n", AppName, Version)
	fmt.Println("Starting server...")

	// 初始化配置
	if err := config.Initialize(); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 初始化服务器配置
	if err := config.InitServers(); err != nil {
		log.Fatalf("Failed to initialize servers: %v", err)
	}

	// 初始化下载管理器
	core.InitDownloadManager()

	// 初始化认证系统
	authService := auth.NewAuthService("data/auth")
	authHandlers := auth.NewAuthHandlers(authService)

	// 启动进程守护管理器
	daemon := server.GetDaemon()
	if err := daemon.Start(); err != nil {
		log.Printf("Failed to start daemon: %v", err)
	}

	// 启动FRP管理器（如果启用）
	cfg := config.Get()
	if cfg.FRP.Enabled {
		if err := frp.StartManager(); err != nil {
			log.Printf("Failed to start FRP manager: %v", err)
		}
	}

	// 创建必要的目录
	if err := createDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// 设置路由
	router := api.SetupRoutes(authHandlers)

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// 创建HTTP服务器
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		fmt.Printf("Server starting on http://localhost:%s\n", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")

	// 停止FRP管理器
	if err := frp.StopManager(); err != nil {
		log.Printf("Failed to stop FRP manager: %v", err)
	}

	// 停止守护管理器
	daemon = server.GetDaemon()
	if err := daemon.Stop(); err != nil {
		log.Printf("Failed to stop daemon: %v", err)
	}

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}

// createDirectories 创建必要的目录
func createDirectories() error {
	dirs := []string{
		"data",
		"data/servers",
		"data/logs",
		"data/cores",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}

	return nil
}
