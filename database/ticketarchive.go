package database

type TicketArchive struct {
	Uuid     string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild    int64  `gorm:"column:GUILDID"`
	User     int64  `gorm:"column:USERID"`
	Username string `gorm:"column:USERNAME;type:varchar(32)"`
	Id       int    `gorm:"column:TICKETID"`
	CdnUrl   string `gorm:"column:CDNURL;type:varchar(200)"`
}

func (TicketArchive) TableName() string {
	return "ticketarchive"
}

func AddArchive(uuid string, guild int64, user int64, name string, id int, cdnUrl string) {
	Db.Create(&TicketArchive{
		Uuid: uuid,
		Guild: guild,
		User: user,
		Username: name,
		Id: id,
		CdnUrl: cdnUrl,
	})
}
