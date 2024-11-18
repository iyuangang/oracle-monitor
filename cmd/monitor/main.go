package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"oracle-monitor/internal/config"
	"oracle-monitor/internal/monitor"
)

func main() {
    // 命令行参数
    configFile := flag.String("config", "configs/config.json", "path to config file")
    flag.Parse()

    // 加载配置
    cfg, err := config.LoadConfig(*configFile)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    // 创建监控服务
    service := monitor.NewService(cfg)

    // 创建上下文，用于优雅退出
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 处理系统信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    // 启动监控服务
    go func() {
        if err := service.Start(ctx); err != nil {
            log.Printf("Service error: %v", err)
            cancel()
        }
    }()

    // 等待退出信号
    <-sigChan
    log.Println("Shutting down...")
    cancel()
    service.Shutdown()
}
