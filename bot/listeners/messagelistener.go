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

func OnMessage(s *gateway.Shard, e *events.MessageCreate) {
	go statsd.IncrementKey(statsd.MESSAGES)

	// Get guild obj
	guild, err := s.GetGuild(e.GuildId); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   e.GuildId,
			User:    e.Author.Id,
			Channel: e.ChannelId,
			Shard:   s.ShardId,
		})
		return
	}

	premiumChan := make(chan bool)
	go utils.IsPremiumGuild(utils.CommandContext{
		Shard:   s,
		Guild:   guild,
	}, premiumChan)

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
