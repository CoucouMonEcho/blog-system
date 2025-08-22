package main

import (
	"log"
	"os"
	"strconv"

	"blog-system/common/pkg/logger"

	"blog-system/services/user/application"
	infra "blog-system/services/user/infrastructure"
	persistence "blog-system/services/user/infrastructure/persistence"
	grpcapi "blog-system/services/user/interfaces/grpcserver"
	httpapi "blog-system/services/user/interfaces/httpserver"
	pb "blog-system/services/user/proto"

	micro "github.com/CoucouMonEcho/go-framework/micro"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	// 加载配置 - 修复路径问题
	var configPath string

	// 优先使用绝对路径（部署环境）
	if _, err := os.Stat("/opt/blog-system/configs/user.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/user.yaml"
	} else if _, err := os.Stat("../../configs/user.yaml"); err == nil {
		// 开发环境
		configPath = "../../configs/user.yaml"
	} else {
		// fallback到相对路径
		configPath = "configs/user.yaml"
	}

	log.Printf("使用配置文件: %s", configPath)

	cfg, err := infra.LoadConfig(configPath)
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
	loggerInstance.LogWithContext("user-service", "main", "INFO", "开始启动用户服务")
	loggerInstance.LogWithContext("user-service", "main", "INFO", "配置文件: %s", configPath)

	// 初始化数据库连接 - 添加错误处理但不退出
	db, err := infra.InitDB(cfg)
	if err != nil {
		loggerInstance.LogWithContext("user-service", "database", "FATAL", "数据库连接失败，服务退出: %v", err)
		return
	}
	loggerInstance.LogWithContext("user-service", "database", "INFO", "数据库连接成功")

	// 初始化缓存 - 添加错误处理但不退出
	cache, err := infra.InitCache(cfg)
	if err != nil {
		loggerInstance.LogWithContext("user-service", "cache", "ERROR", "缓存连接失败: %v", err)
		loggerInstance.LogWithContext("user-service", "cache", "WARN", "缓存连接失败，但继续启动服务")
	} else {
		loggerInstance.LogWithContext("user-service", "cache", "INFO", "缓存连接成功")
	}

	// 初始化仓储层
	userRepo := persistence.NewUserRepository(db)
	loggerInstance.LogWithContext("user-service", "repository", "INFO", "用户仓储层初始化完成")

	// 初始化应用服务
	userService := application.NewUserService(userRepo, cache, loggerInstance)
	loggerInstance.LogWithContext("user-service", "application", "INFO", "用户应用服务初始化完成")

	// 启动 HTTP 服务
	server := httpapi.NewHTTPServer(userService)
	loggerInstance.LogWithContext("user-service", "api", "INFO", "HTTP服务器初始化完成")

	// 启动 gRPC 服务
	grpcSrv, _ := micro.NewServer("user-service")
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, er := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints}); er == nil {
			if r, er2 := regEtcd.NewRegistry(cli); er2 == nil {
				grpcSrv, _ = micro.NewServer("user-service", micro.ServerWithRegistry(r))
			}
		}
	}
	pb.RegisterUserServiceServer(grpcSrv, grpcapi.NewGRPCServer(userService))

	// 注册到注册中心（HTTP）
	if err := infra.RegisterService(cfg); err != nil {
		loggerInstance.LogWithContext("user-service", "registry", "ERROR", "注册到注册中心失败: %v", err)
	} else {
		loggerInstance.LogWithContext("user-service", "registry", "INFO", "注册中心注册成功")
	}

	addr := ":" + strconv.Itoa(cfg.App.Port)
	loggerInstance.LogWithContext("user-service", "main", "INFO", "用户服务启动中，监听端口: %s", addr)

	go func() { _ = grpcSrv.Start(":9002") }()
	if err := server.Run(addr); err != nil {
		loggerInstance.LogWithContext("user-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
