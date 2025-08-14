package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/admin/api"
	"blog-system/services/admin/application"
	"blog-system/services/admin/infrastructure"
)

func main() {
	var configPath string
	if _, err := os.Stat("/opt/blog-system/configs/admin.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/admin.yaml"
	} else if _, err := os.Stat("../../configs/admin.yaml"); err == nil {
		configPath = "../../configs/admin.yaml"
	} else {
		configPath = "configs/admin.yaml"
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
		log.Printf("数据库连接失败: %v", err)
		lgr.LogWithContext("admin-service", "database", "FATAL", "数据库连接失败: %v", err)
		return
	}
	userRepo := infrastructure.NewUserRepository(db)
	artRepo := infrastructure.NewArticleRepository(db)
	catRepo := infrastructure.NewCategoryRepository(db)
	app := application.NewAdminService(userRepo, artRepo, catRepo, lgr, nil)

	http := api.NewHTTPServer(lgr)
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
