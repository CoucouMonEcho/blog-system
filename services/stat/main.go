package main

import (
	"log"
	"os"
	"strconv"

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
	var configPath string
	if _, err := os.Stat("/opt/blog-system/configs/stat.yaml"); err == nil {
		configPath = "/opt/blog-system/configs/stat.yaml"
	} else if _, err := os.Stat("../../configs/stat.yaml"); err == nil {
		configPath = "../../configs/stat.yaml"
	} else {
		configPath = "configs/stat.yaml"
	}
	cfg, err := infrastructure.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	lgr, err := logger.NewLogger(&cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	// 初始化全局 Logger
	logger.Init(lgr)
	db, err := infrastructure.InitDB(cfg)
	if err != nil {
		lgr.LogWithContext("stat-service", "database", "ERROR", "数据库连接失败: %v", err)
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
	go func() { _ = grpcSrv.Start(":9004") }()
	if err := http.Run(addr); err != nil {
		lgr.LogWithContext("stat-service", "main", "FATAL", "服务启动失败: %v", err)
	}
}
