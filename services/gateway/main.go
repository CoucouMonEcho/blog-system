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

	// 初始化全局 Logger
	logger.Init(&cfg.Log)

	// 记录服务启动日志
	logger.Log().Info("main: 开始启动网关服务")

	// 初始化缓存连接 - 添加错误处理但不退出
	cache, err := infrastructure.InitCache(cfg)
	if err != nil {
		logger.Log().Error("cache: 缓存连接失败: %v", err)
		logger.Log().Warn("cache: 缓存连接失败，但继续启动服务")
	} else {
		logger.Log().Info("cache: 缓存连接成功")
	}

	// 初始化路由仓储
	routeRepo := infrastructure.NewRouteRepository(cfg)
	logger.Log().Info("repository: 路由仓储层初始化完成")

	// 初始化服务发现
	serviceDiscovery := infrastructure.NewServiceDiscovery()
	logger.Log().Info("discovery: 服务发现初始化完成")

	// 初始化限流器
	var rateLimiter domain.RateLimiter
	if cache != nil && cfg.RateLimit.Enabled {
		rateLimiter = infrastructure.NewRateLimiter(cache, cfg.RateLimit)
		logger.Log().Info("ratelimit: 限流器初始化完成")
	} else {
		logger.Log().Info("ratelimit: 限流器未启用")
	}

	// 初始化熔断器
	var circuitBreaker domain.CircuitBreaker
	if cfg.CircuitBreaker.Enabled {
		circuitBreaker = infrastructure.NewCircuitBreaker(cfg.CircuitBreaker)
		logger.Log().Info("circuitbreaker: 熔断器初始化完成")
	} else {
		logger.Log().Info("circuitbreaker: 熔断器未启用")
	}

	// 初始化应用服务
	gatewayService := application.NewGatewayService(routeRepo, serviceDiscovery, rateLimiter, circuitBreaker, logger.Log())
	logger.Log().Info("application: 网关应用服务初始化完成")

	// 启动 HTTP 服务
	server := api.NewHTTPServer(gatewayService)
	logger.Log().Info("api: HTTP服务器初始化完成")

	addr := ":" + strconv.Itoa(cfg.App.Port)
	logger.Log().Info("main: 网关服务启动中，监听端口: %s", addr)

	if err := server.Run(addr); err != nil {
		logger.Log().Error("main: 服务启动失败: %v", err)
	}
}
