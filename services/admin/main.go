package main

import (
	"log"
	"strconv"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/admin/application"
	"blog-system/services/admin/infrastructure"
	"blog-system/services/admin/infrastructure/clients"
	httpapi "blog-system/services/admin/interfaces/httpserver"
)

func main() {
	cfg, err := conf.Load("admin")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	lgr, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化全局 Logger
	logger.Init(lgr)

	userCli := clients.NewUserServiceClient(cfg)
	contentCli := clients.NewContentClient(cfg)
	statCli := infrastructure.NewStatServiceClient(cfg)
	promCli := infrastructure.NewPrometheusClient("", 0)
	app := application.NewAdminService(userCli, contentCli, lgr, nil, statCli, promCli)

	http := httpapi.NewHTTPServer()
	http.SetApp(app)
	// 注册服务发现（失败不阻断启动）
	if err := infrastructure.RegisterService(cfg); err != nil {
		log.Printf("注册中心失败: %v (忽略继续)", err)
		lgr.LogWithContext("admin-service", "registry", "ERROR", "注册到注册中心失败: %v", err)
	}
	addr := ":" + strconv.Itoa(cfg.App.Port)
	if err := http.Run(addr); err != nil {
		log.Printf("服务启动失败: %v", err)
		lgr.LogWithContext("admin-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
