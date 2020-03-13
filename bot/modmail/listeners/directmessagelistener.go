package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	"github.com/TicketsBot/TicketsGo/bot/modmail/database"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnDirectMessage(s *discordgo.Session, e *discordgo.MessageCreate) {
	go func() {
		if e.Author.Bot {
			return
		}

		if e.GuildID != "" { // DMs only
			return
		}

		userId, err := strconv.ParseInt(e.Author.ID, 10, 64); if err != nil {
			sentry.Error(err)
			return
		}

		messageId, err := strconv.ParseInt(e.Message.ID, 10, 64); if err != nil {
			sentry.Error(err)
			return
		}

		ctx := utils.CommandContext{
			Session:     s,
			User:        *e.Author,
			UserID:      userId,
			Message:     *e.Message,
			MessageId:   messageId,
			IsPremium:   false,
			ShouldReact: true,
			Member:      e.Member,
		}

		// Create DM channel
		channel, err := s.UserChannelCreate(e.Author.ID); if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext()) // User probably has DMs disabled
			return
		}
		ctx.Channel = channel.ID

		sessionChan := make(chan *database.ModMailSession, 0)
		go database.GetModMailSession(userId, sessionChan)
		session := <-sessionChan

		// No active session
		if session == nil {
			guildsChan := make(chan []modmailutils.UserGuild)
			go modmailutils.GetMutualGuilds(ctx.UserID, guildsChan)
			guilds := <-guildsChan

			if len(ctx.Args) == 0 {
				modmailutils.SendModMailIntro(ctx)
				return
			}

			targetGuildId, err := strconv.Atoi(ctx.Args[0])
			if err != nil || targetGuildId < 1 || targetGuildId > len(guilds) + 1 {
				modmailutils.SendModMailIntro(ctx)
				return
			}

			targetGuild := guilds[targetGuildId + 1]
			err = modmail.OpenModMailTicket(s, targetGuild, e.Author, userId)
			if err == nil {
				ctx.SendEmbedNoDelete(utils.Green, "Modmail", fmt.Sprintf("Your modmail ticket in %s has been opened!", targetGuild.Name))
			} else {
				ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()))
			}
		}
	}()
}
