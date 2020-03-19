package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnDirectMessageReact(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	go func() {
		if e.GuildID != "" { // DMs only
			return
		}

		userId, err := strconv.ParseInt(e.UserID, 10, 64); if err != nil {
			sentry.Error(err)
			return
		}

		sessionChan := make(chan *modmaildatabase.ModMailSession, 0)
		go modmaildatabase.GetModMailSession(userId, sessionChan)
		session := <-sessionChan

		if session != nil {
			return
		}

		// Determine which emoji was used
		reaction := -1
		for i, emoji := range modmailutils.Emojis {
			if emoji == e.Emoji.Name {
				reaction = i
				break
			}
		}

		// Check a number emoji was used
		if reaction == -1 {
			return
		}

		// Remove reaction
		_ = s.MessageReactionRemove(e.ChannelID, e.MessageID, e.Emoji.ID, e.UserID)

		// Determine which guild we should open the channel in
		guildsChan := make(chan []modmailutils.UserGuild)
		go modmailutils.GetMutualGuilds(userId, guildsChan)
		guilds := <-guildsChan
		targetGuild := guilds[reaction - 1]

		// Create DM channel
		dmChannel, err := s.UserChannelCreate(e.UserID); if err != nil {
			// TODO: Error logging
			return
		}

		// Get user object
		user, err := s.User(e.UserID); if err != nil {
			sentry.Error(err)
			return
		}

		staffChannel, err := modmail.OpenModMailTicket(s, targetGuild, user, userId)
		if err == nil {
			utils.SendEmbed(s, dmChannel.ID, utils.Green, "Modmail", fmt.Sprintf("Your modmail ticket in %s has been opened! Use `t!close` to close the session.", targetGuild.Name), 0, true)

			// Send guild's welcome message
			welcomeMessageChan := make(chan string)
			go database.GetWelcomeMessage(targetGuild.Id, welcomeMessageChan)
			welcomeMessage := <-welcomeMessageChan

			utils.SendEmbed(s, dmChannel.ID, utils.Green, "Modmail", welcomeMessage, 0, true)
			utils.SendEmbed(s, staffChannel, utils.Green, "Modmail", welcomeMessage, 0, true)
		} else {
			utils.SendEmbed(s, dmChannel.ID, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), 30, true)
		}
	}()
}
