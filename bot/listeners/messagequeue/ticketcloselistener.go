package messagequeue

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"strings"
)

func ListenTicketClose(shardManager *gateway.ShardManager) {
	closes := make(chan cache.TicketCloseMessage)
	go cache.Client.ListenTicketClose(closes)

	for payload := range closes {
		// Get the ticket properties
		ticketChan := make(chan database.Ticket)
		go database.GetTicketByUuid(payload.Uuid, ticketChan)
		ticket := <-ticketChan

		// Check that this is a valid ticket
		if ticket.Uuid == "" {
			return
		}

		// Get session
		s := shardManager.ShardForGuild(ticket.Guild)
		if s == nil { // Not on this cluster
			continue
		}

		// Create error context for later
		errorContext := sentry.ErrorContext{
			Guild: ticket.Guild,
			User:  payload.User,
			Shard: s.ShardId,
		}

		// Get whether the guild is premium
		// TODO: Check whether we actually need this
		isPremium := make(chan bool)
		go utils.IsPremiumGuild(s, ticket.Guild, isPremium)

		// Get the member object
		member, err := s.GetGuildMember(ticket.Guild, payload.User)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		// Add reason to args
		reason := strings.Split(payload.Reason, " ")

		if ticket.Channel == nil {
			return
		}
		logic.CloseTicket(s, ticket.Guild, *ticket.Channel, 0, member, reason, false, <-isPremium)
	}
}
