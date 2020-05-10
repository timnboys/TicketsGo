package listeners

import (
	"context"
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel/message"
	"golang.org/x/sync/errgroup"
)

func OnPanelReact(s *gateway.Shard, e *events.MessageReactionAdd) {
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

	if e.UserId == s.SelfId() {
		return
	}

	// Get panel from DB
	panel, err := database.Client.Panel.Get(e.MessageId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Verify this is a panel
	if panel.MessageId != 0 {
		emoji := e.Emoji.Name // This is the actual unicode emoji (https://discordapp.com/developers/docs/resources/emoji#emoji-object-gateway-reaction-standard-emoji-example)

		// Check the right emoji ahs been used
		if panel.ReactionEmote != emoji && !(panel.ReactionEmote == "" && emoji == "📩") {
			return
		}

		// TODO: Check perms
		// Remove the reaction from the message
		if err := s.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, emoji); err != nil {
			sentry.LogWithContext(err, errorContext)
		}

		var blacklisted, premium bool

		group, _ := errgroup.WithContext(context.Background())

		// get blacklisted
		group.Go(func() (err error) {
			blacklisted, err = database.Client.Blacklist.IsBlacklisted(e.GuildId, e.UserId)
			return
		})

		// get premium
		group.Go(func() error {
			ch := make(chan bool)
			go utils.IsPremiumGuild(s, e.GuildId, ch)
			premium = <-ch
			return nil
		})

		if err := group.Wait(); err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		if blacklisted {
			return
		}

		// construct fake message
		messageReference := message.MessageReference{
			ChannelId: e.ChannelId,
			GuildId:   e.GuildId,
		}

		// get user object
		user, err := s.GetUser(e.UserId); if err != nil {
			sentry.Error(err)
			return
		}

		go logic.OpenTicket(s, user, messageReference, premium, nil, &panel)
	}
}
