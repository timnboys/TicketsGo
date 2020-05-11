package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnCloseReact(s *gateway.Shard, e *events.MessageReactionAdd) {
	// Check the right emoji has been used
	if e.Emoji.Name != "🔒" {
		return
	}

	// Create error context for later
	errorContext := sentry.ErrorContext{
		Guild:   e.GuildId,
		User:    e.UserId,
		Channel: e.ChannelId,
		Shard:   s.ShardId,
	}

	// In DMs
	if e.GuildId == 0 {
		return
	}

	// Get user object
	user, err := s.GetUser(e.UserId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Ensure that the user is an actual user, not a bot
	if user.Bot {
		return
	}

	// Get the ticket properties
	ticket, err := database.Client.Tickets.GetByChannel(e.ChannelId); if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Check that this channel is a ticket channel
	if ticket.GuildId == 0 {
		return
	}

	// Check that the ticket has a welcome message
	if ticket.WelcomeMessageId == nil {
		return
	}

	// Check that the message being reacted to is the welcome message
	if e.MessageId != *ticket.WelcomeMessageId {
		return
	}

	// No need to remove the reaction since we're deleting the channel anyway

	// Get whether the guild is premium
	// TODO: Check whether we actually need this
	isPremium := make(chan bool)
	go utils.IsPremiumGuild(s, e.GuildId, isPremium)

	// Get the member object
	member, err := s.GetGuildMember(e.GuildId, e.UserId)
	if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	logic.CloseTicket(s, e.GuildId, e.ChannelId, 0, member, nil, true, <-isPremium)
}
