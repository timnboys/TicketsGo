package utils

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/bwmarrin/discordgo"
)

func GetPermissionLevel(session *discordgo.Session, guild string, user string, ch chan PermissionLevel) {
	if g, err := session.Guild(guild); err == nil {
		if user == g.OwnerID {
			ch <- Admin
			return
		}
	}

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
