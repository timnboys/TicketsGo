package database

import "strconv"

type Permissions struct {
	GuildId int64 `gorm:"column:GUILDID"`
	UserId  int64 `gorm:"column:USERID"`
	Support bool  `gorm:"column:ISSUPPORT"`
	Admin   bool  `gorm:"column:ISADMIN"`
}

func (Permissions) TableName() string {
	return "permissions"
}

func IsSupport(guild string, user string, ch chan bool) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		ch <- false
		return
	}

	userId, err := strconv.ParseInt(user, 10, 64); if err != nil {
		ch <- false
		return
	}

	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).First(&node)
	ch <- node.Support
}

func IsAdmin(guild string, user string, ch chan bool) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		ch <- false
		return
	}

	userId, err := strconv.ParseInt(user, 10, 64); if err != nil {
		ch <- false
		return
	}

	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).First(&node)
	ch <- node.Admin
}

func GetAdmins(guild string, ch chan []int64) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		ch <- []int64{}
		return
	}

	var nodes []Permissions
	Db.Where(Permissions{GuildId: guildId, Admin: true}).Find(&nodes)

	ids := make([]int64, 0)
	for _, node := range nodes {
		ids = append(ids, node.UserId)
	}

	ch <- ids
}

func GetSupport(guild string, ch chan []int64) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		ch <- []int64{}
		return
	}

	var nodes []Permissions
	Db.Where(Permissions{GuildId: guildId, Support: true, Admin: false}).Find(&nodes)

	ids := make([]int64, 0)
	for _, node := range nodes {
		ids = append(ids, node.UserId)
	}

	ch <- ids
}

func AddAdmin(guild string, user string) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		return
	}

	userId, err := strconv.ParseInt(user, 10, 64); if err != nil {
		return
	}

	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Assign(Permissions{Admin: true, Support: true}).FirstOrCreate(&node)
}

func AddSupport(guild string, user string) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		return
	}

	userId, err := strconv.ParseInt(user, 10, 64); if err != nil {
		return
	}

	var node Permissions
	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Assign(Permissions{Support: true}).FirstOrCreate(&node)
}

func RemoveAdmin(guild string, user string) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		return
	}

	userId, err := strconv.ParseInt(user, 10, 64); if err != nil {
		return
	}

	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Update("ISADMIN", false)
}

func RemoveSupport(guild string, user string) {
	guildId, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		return
	}

	userId, err := strconv.ParseInt(user, 10, 64); if err != nil {
		return
	}

	Db.Where(Permissions{GuildId: guildId, UserId: userId}).Updates(map[string]interface{}{
		"ISADMIN": false,
		"ISSUPPORT": false,
	})
}
