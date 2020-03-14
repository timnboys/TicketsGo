package database

import "github.com/TicketsBot/TicketsGo/database"

type ModMailArchive struct {
	Uuid     string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild    int64  `gorm:"column:GUILDID"`
	User     int64  `gorm:"column:USERID"`
	Username string `gorm:"column:USERNAME;type:varchar(32)"`
	CdnUrl   string `gorm:"column:CDNURL;type:varchar(200)"`
}

func (ModMailArchive) TableName() string {
	return "modmail_archive"
}

func (m *ModMailArchive) Store() {
	database.Db.Create(m)
}
