package database

type Blacklist struct {
	AssocId int   `gorm:"column:ASSOCID;type:int;primary_key;auto_increment"`
	Guild   uint64 `gorm:"column:GUILDID"`
	User    uint64 `gorm:"column:USERID"`
}

func (Blacklist) TableName() string {
	return "blacklist"
}

func IsBlacklisted(guild uint64, user uint64, ch chan bool) {
	var count int
	Db.Table(Blacklist{}.TableName()).Where(Blacklist{Guild: guild, User: user}).Count(&count)
	ch <- count > 0
}

func RemoveBlacklist(guild uint64, user uint64) {
	var node Blacklist
	Db.Where(Blacklist{Guild: guild, User: user}).Take(&node)
	Db.Delete(&node)
}

func AddBlacklist(guild uint64, user uint64) {
	Db.Create(&Blacklist{Guild: guild, User: user})
}
