package database

type Permissions struct {
	GuildId uint64 `gorm:"column:GUILDID"`
	UserId  uint64 `gorm:"column:USERID"`
	Support bool  `gorm:"column:ISSUPPORT"`
	Admin   bool  `gorm:"column:ISADMIN"`
}

func (Permissions) TableName() string {
	return "permissions"
}

func IsSupport(guildId, userId uint64, ch chan bool) {
	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).First(&node)
	ch <- node.Support
}

func IsAdmin(guildId, userId uint64, ch chan bool) {
	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).First(&node)
	ch <- node.Admin
}

func GetAdmins(guildId uint64, ch chan []uint64) {
	var nodes []Permissions
	Db.Where(Permissions{GuildId: guildId, Admin: true}).Find(&nodes)

	ids := make([]uint64, 0)
	for _, node := range nodes {
		ids = append(ids, node.UserId)
	}

	ch <- ids
}

func GetSupport(guildId uint64, ch chan []uint64) {
	var nodes []Permissions
	Db.Where(Permissions{GuildId: guildId, Support: true, Admin: false}).Find(&nodes)

	ids := make([]uint64, 0)
	for _, node := range nodes {
		ids = append(ids, node.UserId)
	}

	ch <- ids
}

func AddAdmin(guildId, userId uint64) {
	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Assign(Permissions{Admin: true, Support: true}).FirstOrCreate(&node)
}

func AddSupport(guildId, userId uint64) {
	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Assign(Permissions{Support: true}).FirstOrCreate(&node)
}

func RemoveAdmin(guildId, userId uint64) {
	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Take(&node)
	Db.Model(&node).Where("GUILDID = ? AND USERID = ?", guildId, userId).Update("ISADMIN", false)
}

func RemoveSupport(guildId, userId uint64) {
	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Take(&node)
	Db.Model(&node).Where("GUILDID = ? AND USERID = ?", guildId, userId).Updates(map[string]interface{}{
		"ISADMIN": false,
		"ISSUPPORT": false,
	})
}
