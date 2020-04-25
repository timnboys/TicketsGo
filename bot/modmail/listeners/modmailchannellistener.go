package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
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

	sessionChan := make(chan *modmaildatabase.ModMailSession, 0)
	go modmaildatabase.GetModMailSessionByStaffChannel(e.ChannelId, sessionChan)
	session := <-sessionChan

	if session == nil {
		return
	}

	// TODO: Make this less hacky
	// check close
	if isClose, args := isClose(e); isClose {
		modmail.HandleClose(session, utils.CommandContext{
			Shard:       s,
			Message:     e.Message,
			Root:        "close",
			Args:        args,
			IsPremium:   false,
			ShouldReact: true,
			IsFromPanel: false,
		})
		return
	}

	// Make sure we don't mirror the user's message back to them
	var username string
	if user, found := s.Cache.GetUser(session.User); found {
		username = user.Username
	}

	// TODO: Make this less hacky
	if e.Author.Username == username && e.WebhookId != 0 {
		return
	}

	// Create DM channel
	privateMessageChannel, err := s.CreateDM(session.User)
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
	ch := make(chan string)
	go database.GetPrefix(e.GuildId, ch)

	customPrefix := <-ch
	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(e.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if strings.HasPrefix(e.Content, customPrefix) {
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
