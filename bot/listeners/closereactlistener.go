package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"time"
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

	closeConfirmation, err := database.Client.CloseConfirmation.Get(e.GuildId); if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	// Get whether the guild is premium
	isPremium := make(chan bool)
	go utils.IsPremiumGuild(s, e.GuildId, isPremium)

	if closeConfirmation {
		// Remove reaction
		_ = s.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, e.Emoji.Name) // Error is probably a 403, we can ignore

		// Send confirmation message
		msg, err := utils.SendEmbedWithResponse(s, e.ChannelId, utils.Green, "Close Confirmation", "React with ✅ to confirm you want to close the ticket", nil, 10, <-isPremium)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		confirmLock.Lock()
		pendingConfirmations[msg.Id] = e.UserId
		confirmLock.Unlock()

		// Add reaction
		// Error is likely a 403, we can ignore - user can add their own reaction
		_ = s.CreateReaction(e.ChannelId, msg.Id, "✅")

		// timeout later
		go func() {
			time.Sleep(time.Second * 10)

			confirmLock.Lock()
			delete(pendingConfirmations, msg.Id)
			confirmLock.Unlock()
		}()
	} else {
		// No need to remove the reaction since we're deleting the channel anyway

		// Get the member object
		member, err := s.GetGuildMember(e.GuildId, e.UserId)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		logic.CloseTicket(s, e.GuildId, e.ChannelId, 0, member, nil, true, <-isPremium)
	}
}
