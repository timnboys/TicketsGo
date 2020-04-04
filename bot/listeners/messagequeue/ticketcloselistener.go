package messagequeue

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
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

		// Get guild obj
		guild, err := s.GetGuild(ticket.Guild)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		// Get whether the guild is premium
		// TODO: Check whether we actually need this
		isPremium := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Shard: s,
			Guild: &guild,
		}, isPremium)

		// Get the member object
		member, err := s.GetGuildMember(ticket.Guild, payload.User)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		// Add reason to args
		reason := strings.Split(payload.Reason, " ")

		ctx := utils.CommandContext{
			Shard:       s,
			User:        &member.User,
			Guild:       &guild,
			ChannelId:   *ticket.Channel,
			Message:     nil,
			Root:        "close",
			Args:        reason,
			IsPremium:   <-isPremium,
			ShouldReact: false,
			Member:      &member,
		}

		go command.CloseCommand{}.Execute(ctx)
	}
}
