package messagequeue

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	dbclient "github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"strings"
)

func ListenTicketClose(shardManager *gateway.ShardManager) {
	closes := make(chan cache.TicketCloseMessage)
	go cache.Client.ListenTicketClose(closes)

	for payload := range closes {
		// Get the ticket struct
		ticket, err := dbclient.Client.Tickets.Get(payload.TicketId, payload.Guild)
		if err != nil {
			sentry.Error(err)
			continue
		}

		// Check that this is a valid ticket
		if ticket.GuildId == 0 {
			continue
		}

		// Get session
		s := shardManager.ShardForGuild(ticket.GuildId)
		if s == nil { // Not on this cluster
			continue
		}

		// Create error context for later
		errorContext := sentry.ErrorContext{
			Guild: ticket.GuildId,
			User:  payload.User,
			Shard: s.ShardId,
		}

		// Get whether the guild is premium for log archiver
		isPremium := make(chan bool)
		go utils.IsPremiumGuild(s, ticket.GuildId, isPremium)

		// Get the member object
		member, err := s.GetGuildMember(ticket.GuildId, payload.User)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			continue
		}

		// Add reason to args
		reason := strings.Split(payload.Reason, " ")

		if ticket.ChannelId == nil {
			return
		}

		logic.CloseTicket(s, ticket.GuildId, *ticket.ChannelId, 0, member, reason, false, <-isPremium)
	}
}
