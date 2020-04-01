package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnFirstResponse(shard *gateway.Shard, e *events.MessageCreate) {
	// Make sure this is a guild
	if e.GuildId == 0 || e.Member == nil {
		return
	}

	e.Member.User = e.Author

	// Only count replies from support reps
	permLevel := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(shard, e.Member, e.GuildId, permLevel)
	if <-permLevel > 0 {
		// Make sure that the channel is a ticket
		isTicket := make(chan bool)
		go database.IsTicketChannel(e.ChannelId, isTicket)

		if <-isTicket {
			uuidChan := make(chan string)
			go database.GetTicketUuid(e.ChannelId, uuidChan)
			uuid := <-uuidChan

			// Make sure this is the first response
			hasResponse := make(chan bool)
			go database.HasResponse(uuid, hasResponse)
			if !<-hasResponse {
				go database.AddResponseTime(uuid, e.GuildId, e.Author.Id)
			}
		}
	}
}
