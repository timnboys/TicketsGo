package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnChannelDelete(_ *discordgo.Session, e *discordgo.ChannelDelete) {
	channelId, err := strconv.ParseInt(e.ID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	isTicket := make(chan bool)
	go database.IsTicketChannel(channelId, isTicket)

	if <-isTicket {
		go database.CloseByChannel(channelId)
	}
}
