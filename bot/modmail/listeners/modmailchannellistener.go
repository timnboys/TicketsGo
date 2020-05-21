package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"strings"
)

func OnModMailChannelMessage(s *gateway.Shard, e *events.MessageCreate) {
	if e.Author.Id == s.SelfId() {
		return
	}

	if e.GuildId == 0 { // Guilds only
		return
	}

	errorContext := sentry.ErrorContext{
		Guild:   e.GuildId,
		Channel: e.ChannelId,
		Shard:   s.ShardId,
		User:    e.Author.Id,
	}

	session, err := database.Client.ModmailSession.GetByChannel(e.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	if session.UserId == 0 {
		return
	}

	// TODO: Make this less hacky
	// check close
	if isClose, args := isClose(e); isClose {
		// get permission level
		var permLevel utils.PermissionLevel
		if e.GuildId != 0 {
			ch := make(chan utils.PermissionLevel)
			go utils.GetPermissionLevel(s, e.Member, e.GuildId, ch)
			permLevel = <-ch
		}

		modmail.HandleClose(session, utils.CommandContext{
			Shard:       s,
			Message:     e.Message,
			Root:        "close",
			Args:        args,
			IsPremium:   false,
			ShouldReact: true,
			IsFromPanel: false,
			UserPermissionLevel: permLevel,
		})
		return
	}

	// Make sure we don't mirror the user's message back to them
	var username string
	if user, found := s.Cache.GetUser(session.UserId); found {
		username = user.Username
	}

	// TODO: Make this less hacky
	if e.Author.Username == username && e.WebhookId != 0 {
		return
	}

	// Create DM channel
	privateMessageChannel, err := s.CreateDM(session.UserId)
	if err != nil { // User probably has DMs disabled
		sentry.LogWithContext(err, errorContext)
		return
	}

	message := fmt.Sprintf("**%s**: %s", e.Author.Username, e.Message.Content)
	if _, err := s.CreateMessage(privateMessageChannel.Id, message); err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	// forward attachments
	// don't re-upload attachments incase user has uploaded TOS breaking attachment
	if len(e.Message.Attachments) > 0 {
		var content string
		if len(e.Message.Attachments) == 1 {
			content = fmt.Sprintf("%s attached a file:", e.Author.Mention())
		} else {
			content = fmt.Sprintf("%s attached files:", e.Author.Mention())
		}

		for _, attachment := range e.Message.Attachments {
			content += fmt.Sprintf("\n▶️ %s", attachment.ProxyUrl)
		}

		if _, err := s.CreateMessage(privateMessageChannel.Id, content); err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}
	}
}

// isClose, args
func isClose(e *events.MessageCreate) (bool, []string) {
	customPrefix, err := database.Client.Prefix.Get(e.GuildId); if err != nil {
		sentry.Error(err)
	}

	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(e.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if customPrefix != "" && strings.HasPrefix(e.Content, customPrefix) {
		usedPrefix = customPrefix
	} else { // Not a command
		return false, nil
	}

	split := strings.Split(e.Content, " ")
	root := strings.TrimPrefix(split[0], usedPrefix)

	if strings.ToLower(root) != "close" {
		return false, nil
	}

	args := make([]string, 0)
	if len(split) > 1 {
		for _, arg := range split[1:] {
			if arg != "" {
				args = append(args, arg)
			}
		}
	}

	return true, args
}
