package main

import (
	"log"
	"strconv"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/content/application"
	infra "blog-system/services/content/infrastructure"
	persistence "blog-system/services/content/infrastructure/persistence"
	grpcapi "blog-system/services/content/interfaces/grpcserver"
	httpapi "blog-system/services/content/interfaces/httpserver"
	pb "blog-system/services/content/proto"

	"github.com/CoucouMonEcho/go-framework/micro"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	cfg, err := conf.Load("content")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	lgr, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}

	// 初始化全局 Logger
	logger.Init(lgr)
	db, err := infra.InitDB(cfg)
	if err != nil {
		lgr.LogWithContext("content-service", "database", "FATAL", "数据库连接失败: %v", err)
		return
	}
	cache, _ := infra.InitCache(cfg)

	repo := persistence.NewContentRepository(db)
	app := application.NewContentService(repo, lgr, cache)

	http := httpapi.NewHTTPServer(app)

	// gRPC 服务
	grpcSrv, _ := micro.NewServer("content-service")
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, er := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints}); er == nil {
			if r, er2 := regEtcd.NewRegistry(cli); er2 == nil {
				grpcSrv, _ = micro.NewServer("content-service", micro.ServerWithRegistry(r))
			}
		}
	}
	pb.RegisterContentAdminServiceServer(grpcSrv, grpcapi.NewAdminGRPCServer(app))

	// 注册到注册中心
	if err := infra.RegisterService(cfg); err != nil {
		lgr.LogWithContext("content-service", "registry", "ERROR", "注册到注册中心失败: %v", err)
	}
	addr := ":" + strconv.Itoa(cfg.App.Port)
	go func() { _ = grpcSrv.Start(":9002") }()
	if err := http.Run(addr); err != nil {
		lgr.LogWithContext("content-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
