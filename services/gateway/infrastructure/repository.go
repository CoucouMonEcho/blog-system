package infrastructure

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"blog-system/services/gateway/domain"

	"strconv"

	"github.com/CoucouMonEcho/go-framework/cache"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	regEtcd "github.com/CoucouMonEcho/go-framework/micro/registry/etcd"
	redis "github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// RouteRepository 路由仓储实现
type RouteRepository struct {
	routes domain.RouteConfig
	mu     sync.RWMutex
}

// NewRouteRepository 创建路由仓储
func NewRouteRepository(cfg *conf.AppConfig) *RouteRepository {
	// 转换配置到domain层结构
	routes := domain.RouteConfig{
		User: domain.Route{
			Prefix:  cfg.Routes.User.Prefix,
			Target:  cfg.Routes.User.Target,
			Timeout: parseDuration(cfg.Routes.User.Timeout),
			Retries: cfg.Routes.User.Retries,
		},
		Content: domain.Route{
			Prefix:  cfg.Routes.Content.Prefix,
			Target:  cfg.Routes.Content.Target,
			Timeout: parseDuration(cfg.Routes.Content.Timeout),
			Retries: cfg.Routes.Content.Retries,
		},
		Stat: domain.Route{
			Prefix:  cfg.Routes.Stat.Prefix,
			Target:  cfg.Routes.Stat.Target,
			Timeout: parseDuration(cfg.Routes.Stat.Timeout),
			Retries: cfg.Routes.Stat.Retries,
		},
		Admin: domain.Route{
			Prefix:  cfg.Routes.Admin.Prefix,
			Target:  cfg.Routes.Admin.Target,
			Timeout: parseDuration(cfg.Routes.Admin.Timeout),
			Retries: cfg.Routes.Admin.Retries,
		},
	}

	return &RouteRepository{
		routes: routes,
	}
}

// GetRoutes 获取所有路由
func (r *RouteRepository) GetRoutes() domain.RouteConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.routes
}

// GetRouteByPath 根据路径获取路由
func (r *RouteRepository) GetRouteByPath(path string) *domain.Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查各个路由前缀
	routes := []domain.Route{r.routes.User, r.routes.Content, r.routes.Stat, r.routes.Admin}
	for _, route := range routes {
		if strings.HasPrefix(path, route.Prefix) {
			return &route
		}
	}
	return nil
}

// ServiceDiscovery 服务发现实现
type ServiceDiscovery struct {
	registry registry.Registry
}

// NewServiceDiscovery 创建服务发现
func NewServiceDiscovery() *ServiceDiscovery {
	gcfg, _ := conf.Load("gateway")
	cli, _ := clientv3.New(clientv3.Config{
		Endpoints:   gcfg.Registry.Endpoints,
		DialTimeout: 3 * time.Second,
	})
	r, _ := regEtcd.NewRegistry(cli)
	return &ServiceDiscovery{
		registry: r,
	}
}

// resolveTarget 解析 service://name -> http://host:port
func (s *ServiceDiscovery) resolveTarget(target string) (string, bool) {
	if s == nil {
		return target, true
	}
	u, err := url.Parse(target)
	if err != nil || u.Scheme != "service" || u.Host == "" {
		return target, true
	}
	serviceName := u.Host
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	srvs, err := s.registry.ListServices(ctx, serviceName)
	if err != nil || len(srvs) == 0 {
		logger.Log().Error("infrastructure: 查询节点失败: service=%s err=%v", serviceName, err)
		return "", false
	}
	//TODO 简单返回列表第一个地址 可优化为集群分发
	return srvs[0].Address, true
}

// Resolve 对外暴露解析方法
func (s *ServiceDiscovery) Resolve(target string) (string, bool) {
	return s.resolveTarget(target)
}

// GetServiceHealth 获取服务健康状态
func (s *ServiceDiscovery) GetServiceHealth(target string) bool {
	resolved, ok := s.resolveTarget(target)
	if !ok {
		return false
	}
	client := &http.Client{Timeout: 5 * time.Second}
	u, err := url.Parse(resolved)
	if err != nil {
		logger.Log().Error("infrastructure: 解析地址异常: resolved=%s err=%v", resolved, err)
		return false
	}
	// 访问 /health 更准确
	healthURL := u.ResolveReference(&url.URL{Path: "/health"})
	req, err := http.NewRequest("GET", healthURL.String(), nil)
	if err != nil {
		logger.Log().Error("infrastructure: 健康检查失败: url=%s err=%v", healthURL.String(), err)
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log().Error("infrastructure: 健康检查请求失败: url=%s err=%v", healthURL.String(), err)
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode < 500
}

// GetServiceLatency 获取服务延迟
func (s *ServiceDiscovery) GetServiceLatency(target string) time.Duration {
	resolved, ok := s.resolveTarget(target)
	if !ok {
		return 0
	}
	start := time.Now()
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Head(resolved)
	if err != nil {
		logger.Log().Error("infrastructure: 延迟探测失败: resolved=%s err=%v", resolved, err)
		return 0
	}
	defer func() { _ = resp.Body.Close() }()
	return time.Since(start)
}

// RateLimiter 使用Redis滑动窗口限流
type RateLimiter struct {
	redisClient *redis.ClusterClient
	config      conf.RateLimitConfig
}

// NewRateLimiter 创建限流器 - 使用 Redis 实现
func NewRateLimiter(_ cache.Cache, config conf.RateLimitConfig) *RateLimiter {
	gcfg, _ := conf.Load("gateway")
	redisClient := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        gcfg.Redis.Cluster.Addrs,
		Password:     gcfg.Redis.Cluster.Password,
		PoolSize:     gcfg.Redis.Cluster.PoolSize,
		MinIdleConns: gcfg.Redis.Cluster.MinIdleConns,
		MaxRetries:   gcfg.Redis.Cluster.MaxRetries,
		DialTimeout:  parseDuration(gcfg.Redis.Cluster.DialTimeout),
		ReadTimeout:  parseDuration(gcfg.Redis.Cluster.ReadTimeout),
		WriteTimeout: parseDuration(gcfg.Redis.Cluster.WriteTimeout),
	})
	return &RateLimiter{redisClient: redisClient, config: config}
}

// Allow 检查是否允许请求 - 基于1秒滑动窗口
func (r *RateLimiter) Allow(client string) bool {
	if !r.config.Enabled {
		return true
	}
	if client == "" {
		client = "anonymous"
	}
	ctx := context.Background()
	key := "rl:" + client
	// 当前时间（毫秒）
	now := time.Now().UnixNano() / int64(time.Millisecond)
	windowStart := now - 1000 // 1秒窗口
	// 清理窗口外的记录
	_ = r.redisClient.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10)).Err()
	// 统计当前窗口内的请求数
	cnt, err := r.redisClient.ZCard(ctx, key).Result()
	if err != nil {
		logger.Log().Warn("ratelimit: 统计失败: key=%s err=%v", key, err)
		return true // 容错：出错则放行
	}
	limit := int64(r.config.RequestsPerSecond + r.config.Burst)
	if cnt >= limit {
		return false
	}
	// 记录此次请求（score 与 member 都用时间戳保证去重）
	_ = r.redisClient.ZAdd(ctx, key, redis.Z{Score: float64(now), Member: now}).Err()
	_ = r.redisClient.Expire(ctx, key, 2*time.Second).Err()
	return true
}

// Reset 重置限流器
func (r *RateLimiter) Reset(client string) {}

// CircuitBreaker 熔断器实现
type CircuitBreaker struct {
	failures    map[string]int
	lastFailure map[string]time.Time
	mu          sync.RWMutex
	config      conf.CircuitBreakerConfig
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config conf.CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		failures:    make(map[string]int),
		lastFailure: make(map[string]time.Time),
		config:      config,
	}
}

// IsOpen 检查熔断器是否开启
func (c *CircuitBreaker) IsOpen(target string) bool {
	if !c.config.Enabled {
		return false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	failures := c.failures[target]
	lastFailure := c.lastFailure[target]
	if failures >= c.config.FailureThreshold {
		recoveryTimeout := parseDuration(c.config.RecoveryTimeout)
		if time.Since(lastFailure) < recoveryTimeout {
			return true
		}
	}
	return false
}

// RecordSuccess 记录成功
func (c *CircuitBreaker) RecordSuccess(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures[target] = 0
}

// RecordFailure 记录失败
func (c *CircuitBreaker) RecordFailure(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures[target]++
	c.lastFailure[target] = time.Now()
}

// TryReset 尝试重置熔断器
func (c *CircuitBreaker) TryReset(target string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures[target] = 0
}
