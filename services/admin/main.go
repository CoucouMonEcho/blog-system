package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/admin/api"
)

func main() {
	var port = 8005
	if v := os.Getenv("ADMIN_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}
	lgr, err := logger.NewLogger(&logger.Config{Level: "info", Path: "logs/admin-service.log"})
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	http := api.NewHTTPServer(lgr)
	addr := ":" + strconv.Itoa(port)
	if err := http.Run(addr); err != nil {
		lgr.LogWithContext("admin-service", "main", "FATAL", "服务启动失败: %v", err)
	}

}
