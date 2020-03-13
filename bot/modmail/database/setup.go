package database

import "github.com/TicketsBot/TicketsGo/database"

func Setup() {
	database.Db.AutoMigrate(
		ModMailSession{},
	)
}
