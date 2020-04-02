package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
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
	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(e.ChannelId, ticketChan)
	ticket := <-ticketChan

	// Check that this channel is a ticket channel
	if ticket.Uuid == "" {
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

	// No need to remove the reaction since we'ere deleting the channel anyway

	// Get guild obj
	guild, err := s.GetGuild(e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Get whether the guild is premium
	// TODO: Check whether we actually need this
	isPremium := make(chan bool)
	go utils.IsPremiumGuild(utils.CommandContext{
		Shard: s,
		Guild: guild,
	}, isPremium)

	// Get the member object
	member, err := s.GetGuildMember(e.GuildId, e.UserId)
	if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	ctx := utils.CommandContext{
		Shard:       s,
		User:        user,
		Guild:       guild,
		ChannelId:   e.ChannelId,
		Root:        "close",
		Args:        make([]string, 0),
		IsPremium:   <-isPremium,
		ShouldReact: false,
		Member:      member,
	}

	go command.CloseCommand{}.Execute(ctx)
}
