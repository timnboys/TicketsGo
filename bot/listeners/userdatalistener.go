package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnUserUpdate(_ *discordgo.Session, e *discordgo.UserUpdate) {
	id, err := strconv.ParseInt(e.User.ID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	go database.UpdateUser(id, e.Username, e.Discriminator, e.Avatar)
}

func OnUserJoin(_ *discordgo.Session, e *discordgo.GuildMemberAdd) {
	id, err := strconv.ParseInt(e.User.ID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	go database.UpdateUser(id, e.User.Username, e.User.Discriminator, e.User.Avatar)
}
