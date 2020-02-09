package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnPanelReact(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	errorContext := sentry.ErrorContext{
		Guild:   e.GuildID,
		User:    e.UserID,
		Channel: e.ChannelID,
		Shard:   s.ShardID,
	}

	msgId, err := strconv.ParseInt(e.MessageID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	userId, err := strconv.ParseInt(e.UserID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// In DMs
	if e.GuildID == "" {
		return
	}

	guildId, err := strconv.ParseInt(e.GuildID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Get panel from DB
	panelChan := make(chan database.Panel)
	go database.GetPanelByMessageId(msgId, panelChan)
	panel := <-panelChan
	blank := database.Panel{}

	if panel != blank {
		emoji := e.Emoji.Name // This is the actual unicode emoji (https://discordapp.com/developers/docs/resources/emoji#emoji-object-gateway-reaction-standard-emoji-example)

		// Check the right emoji ahs been used
		if panel.ReactionEmote != emoji && !(panel.ReactionEmote == "" && emoji == "📩") {
			return
		}

		user, err := s.User(e.UserID)
		if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		if user.Bot {
			return
		}

		// Remove the reaction from the message
		if err = s.MessageReactionRemove(e.ChannelID, e.MessageID, emoji, e.UserID); err != nil {
			sentry.LogWithContext(err, errorContext)
		}

		blacklisted := make(chan bool)
		go database.IsBlacklisted(guildId, userId, blacklisted)
		if <-blacklisted {
			return
		}

		// Get guild obj
		guild, err := s.State.Guild(e.GuildID)
		if err != nil {
			guild, err = s.Guild(e.GuildID)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
				return
			}
		}

		isPremium := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Session: s,
			GuildId: guildId,
			Guild:   guild,
		}, isPremium)

		member, err := s.State.Member(e.GuildID, e.UserID)
		if err != nil {
			member, err = s.GuildMember(e.GuildID, e.UserID)
			if err != nil {
				sentry.LogWithContext(err, errorContext)
				return
			}
		}

		channelId, err := strconv.ParseInt(e.ChannelID, 10, 64)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}

		ctx := utils.CommandContext{
			Session:     s,
			User:        *user,
			UserID:      userId,
			Guild:       guild,
			GuildId:     guildId,
			Channel:     e.ChannelID,
			ChannelId:   channelId,
			MessageId:   msgId,
			Root:        "new",
			Args:        make([]string, 0),
			IsPremium:   <-isPremium,
			ShouldReact: false,
			Member:      member,
			IsFromPanel: true,
		}

		go command.OpenCommand{}.Execute(ctx)
	}
}
