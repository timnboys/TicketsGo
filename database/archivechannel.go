package database

type ArchiveChannel struct {
	Guild   int64 `gorm:"column:GUILDID;unique;primary_key"`
	Channel int64 `gorm:"column:CHANNELID;unique"`
}

func (ArchiveChannel) TableName() string {
	return "archivechannel"
}

func GetArchiveChannel(guild int64, ch chan int64) {
	var node ArchiveChannel
	Db.Where(ArchiveChannel{Guild: guild}).First(&node)
	ch <- node.Channel
}

func SetArchiveChannel(guild int64, channel int64) {
	var node ArchiveChannel
	Db.Where(ArchiveChannel{Guild: guild}).Assign(ArchiveChannel{Channel: channel}).FirstOrCreate(&node)
}
