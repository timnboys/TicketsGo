package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnChannelDelete(s *gateway.Shard, e *events.ChannelDelete) {
	isTicket := make(chan bool)
	go database.IsTicketChannel(e.Channel.Id, isTicket)

	if <-isTicket {
		go database.CloseByChannel(e.Channel.Id)
	}

	go database.DeleteChannel(e.Channel.Id)
}
