package infrastructure

import (
	"fmt"

	"github.com/CoucouMonEcho/go-framework/orm"
	ormotel "github.com/CoucouMonEcho/go-framework/orm/middlewares/opentelemetry"
	ormprom "github.com/CoucouMonEcho/go-framework/orm/middlewares/prometheus"
)

func InitDB(cfg *AppConfig) (*orm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
	)
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "mysql"
	}
	return orm.Open(cfg.Database.Driver, dsn, orm.DBWithMiddlewares(
		ormotel.NewMiddlewareBuilder(nil).Build(),
		ormprom.NewMiddlewareBuilder("blog", "admin", "orm", "admin orm latency").Build(),
	))
}
