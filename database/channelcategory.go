package database

type ChannelCategory struct {
	Guild int64  `gorm:"column:GUILDID;unique;primary_key"`
	Category  int64 `gorm:"column:CATEGORYID"`
}

func (ChannelCategory) TableName() string {
	return "channelcategory"
}

func GetCategory(guild int64, ch chan int64) {
	var node ChannelCategory
	Db.Where(ChannelCategory{Guild: guild}).First(&node)
	ch <- node.Category
}

func SetCategory(guild int64, category int64) {
	var node ChannelCategory
	Db.Where(ChannelCategory{Guild: guild}).Assign(ChannelCategory{Category: category}).FirstOrCreate(&node)
}

func DeleteCategory(guild int64) {
	var node ChannelCategory
	Db.Where(ChannelCategory{Guild: guild}).Take(&node)
	Db.Delete(&node)
}
