package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/api"
	"blog-system/services/gateway/application"
	"blog-system/services/gateway/domain"
	"blog-system/services/gateway/infrastructure"
)

func main() {
	// 加载配置 - 修复路径问题
	var configPath string

	// 优先使用绝对路径（部署环境）
	if _, err := os.Stat("/opt/blog-system/configs/gateway.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/gateway.yaml"
	} else if _, err := os.Stat("../../configs/gateway.yaml"); err == nil {
		// 开发环境
		configPath = "../../configs/gateway.yaml"
	} else {
		// fallback到相对路径
		configPath = "configs/gateway.yaml"
	}

	log.Printf("使用配置文件: %s", configPath)

	cfg, err := infrastructure.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志系统
	loggerInstance, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	// 记录服务启动日志
	loggerInstance.LogWithContext("gateway-service", "main", "INFO", "开始启动网关服务")
	loggerInstance.LogWithContext("gateway-service", "main", "INFO", "配置文件: %s", configPath)

	// 初始化缓存连接 - 添加错误处理但不退出
	cache, err := infrastructure.InitCache(cfg)
	if err != nil {
		loggerInstance.LogWithContext("gateway-service", "cache", "ERROR", "缓存连接失败: %v", err)
		loggerInstance.LogWithContext("gateway-service", "cache", "WARN", "缓存连接失败，但继续启动服务")
	} else {
		loggerInstance.LogWithContext("gateway-service", "cache", "INFO", "缓存连接成功")
	}

	// 初始化路由仓储
	routeRepo := infrastructure.NewRouteRepository(cfg)
	loggerInstance.LogWithContext("gateway-service", "repository", "INFO", "路由仓储层初始化完成")

	// 初始化服务发现
	serviceDiscovery := infrastructure.NewServiceDiscovery()
	loggerInstance.LogWithContext("gateway-service", "discovery", "INFO", "服务发现初始化完成")

	// 初始化限流器
	var rateLimiter domain.RateLimiter
	if cache != nil && cfg.RateLimit.Enabled {
		rateLimiter = infrastructure.NewRateLimiter(cache, cfg.RateLimit)
		loggerInstance.LogWithContext("gateway-service", "ratelimit", "INFO", "限流器初始化完成")
	} else {
		loggerInstance.LogWithContext("gateway-service", "ratelimit", "INFO", "限流器未启用")
	}

	// 初始化熔断器
	var circuitBreaker domain.CircuitBreaker
	if cfg.CircuitBreaker.Enabled {
		circuitBreaker = infrastructure.NewCircuitBreaker(cfg.CircuitBreaker)
		loggerInstance.LogWithContext("gateway-service", "circuitbreaker", "INFO", "熔断器初始化完成")
	} else {
		loggerInstance.LogWithContext("gateway-service", "circuitbreaker", "INFO", "熔断器未启用")
	}

	// 初始化应用服务
	gatewayService := application.NewGatewayService(routeRepo, serviceDiscovery, rateLimiter, circuitBreaker)
	loggerInstance.LogWithContext("gateway-service", "application", "INFO", "网关应用服务初始化完成")

	// 启动 HTTP 服务
	server := api.NewHTTPServer(gatewayService, loggerInstance)
	loggerInstance.LogWithContext("gateway-service", "api", "INFO", "HTTP服务器初始化完成")

	addr := ":" + strconv.Itoa(cfg.App.Port)
	loggerInstance.LogWithContext("gateway-service", "main", "INFO", "网关服务启动中，监听端口: %s", addr)

	if err := server.Run(addr); err != nil {
		loggerInstance.LogWithContext("gateway-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
