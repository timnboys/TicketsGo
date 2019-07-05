package database

import "strconv"

type Prefix struct {
	GuildId int64  `gorm:"column:GUILDID;unique;primary_key"`
	Prefix  string `gorm:"column:PREFIX;type:varchar(8)"`
}

func (Prefix) TableName() string {
	return "prefix"
}

func SetPrefix(guild int64, prefix string) {
	var node Prefix
	Db.Where(Prefix{GuildId: guild}).Assign(Prefix{Prefix: prefix}).FirstOrCreate(&node)
}

func GetPrefix(guild string, ch chan string) {
	node := Prefix{
            Prefix: "t!",
        }


	i, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		ch <- ""
		return
	}

	Db.Where(Prefix{GuildId: i}).First(&node)
	ch <- node.Prefix
}
