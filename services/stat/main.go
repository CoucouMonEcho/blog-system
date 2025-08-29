package main

import (
	"log"
	"strconv"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/stat/infrastructure"
	persistence "blog-system/services/stat/infrastructure/persistence"
	grpcapi "blog-system/services/stat/interfaces/grpcserver"
	httpapi "blog-system/services/stat/interfaces/httpserver"
	pb "blog-system/services/stat/proto"

	micro "github.com/CoucouMonEcho/go-framework/micro"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	cfg, err := conf.Load("stat")
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	// 初始化全局 Logger
	logger.Init(&cfg.Log)
	db, err := infrastructure.InitDB(cfg)
	if err != nil {
		logger.Log().Error("main: 数据库连接失败: %v", err)
	}
	repo := persistence.NewStatRepository(db)
	http := httpapi.NewHTTPServer()
	// 注入 repo
	http.SetRepository(repo)
	// 启动 gRPC 服务（go-framework/micro）
	agg := infrastructure.NewPVAggregator()
	grpcSrv, _ := micro.NewServer("stat-service")
	// 注册到 etcd
	if len(cfg.Registry.Endpoints) > 0 {
		cli, er := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints})
		if er == nil {
			r, er2 := regEtcd.NewRegistry(cli)
			if er2 == nil {
				grpcSrv, _ = micro.NewServer("stat-service", micro.ServerWithRegistry(r))
			}
		}
	}
	// 注册 gRPC handlers
	pbSrv := grpcapi.NewGRPCServer(agg)
	pb.RegisterStatServiceServer(grpcSrv, pbSrv)
	addr := ":" + strconv.Itoa(cfg.App.Port)
	go func() { _ = grpcSrv.Start(":" + strconv.Itoa(cfg.GRPC.Port)) }()
	if err := http.Run(addr); err != nil {
		logger.Log().Error("main: 服务启动失败: %v", err)
	}
}
