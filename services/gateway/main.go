package main

import (
	"log"
	"strconv"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/api"
	"blog-system/services/gateway/application"
	"blog-system/services/gateway/domain"
	"blog-system/services/gateway/infrastructure"
)

func main() {
	cfg, err := conf.Load("gateway")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志系统
	loggerInstance, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志系统失败: %v", err)
	}

	// 初始化全局 Logger
	logger.Init(loggerInstance)

	// 记录服务启动日志
	logger.L().LogWithContext("gateway-service", "main", "INFO", "开始启动网关服务")

	// 初始化缓存连接 - 添加错误处理但不退出
	cache, err := infrastructure.InitCache(cfg)
	if err != nil {
		logger.L().LogWithContext("gateway-service", "cache", "ERROR", "缓存连接失败: %v", err)
		logger.L().LogWithContext("gateway-service", "cache", "WARN", "缓存连接失败，但继续启动服务")
	} else {
		logger.L().LogWithContext("gateway-service", "cache", "INFO", "缓存连接成功")
	}

	// 初始化路由仓储
	routeRepo := infrastructure.NewRouteRepository(cfg)
	logger.L().LogWithContext("gateway-service", "repository", "INFO", "路由仓储层初始化完成")

	// 初始化服务发现
	serviceDiscovery := infrastructure.NewServiceDiscovery()
	logger.L().LogWithContext("gateway-service", "discovery", "INFO", "服务发现初始化完成")

	// 初始化限流器
	var rateLimiter domain.RateLimiter
	if cache != nil && cfg.RateLimit.Enabled {
		rateLimiter = infrastructure.NewRateLimiter(cache, cfg.RateLimit)
		logger.L().LogWithContext("gateway-service", "ratelimit", "INFO", "限流器初始化完成")
	} else {
		logger.L().LogWithContext("gateway-service", "ratelimit", "INFO", "限流器未启用")
	}

	// 初始化熔断器
	var circuitBreaker domain.CircuitBreaker
	if cfg.CircuitBreaker.Enabled {
		circuitBreaker = infrastructure.NewCircuitBreaker(cfg.CircuitBreaker)
		logger.L().LogWithContext("gateway-service", "circuitbreaker", "INFO", "熔断器初始化完成")
	} else {
		logger.L().LogWithContext("gateway-service", "circuitbreaker", "INFO", "熔断器未启用")
	}

	// 初始化应用服务
	gatewayService := application.NewGatewayService(routeRepo, serviceDiscovery, rateLimiter, circuitBreaker, logger.L())
	logger.L().LogWithContext("gateway-service", "application", "INFO", "网关应用服务初始化完成")

	// 启动 HTTP 服务
	server := api.NewHTTPServer(gatewayService)
	logger.L().LogWithContext("gateway-service", "api", "INFO", "HTTP服务器初始化完成")

	addr := ":" + strconv.Itoa(cfg.App.Port)
	logger.L().LogWithContext("gateway-service", "main", "INFO", "网关服务启动中，监听端口: %s", addr)

	if err := server.Run(addr); err != nil {
		logger.L().LogWithContext("gateway-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
