package database

type WelcomeMessage struct {
	Guild uint64  `gorm:"column:GUILDID;unique;primary_key"`
	Message string `gorm:"column:MESSAGE;type:text CHARACTER SET utf8 COLLATE utf8_unicode_ci"`
}

func (WelcomeMessage) TableName() string {
	return "welcomemessages"
}

func GetWelcomeMessage(guild uint64, ch chan string) {
	var node WelcomeMessage
	Db.Where(WelcomeMessage{Guild: guild}).First(&node)
	ch <- node.Message
}

func SetWelcomeMessage(guild uint64, message string) {
	var node WelcomeMessage
	Db.Where(WelcomeMessage{Guild: guild}).Assign(WelcomeMessage{Message: message}).FirstOrCreate(&node)
}
