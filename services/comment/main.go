package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/comment/api"
	"blog-system/services/comment/application"
	"blog-system/services/comment/infrastructure"
)

func main() {
	var configPath string
	if _, err := os.Stat("/opt/blog-system/configs/comment.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/comment.yaml"
	} else if _, err := os.Stat("../../configs/comment.yaml"); err == nil {
		configPath = "../../configs/comment.yaml"
	} else {
		configPath = "configs/comment.yaml"
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
		lgr.LogWithContext("comment-service", "database", "FATAL", "数据库连接失败: %v", err)
		return
	}
	if err := infrastructure.RegisterService(cfg); err != nil {
		lgr.LogWithContext("comment-service", "registry", "ERROR", "注册到注册中心失败: %v", err)
	}
	repo := infrastructure.NewCommentRepository(db)
	app := application.NewCommentService(repo, lgr)
	http := api.NewHTTPServer(app, lgr)
	addr := ":" + strconv.Itoa(cfg.App.Port)
	if err := http.Run(addr); err != nil {
		lgr.LogWithContext("comment-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
