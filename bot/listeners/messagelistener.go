package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
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
		ticket, err := database.Client.Tickets.GetByChannel(e.ChannelId)
		if err != nil {
			sentry.Error(err)
			return
		}

		// Verify that this is a ticket
		if ticket.UserId != 0 {
			go cache.Client.PublishMessage(cache.TicketMessage{
				GuildId:  e.GuildId,
				TicketId: ticket.Id,
				Username: e.Author.Username,
				Content:  e.Content,
			})
		}
	}
}
