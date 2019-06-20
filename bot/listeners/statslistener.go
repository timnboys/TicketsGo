package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnFirstResponse(s *discordgo.Session, e *discordgo.MessageCreate) {
	channelId, err := strconv.ParseInt(e.ChannelID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	guildId, err := strconv.ParseInt(e.GuildID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	userId, err := strconv.ParseInt(e.Author.ID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	// Only count replies from support reps
	permLevel := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(s, e.GuildID, e.Author.ID, permLevel)
	if <-permLevel > 0 {
		// Make sure that the channel is a ticket
		isTicket := make(chan bool)
		go database.IsTicketChannel(channelId, isTicket)

		if <-isTicket {
			uuidChan := make(chan string)
			go database.GetTicketUuid(channelId, uuidChan)
			uuid := <-uuidChan

			// Make sure this is the first response
			hasResponse := make(chan bool)
			go database.HasResponse(uuid, hasResponse)
			if !<-hasResponse {
				go database.AddResponseTime(uuid, guildId, userId)
			}
		}
	}
}
