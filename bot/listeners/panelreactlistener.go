package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel/message"
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

	// Get panel from DB
	panelChan := make(chan database.Panel)
	go database.GetPanelByMessageId(e.MessageId, panelChan)
	panel := <-panelChan
	blank := database.Panel{}

	if panel != blank {
		emoji := e.Emoji.Name // This is the actual unicode emoji (https://discordapp.com/developers/docs/resources/emoji#emoji-object-gateway-reaction-standard-emoji-example)

		// Check the right emoji ahs been used
		if panel.ReactionEmote != emoji && !(panel.ReactionEmote == "" && emoji == "📩") {
			return
		}

		user, err := s.GetUser(e.UserId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		if user.Bot {
			return
		}

		// Remove the reaction from the message
		if err = s.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, emoji); err != nil {
			sentry.LogWithContext(err, errorContext)
		}

		blacklisted := make(chan bool)
		go database.IsBlacklisted(e.GuildId, e.UserId, blacklisted)
		if <-blacklisted {
			return
		}

		// Get guild obj
		guild, err := s.GetGuild(e.GuildId)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		isPremium := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Shard: s,
			Guild: &guild,
		}, isPremium)

		member, err := s.GetGuildMember(e.GuildId, e.UserId)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		// construct fake message
		panelMessage := message.Message{
			Id:        e.MessageId,
			ChannelId: e.ChannelId,
			GuildId:   e.GuildId,
		}

		ctx := utils.CommandContext{
			Shard:       s,
			User:        &user,
			Guild:       &guild,
			ChannelId:   e.ChannelId,
			Message:     &panelMessage,
			Root:        "new",
			Args:        make([]string, 0),
			IsPremium:   <-isPremium,
			ShouldReact: false,
			Member:      &member,
			IsFromPanel: true,
		}

		go command.OpenCommand{}.Execute(ctx)
	}
}
