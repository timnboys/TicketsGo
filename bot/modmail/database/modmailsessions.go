package database

import "github.com/TicketsBot/TicketsGo/database"

type ModMailSession struct {
	Uuid         string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild        int64  `gorm:"column:GUILDID"`
	User         int64  `gorm:"column:USERID"`
	StaffChannel int64  `gorm:"column:CHANNELID;UN"`
}

func (ModMailSession) TableName() string {
	return "modmailsessions"
}

func CreateModMailSession(uuid string, guild, user, channel int64) {
	node := ModMailSession{
		Uuid:  uuid,
		Guild: guild,
		User:  user,
		StaffChannel: channel,
	}

	database.Db.Create(&node)
}

func GetModMailSession(userId int64, ch chan *ModMailSession) {
	var node ModMailSession
	database.Db.Where(ModMailSession{User: userId}).Take(&node)

	if node.User == 0 {
		ch <- nil
	} else {
		ch <- &node
	}
}

func GetModMailSessionByStaffChannel(channelId int64, ch chan *ModMailSession) {
	var node ModMailSession
	database.Db.Where(ModMailSession{StaffChannel: channelId}).Take(&node)

	if node.User == 0 {
		ch <- nil
	} else {
		ch <- &node
	}
}

func CloseModMailSessions(userId int64) {
	database.Db.Where(ModMailSession{User: userId}).Delete(ModMailSession{})
}
