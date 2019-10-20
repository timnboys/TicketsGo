package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnMessage(s *discordgo.Session, e *discordgo.MessageCreate) {
	guildId, err := strconv.ParseInt(e.GuildID, 10, 64); if err != nil {
		return
	}

	channelId, err := strconv.ParseInt(e.ChannelID, 10, 64); if err != nil {
		return
	}

	premiumChan := make(chan bool)
	go utils.IsPremiumGuild(utils.CommandContext{
		Session: s,
		Guild: e.GuildID,
		GuildId: guildId,
	}, premiumChan)

	if <-premiumChan {
		isTicket := make(chan bool)
		go database.IsTicketChannel(channelId, isTicket)
		if <-isTicket {
			ticket := make(chan int)
			go database.GetTicketId(channelId, ticket)

			go cache.Client.PublishMessage(cache.TicketMessage{
				GuildId:  e.GuildID,
				TicketId: <-ticket,
				Username: e.Author.Username,
				Content:  e.Content,
			})
		}
	}
}