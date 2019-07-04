package database

type Panels struct {
	MessageId int64 `gorm:"column:MESSAGEID"`
	GuildId int64 `gorm:"column:GUILDID"` // Might be useful in the future so we store it
}

func (Panels) TableName() string {
	return "panels"
}

func AddPanel(messageId int64, guildId int64) {
	Db.Create(&Panels{
		MessageId: messageId,
		GuildId: guildId,
	})
}

func IsPanel(messageId int64, ch chan bool) {
	var count int
	Db.Table(Panels{}.TableName()).Where(Panels{MessageId: messageId}).Count(&count)
	ch <- count > 0
}
