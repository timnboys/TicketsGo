package utils

import "github.com/TicketsBot/TicketsGo/database"

func GetPermissionLevel(guild string, user string, ch chan PermissionLevel) {
	admin := make(chan bool)
	database.IsAdmin(guild, user, admin)
	if <- admin {
		ch <- Admin
		return
	}

	support := make(chan bool)
	database.IsSupport(guild, user, support)
	if <- support {
		ch <- Support
		return
	}

	ch <- Everyone
}
