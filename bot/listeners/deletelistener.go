package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnChannelDelete(s *gateway.Shard, e *events.ChannelDelete) {
	if err := database.Client.Tickets.CloseByChannel(e.Id); err != nil {
		sentry.Error(err)
	}
}
