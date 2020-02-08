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
	msgId, err := strconv.ParseInt(e.MessageID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   e.GuildID,
			User:    e.UserID,
			Channel: e.ChannelID,
			Shard:   s.ShardID,
		})
		return
	}

	userId, err := strconv.ParseInt(e.UserID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   e.GuildID,
			User:    e.UserID,
			Channel: e.ChannelID,
			Shard:   s.ShardID,
		})
		return
	}

	// In DMs
	if e.GuildID == "" {
		return
	}

	guildId, err := strconv.ParseInt(e.GuildID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   e.GuildID,
			User:    e.UserID,
			Channel: e.ChannelID,
			Shard:   s.ShardID,
		})
		return
	}

	isPanel := make(chan bool)
	go database.IsPanel(msgId, isPanel)
	if <-isPanel {
		user, err := s.User(e.UserID)
		if err != nil {
			sentry.ErrorWithContext(err, sentry.ErrorContext{
				Guild:   e.GuildID,
				User:    e.UserID,
				Channel: e.ChannelID,
				Shard:   s.ShardID,
			})
			return
		}

		if user.Bot {
			return
		}

		if err = s.MessageReactionRemove(e.ChannelID, e.MessageID, "📩", e.UserID); err != nil {
			sentry.LogWithContext(err, sentry.ErrorContext{
				Guild:   e.GuildID,
				User:    e.UserID,
				Channel: e.ChannelID,
				Shard:   s.ShardID,
			})
		}

		blacklisted := make(chan bool)
		go database.IsBlacklisted(guildId, userId, blacklisted)
		if <-blacklisted {
			return
		}

		msg, err := s.ChannelMessage(e.ChannelID, e.MessageID)
		if err != nil {
			sentry.LogWithContext(err, sentry.ErrorContext{
				Guild:   e.GuildID,
				User:    e.UserID,
				Channel: e.ChannelID,
				Shard:   s.ShardID,
			})
			return
		}

		// Get guild obj
		guild, err := s.State.Guild(e.GuildID)
		if err != nil {
			guild, err = s.Guild(e.GuildID)
			if err != nil {
				sentry.ErrorWithContext(err, sentry.ErrorContext{
					Guild:   e.GuildID,
					User:    e.UserID,
					Channel: e.ChannelID,
					Shard:   s.ShardID,
				})
				return
			}
		}

		isPremium := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Session: s,
			GuildId: guildId,
			Guild: guild,
		}, isPremium)

		member, err := s.State.Member(e.GuildID, e.UserID)
		if err != nil {
			member, err = s.GuildMember(e.GuildID, e.UserID)
			if err != nil {
				sentry.LogWithContext(err, sentry.ErrorContext{
					Guild:   e.GuildID,
					User:    e.UserID,
					Channel: e.ChannelID,
					Shard:   s.ShardID,
				})
				return
			}
		}

		channelId, err := strconv.ParseInt(e.ChannelID, 10, 64); if err != nil {
			sentry.LogWithContext(err, sentry.ErrorContext{
				Guild:   e.GuildID,
				User:    e.UserID,
				Channel: e.ChannelID,
				Shard:   s.ShardID,
			})
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
			Message:     *msg,
			Root:        "new",
			Args:        make([]string, 0),
			IsPremium:   <-isPremium,
			ShouldReact: false,
			Member:      member,
		}

		go command.OpenCommand{}.Execute(ctx)
	}
}
