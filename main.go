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
	"easilypanel5/config"
	"easilypanel5/core"
)

const (
	DefaultPort = "8080"
	AppName     = "EasilyPanel5"
	Version     = "1.0.1"
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

	// 创建必要的目录
	if err := createDirectories(); err != nil {
		log.Fatalf("Failed to create directories: %v", err)
	}

	// 设置路由
	router := api.SetupRoutes()

	// 获取端口
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}

	// 创建HTTP服务器
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 启动服务器
	go func() {
		fmt.Printf("Server starting on http://localhost:%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
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
