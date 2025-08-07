package infrastructure

import (
	"fmt"

	"github.com/CoucouMonEcho/go-framework/cache"
	"github.com/CoucouMonEcho/go-framework/orm"
	redis "github.com/redis/go-redis/v9"
)

// InitDB 初始化数据库连接
func InitDB(cfg *AppConfig) (*orm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)
	db, err := orm.Open(cfg.Database.Driver, dsn, orm.DBWithMiddlewares())
	if err != nil {
		return nil, err
	}
	return db, nil
}

// InitCache 初始化缓存连接
func InitCache(cfg *AppConfig) (cache.Cache, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{cfg.Redis.Addr},
		Password: cfg.Redis.Password,
	})
	return cache.NewRedisCache(client), nil
}
