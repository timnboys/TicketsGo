package database

type WelcomeMessage struct {
	Guild int64  `gorm:"column:GUILDID;unique;primary_key"`
	Message string `gorm:"column:MESSAGE;type:text"`
}

func (WelcomeMessage) TableName() string {
	return "welcomemessages"
}

func GetWelcomeMessage(guild int64, ch chan string) {
	var node WelcomeMessage
	Db.Where(WelcomeMessage{Guild: guild}).First(&node)
	ch <- node.Message
}

func SetWelcomeMessage(guild int64, message string) {
	var node WelcomeMessage
	Db.Where(WelcomeMessage{Guild: guild}).Assign(WelcomeMessage{Message: message}).FirstOrCreate(&node)
}
