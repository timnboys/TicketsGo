package database

type Blacklist struct {
	Guild int64  `gorm:"column:GUILDID"`
	User  int64 `gorm:"column:USERID"`
}

func (Blacklist) TableName() string {
	return "blacklist"
}

func IsBlacklisted(guild int64, user int64, ch chan bool) {
	var count int
	Db.Where(Blacklist{Guild: guild, User: user}).Count(&count)
	ch <- count > 0
}

func RemoveBlacklist(guild int64, user int64) {
	Db.Delete(&Blacklist{Guild: guild, User: user})
}

func AddBlacklist(guild int64, user int64) {
	Db.Create(&Blacklist{Guild: guild, User: user})
}
