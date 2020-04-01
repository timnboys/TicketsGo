package database

type Prefix struct {
	GuildId uint64  `gorm:"column:GUILDID;unique;primary_key"`
	Prefix  string `gorm:"column:PREFIX;type:varchar(8)"`
}

func (Prefix) TableName() string {
	return "prefix"
}

func SetPrefix(guild uint64, prefix string) {
	var node Prefix
	Db.Where(Prefix{GuildId: guild}).Assign(Prefix{Prefix: prefix}).FirstOrCreate(&node)
}

func GetPrefix(guild uint64, ch chan string) {
	node := Prefix{
		Prefix: "t!",
	}

	Db.Where(Prefix{GuildId: guild}).First(&node)
	ch <- node.Prefix
}
