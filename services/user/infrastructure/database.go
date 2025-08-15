package infrastructure

import (
	"fmt"
	"time"

	"github.com/CoucouMonEcho/go-framework/cache"
	"github.com/CoucouMonEcho/go-framework/orm"
	_ "github.com/go-sql-driver/mysql"
	redis "github.com/redis/go-redis/v9"
)

// InitDB 初始化数据库连接
func InitDB(cfg *AppConfig) (*orm.DB, error) {
	//FIXME 简化 DSN，移除可能有问题的参数
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	//// 注入 query log，使用业务 logger
	//ql := ormql.NewMiddlewareBuilder().LogFunc(func(sql string, args []any) {
	//	logger.L().LogWithContext("user-service", "orm", "DEBUG", "sql=%s args=%v", sql, args)
	//}).Build()
	//db, err := orm.Open(cfg.Database.Driver, dsn, orm.DBWithMiddlewares(
	//	ormotel.NewMiddlewareBuilder(nil).Build(),
	//	ormprom.NewMiddlewareBuilder("blog", "user", "orm", "user orm latency").Build(),
	//	ql,
	//))

	// 完全移除所有中间件，直接使用原始连接
	db, err := orm.Open(cfg.Database.Driver, dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

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
func InitCache(cfg *AppConfig) (cache.Cache, error) {
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
