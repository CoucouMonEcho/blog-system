package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/stat/api"
	"blog-system/services/stat/infrastructure"
)

func main() {
	var configPath string
	if _, err := os.Stat("/opt/blog-system/configs/stat.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/stat.yaml"
	} else if _, err := os.Stat("../../configs/stat.yaml"); err == nil {
		configPath = "../../configs/stat.yaml"
	} else {
		configPath = "configs/stat.yaml"
	}
	cfg, err := infrastructure.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	lgr, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	if _, err := infrastructure.InitDB(cfg); err != nil {
		lgr.LogWithContext("stat-service", "database", "ERROR", "数据库连接失败: %v", err)
	}
	http := api.NewHTTPServer(lgr)
	addr := ":" + strconv.Itoa(cfg.App.Port)
	if err := http.Run(addr); err != nil {
		lgr.LogWithContext("stat-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
