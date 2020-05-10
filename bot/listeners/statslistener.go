package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
)

func OnFirstResponse(shard *gateway.Shard, e *events.MessageCreate) {
	// Make sure this is a guild
	if e.GuildId == 0 || e.Author.Id == shard.SelfId() {
		return
	}

	e.Member.User = e.Author

	// Only count replies from support reps
	permLevel := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(shard, e.Member, e.GuildId, permLevel)
	if <-permLevel > 0 {
		// Retrieve ticket struct
		ticket, err := database.Client.Tickets.GetByChannel(e.ChannelId)
		if err != nil {
			sentry.Error(err)
			return
		}

		// Make sure that the channel is a ticket
		if ticket.UserId != 0 {
			// We don't have to check for previous responses due to ON CONFLICT DO NOTHING
			if err := database.Client.FirstResponseTime.Set(ticket.GuildId, e.Author.Id, ticket.Id, time.Now().Sub(ticket.OpenTime)); err != nil {
				sentry.Error(err)
			}
		}
	}
}
