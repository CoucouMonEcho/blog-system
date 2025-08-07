package main

import (
	"os"
	"path/filepath"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/user/api"
	"blog-system/services/user/application"
	"blog-system/services/user/infrastructure"
)

func main() {
	// 加载配置
	configPath := filepath.Join("../../..", "configs", "user.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 使用默认配置路径
		configPath = "configs/user.yaml"
	}
	cfg, err := infrastructure.LoadConfig(configPath)
	if err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 初始化日志系统
	loggerInstance, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		panic("初始化日志系统失败: " + err.Error())
	}

	// 记录服务启动日志
	loggerInstance.LogWithContext("user-service", "main", "INFO", "开始启动用户服务")

	// 初始化数据库连接
	db, err := infrastructure.InitDB(cfg)
	if err != nil {
		loggerInstance.LogWithContext("user-service", "database", "FATAL", "数据库连接失败: %v", err)
	}

	loggerInstance.LogWithContext("user-service", "database", "INFO", "数据库连接成功")

	// 初始化缓存
	cache, err := infrastructure.InitCache(cfg)
	if err != nil {
		loggerInstance.LogWithContext("user-service", "cache", "FATAL", "缓存连接失败: %v", err)
	}

	loggerInstance.LogWithContext("user-service", "cache", "INFO", "缓存连接成功")

	// 初始化仓储层
	userRepo := infrastructure.NewUserRepository(db)
	loggerInstance.LogWithContext("user-service", "repository", "INFO", "用户仓储层初始化完成")

	// 初始化应用服务
	userService := application.NewUserService(userRepo, cache)
	loggerInstance.LogWithContext("user-service", "application", "INFO", "用户应用服务初始化完成")

	// 启动 HTTP 服务
	server := api.NewHTTPServer(userService)
	loggerInstance.LogWithContext("user-service", "api", "INFO", "HTTP服务器初始化完成")

	addr := ":" + strconv.Itoa(cfg.App.Port)
	loggerInstance.LogWithContext("user-service", "main", "INFO", "用户服务启动中，监听端口: %s", addr)

	if err := server.Run(addr); err != nil {
		loggerInstance.LogWithContext("user-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
