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
	msgId, err := strconv.ParseInt(e.MessageID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	userId, err := strconv.ParseInt(e.UserID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	guildId, err := strconv.ParseInt(e.GuildID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	isPanel := make(chan bool)
	go database.IsPanel(msgId, isPanel)
	if <-isPanel {
		user, err := s.User(e.UserID); if err != nil {
			sentry.Error(err)
			return
		}

		if user.Bot {
			return
		}

		if err = s.MessageReactionRemove(e.ChannelID, e.MessageID, "📩", e.UserID); err != nil {
			sentry.Error(err)
		}

		blacklisted := make(chan bool)
		go database.IsBlacklisted(guildId, userId, blacklisted)
		if <-blacklisted {
			return
		}

		msg, err := s.ChannelMessage(e.ChannelID, e.MessageID); if err != nil {
			sentry.Error(err)
			return
		}

		isPremium := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Session: s,
			Guild: e.GuildID,
			GuildId: guildId,
		}, isPremium)

		ctx := utils.CommandContext{
			Session: s,
			User: *user,
			UserID: userId,
			Guild: e.GuildID,
			GuildId: guildId,
			Channel: e.ChannelID,
			Message: *msg,
			Root: "new",
			Args: make([]string, 0),
			IsPremium: <-isPremium,
			ShouldReact: false,
		}

		command.OpenCommand{}.Execute(ctx)
	}
}
