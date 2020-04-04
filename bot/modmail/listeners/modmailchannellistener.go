package listeners

import (
	"fmt"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnModMailChannelMessage(s *gateway.Shard, e *events.MessageCreate) {
	go func() {
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

		// Make sure we don't mirror the user's message back to them
		// Get username
		usernameChan := make(chan string)
		go database.GetUsername(session.User, usernameChan)
		username := <-usernameChan

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
	}()
}
