package clients

import (
	"context"
	"errors"
	"time"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/admin/application"
	"blog-system/services/admin/domain"
	upb "blog-system/services/user/proto"

	micro "github.com/CoucouMonEcho/go-framework/micro"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

// UserServiceClient 通过 gRPC 访问 user-service
type UserServiceClient struct {
	cc       *grpc.ClientConn
	cli      upb.UserServiceClient
	registry registry.Registry
}

// NewUserServiceClient 创建 gRPC 客户端
func NewUserServiceClient(cfg *conf.AppConfig) *UserServiceClient {
	var reg registry.Registry
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, err := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints, DialTimeout: 3 * time.Second}); err == nil {
			if r, err2 := regEtcd.NewRegistry(cli); err2 == nil {
				reg = r
			}
		}
	}
	c, _ := micro.NewClient(micro.ClientWithInsecure(), micro.ClientWithRegistry(reg, 3*time.Second))
	cc, _ := c.Dial(context.Background(), "user-grpc")
	return &UserServiceClient{cc: cc, cli: upb.NewUserServiceClient(cc), registry: reg}
}

func (c *UserServiceClient) Create(ctx context.Context, u *domain.User) error {
	_, err := c.cli.Register(ctx, &upb.RegisterRequest{Username: u.Username, Email: u.Email, Password: u.Password})
	return err
}

func (c *UserServiceClient) Update(ctx context.Context, u *domain.User) error {
	_, err := c.cli.UpdateUserInfo(ctx, &upb.UpdateUserInfoRequest{UserId: u.ID, Username: u.Username, Email: u.Email, Avatar: u.Avatar, Role: u.Role})
	return err
}

func (c *UserServiceClient) Delete(ctx context.Context, id int64) error {
	_, err := c.cli.UpdateUserStatus(ctx, &upb.UpdateUserStatusRequest{UserId: id, Status: 1})
	return err
}

func (c *UserServiceClient) List(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	resp, err := c.cli.ListUsers(ctx, &upb.ListUsersRequest{Page: int32(page), PageSize: int32(pageSize)})
	if err != nil {
		logger.Log().Error("clients: 列表用户失败: %v", err)
		return nil, 0, err
	}
	if resp.Code != 0 {
		return nil, 0, errors.New(resp.Message)
	}
	out := make([]*domain.User, 0, len(resp.Data))
	for _, u := range resp.Data {
		out = append(out, &domain.User{ID: u.Id, Username: u.Username, Email: u.Email, Role: u.Role, Status: int(u.Status)})
	}
	return out, resp.Total, nil
}

var _ application.UserClient = (*UserServiceClient)(nil)
