package infrastructure

import (
	"context"
	"time"

	conf "blog-system/common/pkg/config"
	"blog-system/services/admin/application"
	pb "blog-system/services/stat/proto"

	micro "github.com/CoucouMonEcho/go-framework/micro"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// StatServiceClient 调用 stat-service
type StatServiceClient struct {
	cc  *grpc.ClientConn
	cli pb.StatServiceClient
}

func NewStatServiceClient(cfg *conf.AppConfig) *StatServiceClient {
	var r registry.Registry
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, err := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints, DialTimeout: 3 * time.Second}); err == nil {
			if rg, err2 := regEtcd.NewRegistry(cli); err2 == nil {
				r = rg
			}
		}
	}
	c, _ := micro.NewClient(micro.ClientWithInsecure(), micro.ClientWithRegistry(r, 3*time.Second))
	cc, _ := c.Dial(context.Background(), "stat-service")
	return &StatServiceClient{cc: cc, cli: pb.NewStatServiceClient(cc)}
}

// go-framework/micro resolver 内部处理 service 发现，无需显式 resolve

func (c *StatServiceClient) Overview(ctx context.Context) (int64, int64, int64, error) {
	resp, err := c.cli.Overview(ctx, &pb.OverviewRequest{})
	if err != nil {
		return 0, 0, 0, err
	}
	return resp.PvToday, resp.UvToday, resp.OnlineUsers, nil
}

func (c *StatServiceClient) PVSeries(ctx context.Context, from, to, interval string) ([]map[string]int64, error) {
	resp, err := c.cli.PVTimeSeries(ctx, &pb.PVTimeSeriesRequest{From: from, To: to, Interval: interval})
	if err != nil {
		return nil, err
	}
	out := make([]map[string]int64, 0, len(resp.Points))
	for _, p := range resp.Points {
		if p == nil {
			continue
		}
		out = append(out, map[string]int64{"ts": p.Ts, "value": p.Value})
	}
	return out, nil
}

var _ application.StatClient = (*StatServiceClient)(nil)
