package database

type Panel struct {
	MessageId uint64 `gorm:"column:MESSAGEID"`
	ChannelId uint64 `gorm:"column:CHANNELID"`
	GuildId   uint64 `gorm:"column:GUILDID"`

	Title          string `gorm:"column:TITLE;type:VARCHAR(255)"`
	Content        string `gorm:"column:CONTENT;type:TEXT"`
	Colour         int    `gorm:"column:COLOUR`
	TargetCategory uint64  `gorm:"column:TARGETCATEGORY"`
	ReactionEmote  string `gorm:"column:REACTIONEMOTE;type:VARCHAR(32)"`
}
func (Panel) TableName() string {
	return "panels"
}

func AddPanel(messageId, channelId, guildId uint64, title, content string, colour int, targetCategory uint64, reactionEmote string) {
	Db.Create(&Panel{
		MessageId: messageId,
		ChannelId: channelId,
		GuildId:   guildId,

		Title:          title,
		Content:        content,
		Colour:         colour,
		TargetCategory: targetCategory,
		ReactionEmote:  reactionEmote,
	})
}

func IsPanel(messageId uint64, ch chan bool) {
	var count int
	Db.Table(Panel{}.TableName()).Where(Panel{MessageId: messageId}).Count(&count)
	ch <- count > 0
}

func GetPanelByMessageId(messageId uint64, ch chan Panel) {
	var panel Panel
	Db.Where(Panel{MessageId: messageId}).Take(&panel)
	ch <- panel
}

func GetPanelsByGuild(guildId uint64, ch chan []Panel) {
	var panels []Panel
	Db.Where(Panel{GuildId: guildId}).Find(&panels)
	ch <- panels
}

func DeletePanel(msgId uint64) {
	Db.Where(Panel{MessageId: msgId}).Delete(Panel{})
}
