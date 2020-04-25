package database

import "github.com/TicketsBot/TicketsGo/database"

type ModMailSession struct {
	Uuid           string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild          uint64 `gorm:"column:GUILDID"`
	User           uint64 `gorm:"column:USERID"`
	StaffChannel   uint64 `gorm:"column:CHANNELID;UNIQUE"`
	WelcomeMessage uint64 `gorm:"column:WELCOME_MESSAGE;UNIQUE"`
}

func (ModMailSession) TableName() string {
	return "modmailsessions"
}

func CreateModMailSession(uuid string, guild, user, channel, welcomeMessage uint64) {
	node := ModMailSession{
		Uuid:           uuid,
		Guild:          guild,
		User:           user,
		StaffChannel:   channel,
		WelcomeMessage: welcomeMessage,
	}

	database.Db.Create(&node)
}

func GetModMailSession(userId uint64, ch chan *ModMailSession) {
	var node ModMailSession
	database.Db.Where(ModMailSession{User: userId}).Take(&node)

	if node.User == 0 {
		ch <- nil
	} else {
		ch <- &node
	}
}

func GetModMailSessionByStaffChannel(channelId uint64, ch chan *ModMailSession) {
	var node ModMailSession
	database.Db.Where(ModMailSession{StaffChannel: channelId}).Take(&node)

	if node.User == 0 {
		ch <- nil
	} else {
		ch <- &node
	}
}

func CloseModMailSessions(userId uint64) {
	database.Db.Where(ModMailSession{User: userId}).Delete(ModMailSession{})
}
