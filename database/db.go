package database

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type Database struct {
	*gorm.DB
}

var(
	Db Database
)

func Connect() {
	uri := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Conf.Database.Username,
		config.Conf.Database.Password,
		config.Conf.Database.Host,
		config.Conf.Database.Port,
		config.Conf.Database.Database,
	)

	db, err := gorm.Open("mysql", uri); if err != nil {
		panic(err)
	}

	db.DB().SetMaxOpenConns(config.Conf.Database.Pool.MaxConnections)
	db.DB().SetMaxIdleConns(config.Conf.Database.Pool.MaxIdle)

	Db = Database {db}
}

func Setup() {
	Db.AutoMigrate(
		Prefix{},
		PremiumGuilds{},
		)
}
