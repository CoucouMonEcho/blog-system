package main

import (
	"log"
	"strconv"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/application"
	"blog-system/services/gateway/domain"
	"blog-system/services/gateway/infrastructure"
	httpapi "blog-system/services/gateway/interfaces/httpserver"
)

func main() {
	cfg, err := conf.Load("gateway")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	logger.Init(&cfg.Log)
	logger.Log().Info("main: 开始启动网关服务")

	// 初始化缓存
	cache, err := infrastructure.InitCache(cfg)
	if err != nil {
		logger.Log().Error("main: 缓存连接失败: %v", err)
	} else {
		logger.Log().Info("main: 缓存连接成功")
	}

	// 初始化仓储层
	routeRepo := infrastructure.NewRouteRepository(cfg)
	logger.Log().Info("main: 路由仓储层初始化完成")

	// 初始化服务发现
	serviceDiscovery := infrastructure.NewServiceDiscovery()
	logger.Log().Info("main: 服务发现初始化完成")

	// 初始化限流器
	var rateLimiter domain.RateLimiter
	if cache != nil && cfg.RateLimit.Enabled {
		rateLimiter = infrastructure.NewRateLimiter(cache, cfg.RateLimit)
		logger.Log().Info("main: 限流器初始化完成")
	} else {
		logger.Log().Info("main: 限流器未启用")
	}

	// 初始化熔断器
	var circuitBreaker domain.CircuitBreaker
	if cfg.CircuitBreaker.Enabled {
		circuitBreaker = infrastructure.NewCircuitBreaker(cfg.CircuitBreaker)
		logger.Log().Info("main: 熔断器初始化完成")
	} else {
		logger.Log().Info("main: 熔断器未启用")
	}

	// 初始化应用服务
	gatewayService := application.NewGatewayService(routeRepo, cache, serviceDiscovery, rateLimiter, circuitBreaker)
	logger.Log().Info("main: 网关应用服务初始化完成")

	// 启动 HTTP 服务
	server := httpapi.NewHTTPServer(gatewayService)
	logger.Log().Info("main: HTTP服务器初始化完成")

	addr := ":" + strconv.Itoa(cfg.App.Port)
	logger.Log().Info("main: 网关服务启动中，监听端口: %s", addr)

	if err := server.Run(addr); err != nil {
		logger.Log().Error("main: 服务启动失败: %v", err)
	}
}
