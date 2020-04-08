package database

type UserCanClose struct {
	Guild    uint64 `gorm:"column:GUILDID;unique;primary_key"`
	CanClose *bool  `gorm:"column:CANCLOSE"`
}

func (UserCanClose) TableName() string {
	return "usercanclose"
}

func IsUserCanClose(guild uint64, ch chan bool) {
	var node UserCanClose
	Db.Where(UserCanClose{Guild: guild}).First(&node)

	if node.CanClose == nil {
		ch <- true
	} else {
		ch <- *node.CanClose
	}
}
