package main

import (
	"log"
	"github.com/hushuaishuai1/go-backend/internal/api"
	"github.com/hushuaishuai1/go-backend/internal/config"
	"github.com/hushuaishuai1/go-backend/internal/services"
)

func main() {
	// 加载配置 (无变化)
	cfg := config.LoadConfig()

	// 初始化数据库和服务 (无变化)
	// 我们为数据库文件指定一个在容器内持久化的路径
	alertManager, err := services.NewAlertManager("./data/alerts.db") 
	if err != nil {
		log.Fatalf("Failed to initialize alert manager: %v", err)
	}

	mailer := services.NewMailer(cfg.SMTP)

	// 启动Binance WebSocket监控 (无变化)
	go services.StartBinanceWebSocket(alertManager, mailer)

	// --- 启动API服务器 (已从Gin改为Echo) ---
	// 1. 设置Echo路由
	e := api.SetupRouter(alertManager)
	
	// 2. 启动Echo服务器
	log.Println("Starting Echo API server on :8080")
	//    Echo使用 e.Start() 方法来启动服务
	if err := e.Start(":8080"); err != nil {
		// Echo的启动错误需要从 e.Start() 的返回值获取
		log.Fatalf("Failed to start server: %v", err)
	}
}