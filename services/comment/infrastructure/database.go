package infrastructure

import (
	"fmt"
	"github.com/CoucouMonEcho/go-framework/orm"
	_ "github.com/go-sql-driver/mysql"
)

func InitDB(cfg *AppConfig) (*orm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.Name,
	)
	return orm.Open(cfg.Database.Driver, dsn, orm.DBWithMiddlewares())
}
