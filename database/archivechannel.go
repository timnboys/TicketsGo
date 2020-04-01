package database

type ArchiveChannel struct {
	Guild   uint64 `gorm:"column:GUILDID;unique;primary_key"`
	Channel uint64 `gorm:"column:CHANNELID;unique"`
}

func (ArchiveChannel) TableName() string {
	return "archivechannel"
}

func GetArchiveChannel(guild uint64, ch chan uint64) {
	var node ArchiveChannel
	Db.Where(ArchiveChannel{Guild: guild}).First(&node)
	ch <- node.Channel
}

func SetArchiveChannel(guild uint64, channel uint64) {
	var node ArchiveChannel
	Db.Where(ArchiveChannel{Guild: guild}).Assign(ArchiveChannel{Channel: channel}).FirstOrCreate(&node)
}
