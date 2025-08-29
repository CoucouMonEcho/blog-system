package infrastructure

import (
	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"fmt"
	"time"

	"github.com/CoucouMonEcho/go-framework/cache"
	"github.com/CoucouMonEcho/go-framework/orm"
	ormotel "github.com/CoucouMonEcho/go-framework/orm/middlewares/opentelemetry"
	ormprom "github.com/CoucouMonEcho/go-framework/orm/middlewares/prometheus"
	ormql "github.com/CoucouMonEcho/go-framework/orm/middlewares/querylog"
	_ "github.com/go-sql-driver/mysql"
	redis "github.com/redis/go-redis/v9"
)

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

func InitDB(cfg *conf.AppConfig) (*orm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&interpolateParams=true",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	return orm.Open(cfg.Database.Driver, dsn, orm.DBWithMiddlewares(
		ormotel.NewMiddlewareBuilder(nil).Build(),
		ormprom.NewMiddlewareBuilder("blog-system", "content", "orm", "content orm latency").Build(),
		ormql.NewMiddlewareBuilder().LogFunc(func(sql string, args []any) {
			logger.Log().Debug("infrastructure: sql=%s args=%v", sql, args)
		}).Build(),
	))
}

func InitCache(cfg *conf.AppConfig) (cache.Cache, error) {
	if len(cfg.Redis.Cluster.Addrs) == 0 {
		return nil, fmt.Errorf("未配置Redis Cluster地址")
	}
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
