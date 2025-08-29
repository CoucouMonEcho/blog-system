package infrastructure

import (
	conf "blog-system/common/pkg/config"
	"blog-system/common/pkg/logger"
	"fmt"

	"github.com/CoucouMonEcho/go-framework/orm"
	ormotel "github.com/CoucouMonEcho/go-framework/orm/middlewares/opentelemetry"
	ormprom "github.com/CoucouMonEcho/go-framework/orm/middlewares/prometheus"
	ormql "github.com/CoucouMonEcho/go-framework/orm/middlewares/querylog"
	_ "github.com/go-sql-driver/mysql"
)

func InitDB(cfg *conf.AppConfig) (*orm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local&interpolateParams=true", cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name)
	return orm.Open(cfg.Database.Driver, dsn, orm.DBWithMiddlewares(
		ormotel.NewMiddlewareBuilder(nil).Build(),
		ormprom.NewMiddlewareBuilder("blog-system", "stat", "orm", "stat orm latency").Build(),
		ormql.NewMiddlewareBuilder().LogFunc(func(sql string, args []any) {
			logger.Log().Debug("infrastructure: sql=%s args=%v", sql, args)
		}).Build(),
	))
}
