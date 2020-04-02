package database

type PingEveryone struct {
	Guild        uint64 `gorm:"column:GUILDID;unique;primary_key"`
	PingEveryone bool  `gorm:"column:PINGEVERYONE"`
}

func (PingEveryone) TableName() string {
	return "pingeveryone"
}

func IsPingEveryone(guild uint64, ch chan bool) {
	var node PingEveryone
	Db.Where(PingEveryone{Guild: guild}).First(&node)
	ch <- node.PingEveryone
}
