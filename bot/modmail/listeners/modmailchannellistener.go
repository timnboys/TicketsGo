package listeners

import (
	"fmt"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

func OnModMailChannelMessage(s *discordgo.Session, e *discordgo.MessageCreate) {
	go func() {
		if e.Author.ID == utils.Id {
			return
		}

		if e.GuildID == "" { // Guilds only
			return
		}

		errorContext := sentry.ErrorContext{
			Guild:       e.GuildID,
			Channel:     e.ChannelID,
			Shard:       s.ShardID,
		}

		if e.Author != nil {
			errorContext.User = e.Author.ID
		}

		channelId, err := strconv.ParseInt(e.ChannelID, 10, 64); if err != nil {
			sentry.Error(err)
			return
		}

		sessionChan := make(chan *modmaildatabase.ModMailSession, 0)
		go modmaildatabase.GetModMailSessionByStaffChannel(channelId, sessionChan)
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
		if strings.HasPrefix(e.Message.Content, fmt.Sprintf("**%s**: ", username)) {
			return
		}

		// Create DM channel
		privateMessageChannel, err := s.UserChannelCreate(strconv.Itoa(int(session.User))); if err != nil { // User probably has DMs disabled
			sentry.LogWithContext(err, errorContext)
			return
		}

		message := fmt.Sprintf("**%s**: %s", e.Author.Username, e.Message.Content)
		if _, err := s.ChannelMessageSend(privateMessageChannel.ID, message); err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}
	}()
}
