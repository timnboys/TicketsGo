package database

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"time"
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

	db.DB().SetConnMaxLifetime(time.Duration(config.Conf.Database.Lifetime) * time.Second)

	db.Set("gorm:table_options", "charset=utf8mb4")
	db.BlockGlobalUpdate(true)

	Db = Database {db}
}

func Setup() {
	Db.AutoMigrate(
		ArchiveChannel{},
		Blacklist{},
		CannedResponse{},
		ChannelCategory{},
		DmOnOpen{},
		Panel{},
		PanelSettings{},
		Permissions{},
		PingEveryone{},
		Prefix{},
		PremiumGuilds{},
		PremiumKeys{},
		RolePermissions{},
		TicketArchive{},
		TicketFirstResponse{},
		TicketLimit{},
		Ticket{},
		TicketNamingScheme{},
		TicketWebhook{},
		UserCanClose{},
		UserData{},
		WelcomeMessage{},
		)
}

func IsConnected(ch chan bool) {
	ch <- Db.DB.DB().Ping() == nil
}

