package database

import (
	"github.com/TicketsBot/TicketsGo/database"
	"time"
)

type ModMailArchive struct {
	Uuid      string    `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild     uint64    `gorm:"column:GUILDID"`
	User      uint64    `gorm:"column:USERID"`
	CloseTime time.Time `gorm:"column:CLOSETIME"`
}

func (ModMailArchive) TableName() string {
	return "modmail_archive"
}

func (m *ModMailArchive) Store() {
	database.Db.Create(m)
}
