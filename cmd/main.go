package main

import (
	"log"
	"github.com/hushuaishuai1/go-backend/internal/api"
	"github.com/hushuaishuai1/go-backend/internal/config"
	"github.com/hushuaishuai1/go-backend/internal/services"
)

func main() {
	// 加载配置 (未来可以从环境变量加载)
	cfg := config.LoadConfig()

	// 初始化数据库和服务
	alertManager, err := services.NewAlertManager("alerts.db")
	if err != nil {
		log.Fatalf("Failed to initialize alert manager: %v", err)
	}

	mailer := services.NewMailer(cfg.SMTP)

	// 启动Binance WebSocket监控
	go services.StartBinanceWebSocket(alertManager, mailer)

	// 启动API服务器
	router := api.SetupRouter(alertManager)
	log.Println("Starting API server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}