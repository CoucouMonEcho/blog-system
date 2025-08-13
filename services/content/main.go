package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/content/api"
	"blog-system/services/content/application"
	"blog-system/services/content/infrastructure"
)

func main() {
	var configPath string
	if _, err := os.Stat("/opt/blog-system/configs/content.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/content.yaml"
	} else if _, err := os.Stat("../../configs/content.yaml"); err == nil {
		configPath = "../../configs/content.yaml"
	} else {
		configPath = "configs/content.yaml"
	}

	cfg, err := infrastructure.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	lgr, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	db, err := infrastructure.InitDB(cfg)
	if err != nil {
		lgr.LogWithContext("content-service", "database", "FATAL", "数据库连接失败: %v", err)
		return
	}
	cache, _ := infrastructure.InitCache(cfg)
	_ = cache

	repo := infrastructure.NewContentRepository(db)
	app := application.NewContentService(repo, lgr)

	http := api.NewHTTPServer(app, lgr)
	// 注册到注册中心
	if err := infrastructure.RegisterService(cfg); err != nil {
		lgr.LogWithContext("content-service", "registry", "ERROR", "注册到注册中心失败: %v", err)
	}
	addr := ":" + strconv.Itoa(cfg.App.Port)
	if err := http.Run(addr); err != nil {
		lgr.LogWithContext("content-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
