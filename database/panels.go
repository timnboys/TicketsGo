package database

type Panel struct {
	MessageId int64 `gorm:"column:MESSAGEID"`
	ChannelId int64 `gorm:"column:CHANNELID"`
	GuildId   int64 `gorm:"column:GUILDID"` // Might be useful in the future so we store it
}

func (Panel) TableName() string {
	return "panels"
}

func AddPanel(messageId, channelId, guildId int64) {
	Db.Create(&Panel{
		MessageId: messageId,
		ChannelId: channelId,
		GuildId:   guildId,
	})
}

func IsPanel(messageId int64, ch chan bool) {
	var count int
	Db.Table(Panel{}.TableName()).Where(Panel{MessageId: messageId}).Count(&count)
	ch <- count > 0
}

func GetPanelsByGuild(guildId int64, ch chan []Panel) {
	var panels []Panel
	Db.Where(Panel{GuildId: guildId}).Find(&panels)
	ch <- panels
}

func DeletePanel(msgId int64) {
	var node Panel
	Db.Where(Panel{MessageId: msgId}).Take(&node)
	Db.Delete(&node)
}
