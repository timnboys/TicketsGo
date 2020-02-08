package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnChannelDelete(s *discordgo.Session, e *discordgo.ChannelDelete) {
	channelId, err := strconv.ParseInt(e.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   e.GuildID,
			Channel: e.Channel.ID,
			Shard:   s.ShardID,
		})
		return
	}

	isTicket := make(chan bool)
	go database.IsTicketChannel(channelId, isTicket)

	if <-isTicket {
		go database.CloseByChannel(channelId)
	}

	go database.DeleteChannel(channelId)
}
