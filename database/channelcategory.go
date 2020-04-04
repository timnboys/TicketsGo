package database

type ChannelCategory struct {
	Guild    uint64 `gorm:"column:GUILDID;unique;primary_key"`
	Category uint64 `gorm:"column:CATEGORYID"`
}

func (ChannelCategory) TableName() string {
	return "channelcategory"
}

func GetCategory(guild uint64, ch chan uint64) {
	var node ChannelCategory
	Db.Where(ChannelCategory{Guild: guild}).First(&node)
	ch <- node.Category
}

func SetCategory(guild uint64, category uint64) {
	var node ChannelCategory
	Db.Where(ChannelCategory{Guild: guild}).Assign(ChannelCategory{Category: category}).FirstOrCreate(&node)
}

func DeleteCategory(guild uint64) {
	var node ChannelCategory
	Db.Where(ChannelCategory{Guild: guild}).Take(&node)
	Db.Delete(&node)
}
