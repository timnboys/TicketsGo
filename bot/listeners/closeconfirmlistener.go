package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"sync"
)

// message ID -> closer
var (
	pendingConfirmations = make(map[uint64]uint64)
	confirmLock sync.RWMutex
)

func OnCloseConfirm(s *gateway.Shard, e *events.MessageReactionAdd) {
	// Check reaction is a ✅ and not from a bot
	if e.Member.User.Bot || e.Emoji.Name != "✅" {
		return
	}

	confirmLock.Lock()
	closer, ok := pendingConfirmations[e.MessageId]
	delete(pendingConfirmations, e.MessageId)
	confirmLock.Unlock()

	if !ok {
		return
	}

	// Verify it's the same user reacting
	if closer != e.UserId {
		return
	}

	// Create error context for later
	errorContext := sentry.ErrorContext{
		Guild:   e.GuildId,
		User:    e.UserId,
		Channel: e.ChannelId,
		Shard:   s.ShardId,
	}

	// Get whether the guild is premium
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
