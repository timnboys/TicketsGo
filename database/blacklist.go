package database

type Blacklist struct {
	AssocId int   `gorm:"column:ASSOCID;type:int;primary_key;auto_increment"`
	Guild   int64 `gorm:"column:GUILDID"`
	User    int64 `gorm:"column:USERID"`
}

func (Blacklist) TableName() string {
	return "blacklist"
}

func IsBlacklisted(guild int64, user int64, ch chan bool) {
	var count int
	Db.Table(Blacklist{}.TableName()).Where(Blacklist{Guild: guild, User: user}).Count(&count)
	ch <- count > 0
}

func RemoveBlacklist(guild int64, user int64) {
	var node Blacklist
	Db.Where(Blacklist{Guild: guild, User: user}).Take(&node)
	Db.Delete(&node)
}

func AddBlacklist(guild int64, user int64) {
	Db.Create(&Blacklist{Guild: guild, User: user})
}
