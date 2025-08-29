package main

import (
	"log"
	"strconv"

	"blog-system/common/pkg/logger"

	conf "blog-system/common/pkg/config"
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
	cfg, err := conf.Load("user")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	logger.Init(&cfg.Log)
	logger.Log().Info("main: 开始启动用户服务")

	// 初始化数据库连接
	db, err := infra.InitDB(cfg)
	if err != nil {
		logger.Log().Error("main: 数据库连接失败，服务退出: %v", err)
		return
	}
	logger.Log().Info("main: 数据库连接成功")

	// 初始化缓存
	cache, err := infra.InitCache(cfg)
	if err != nil {
		logger.Log().Error("main: 缓存连接失败: %v", err)
	} else {
		logger.Log().Info("main: 缓存连接成功")
	}

	// 初始化仓储层
	userRepo := persistence.NewUserRepository(db)
	logger.Log().Info("main: 用户仓储层初始化完成")

	// 初始化应用服务
	userService := application.NewUserService(userRepo, cache)
	logger.Log().Info("main: 用户应用服务初始化完成")

	// 启动 HTTP 服务
	server := httpapi.NewHTTPServer(userService)
	logger.Log().Info("main: HTTP服务器初始化完成")

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

	// 注册到注册中心 HTTP
	if err := infra.RegisterService(cfg); err != nil {
		logger.Log().Error("main: 注册到注册中心失败: %v", err)
	} else {
		logger.Log().Info("main: 注册中心注册成功")
	}

	addr := ":" + strconv.Itoa(cfg.App.Port)
	logger.Log().Info("main: 用户服务启动中，监听端口: %s", addr)

	go func() { _ = grpcSrv.Start(":9001") }()
	if err := server.Run(addr); err != nil {
		logger.Log().Error("main: 服务启动失败: %v", err)
	}
}
