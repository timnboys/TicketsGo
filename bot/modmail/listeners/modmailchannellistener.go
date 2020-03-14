package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail/database"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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

		sessionChan := make(chan *database.ModMailSession, 0)
		go database.GetModMailSessionByStaffChannel(channelId, sessionChan)
		session := <-sessionChan

		if session == nil {
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
