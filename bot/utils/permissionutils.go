package utils

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
)

func GetPermissionLevel(session *discordgo.Session, guild string, user string, ch chan PermissionLevel) {
	// Check if user is guild owner
	g, err := session.State.Guild(guild); if err != nil {
		// Not cached
		g, err = session.Guild(guild)
		if err != nil {
			log.Error(err.Error())
		}
	}

	if g != nil {
		if user == g.OwnerID {
			ch <- Admin
			return
		}
	}

	admin := make(chan bool)
	go database.IsAdmin(guild, user, admin)
	if <- admin {
		ch <- Admin
		return
	}

	support := make(chan bool)
	go database.IsSupport(guild, user, support)
	if <- support {
		ch <- Support
		return
	}

	ch <- Everyone
}
