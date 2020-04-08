package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// proxy messages to web UI
func OnMessage(s *gateway.Shard, e *events.MessageCreate) {
	go statsd.IncrementKey(statsd.MESSAGES)

	// ignore DMs
	if e.GuildId == 0 {
		return
	}

	premiumChan := make(chan bool)
	go utils.IsPremiumGuild(s, e.GuildId, premiumChan)

	if <-premiumChan {
		isTicket := make(chan bool)
		go database.IsTicketChannel(e.ChannelId, isTicket)
		if <-isTicket {
			ticket := make(chan int)
			go database.GetTicketId(e.ChannelId, ticket)

			go cache.Client.PublishMessage(cache.TicketMessage{
				GuildId:  e.GuildId,
				TicketId: <-ticket,
				Username: e.Author.Username,
				Content:  e.Content,
			})
		}
	}
}
