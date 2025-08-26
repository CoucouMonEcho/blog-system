package infrastructure

import (
	"context"
	"fmt"
	"net"
	"time"

	conf "blog-system/common/pkg/config"

	"github.com/CoucouMonEcho/go-framework/micro/registry"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func RegisterService(cfg *conf.AppConfig) error {
	if len(cfg.Registry.Endpoints) == 0 {
		return nil
	}
	cli, err := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints, DialTimeout: 3 * time.Second})
	if err != nil {
		return err
	}
	r, err := regEtcd.NewRegistry(cli)
	if err != nil {
		return err
	}
	ip := localIP()
	si := registry.ServiceInstance{Name: cfg.App.Name, Address: fmt.Sprintf("http://%s:%d", ip, cfg.App.Port), Weight: 10}
	return r.Register(context.Background(), si)
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}
	return "127.0.0.1"
}
