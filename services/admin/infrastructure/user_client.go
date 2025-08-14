package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"blog-system/services/admin/application"

	"github.com/CoucouMonEcho/go-framework/micro/registry"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// UserServiceClient 通过服务发现（或直连）访问 user-service
type UserServiceClient struct {
	baseURL  string
	client   *http.Client
	registry registry.Registry
}

// NewUserServiceClient 创建客户端，支持 service://name 与 http://host:port
func NewUserServiceClient(cfg *AppConfig) *UserServiceClient {
	timeout := time.Second * 5
	if cfg.UserService.Timeout > 0 {
		timeout = time.Duration(cfg.UserService.Timeout) * time.Millisecond
	}
	var reg registry.Registry
	if len(cfg.Registry.Endpoints) > 0 {
		if cli, err := clientv3.New(clientv3.Config{Endpoints: cfg.Registry.Endpoints, DialTimeout: 3 * time.Second}); err == nil {
			if r, err2 := regEtcd.NewRegistry(cli); err2 == nil {
				reg = r
			}
		}
	}
	base := cfg.UserService.BaseURL
	if base == "" {
		base = "service://user-service"
	}
	return &UserServiceClient{
		baseURL:  base,
		client:   &http.Client{Timeout: timeout},
		registry: reg,
	}
}

func (c *UserServiceClient) resolve() string {
	if c.registry == nil {
		return c.baseURL
	}
	// 支持 service://user-service 解析
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 解析 service://<name>
	// 简化解析：直接把 baseURL 当作 name（若包含 service:// 则取 Host）
	base := c.baseURL
	if len(base) >= 10 && base[:10] == "service://" {
		name := base[10:]
		svcs, err := c.registry.ListServices(ctx, name)
		if err == nil && len(svcs) > 0 {
			return svcs[0].Address
		}
		return ""
	}
	return base
}

// Login 实现 application.UserAuthClient
func (c *UserServiceClient) Login(ctx context.Context, username, password string) (string, string, error) {
	base := c.resolve()
	if base == "" {
		return "", "", errors.New("user-service 未发现")
	}
	payload := map[string]string{"username": username, "password": password}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/login", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 300 {
		return "", "", errors.New("登录失败")
	}
	var res struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
			User  struct {
				Role string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", "", err
	}
	if res.Code != 0 || res.Data.Token == "" {
		return "", "", errors.New("登录响应异常")
	}
	return res.Data.Token, res.Data.User.Role, nil
}

var _ application.UserAuthClient = (*UserServiceClient)(nil)
