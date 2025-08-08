package infrastructure

import (
	"fmt"
	"time"

	"github.com/CoucouMonEcho/go-framework/cache"
	redis "github.com/redis/go-redis/v9"
)

// parseDuration 解析时间字符串
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}

// InitCache 初始化缓存连接
func InitCache(cfg *GatewayConfig) (cache.Cache, error) {
	// 检查是否配置了Redis Cluster
	if len(cfg.Redis.Cluster.Addrs) == 0 {
		return nil, fmt.Errorf("未配置Redis Cluster地址")
	}

	// 使用Redis Cluster配置
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        cfg.Redis.Cluster.Addrs,
		Password:     cfg.Redis.Cluster.Password,
		PoolSize:     cfg.Redis.Cluster.PoolSize,
		MinIdleConns: cfg.Redis.Cluster.MinIdleConns,
		MaxRetries:   cfg.Redis.Cluster.MaxRetries,
		DialTimeout:  parseDuration(cfg.Redis.Cluster.DialTimeout),
		ReadTimeout:  parseDuration(cfg.Redis.Cluster.ReadTimeout),
		WriteTimeout: parseDuration(cfg.Redis.Cluster.WriteTimeout),
	})
	return cache.NewRedisCache(client), nil
}
