package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
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

		sessionChan := make(chan *modmaildatabase.ModMailSession, 0)
		go modmaildatabase.GetModMailSession(userId, sessionChan)
		session := <-sessionChan

		// Create DM channel
		dmChannel, err := s.UserChannelCreate(e.Author.ID); if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext()) // User probably has DMs disabled
			return
		}

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
				utils.SendEmbed(s, dmChannel.ID, utils.Green, "Modmail", fmt.Sprintf("Your modmail ticket in %s has been opened! Use `t!close` to close the session.", targetGuild.Name), 0, true)

				// Send guild's welcome message
				welcomeMessageChan := make(chan string)
				go database.GetWelcomeMessage(targetGuild.Id, welcomeMessageChan)
				welcomeMessage := <-welcomeMessageChan

				utils.SendEmbed(s, dmChannel.ID, utils.Green, "Modmail", welcomeMessage, 0, true)
			} else {
				utils.SendEmbed(s, dmChannel.ID, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), 30, true)
			}
		} else { // Forward message to guild or handle command
			// Get guild object
			guild, err := s.State.Guild(e.GuildID); if err != nil { // Not cached
				guild, err = s.Guild(e.GuildID); if err != nil { // TODO: Guild may have been deleted. Handle this better
					utils.SendEmbed(s, dmChannel.ID, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), 30, true)
					return
				}
			}

			// Determine whether premium guild
			premiumChan := make(chan bool)
			go utils.IsPremiumGuild(utils.CommandContext{
				GuildId: session.Guild,
				Guild:   guild,
			}, premiumChan)
			isPremium := <-premiumChan

			// Update context
			ctx.Guild = guild
			ctx.GuildId = session.Guild
			ctx.IsPremium = isPremium
			ctx.Channel = dmChannel.ID

			// Parse DM channel ID
			dmChannelId, err := strconv.ParseInt(dmChannel.ID, 10, 64); if err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
				return
			}
			ctx.ChannelId = dmChannelId

			var isCommand bool
			ctx, isCommand = handleCommand(ctx, session)

			if isCommand {
				switch ctx.Root {
				case "close": modmail.HandleClose(session, ctx)
				}
			} else {
				channel := strconv.Itoa(int(session.StaffChannel))
				if _, err := s.ChannelMessageSend(channel, e.Message.ContentWithMentionsReplaced()); err != nil {
					utils.SendEmbed(s, dmChannel.ID, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), 30, isPremium)
					sentry.LogWithContext(err, ctx.ToErrorContext())
				}
			}
		}
	}()
}

// TODO: Make this less hacky
func handleCommand(ctx utils.CommandContext, session *modmaildatabase.ModMailSession) (utils.CommandContext, bool) {
	prefixChannel := make(chan string)
	go database.GetPrefix(session.Guild, prefixChannel)
	customPrefix := <-prefixChannel

	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(ctx.Message.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if strings.HasPrefix(ctx.Message.Content, customPrefix) {
		usedPrefix = customPrefix
	} else { // Not a command
		return ctx, false
	}

	split := strings.Split(ctx.Message.Content, " ")
	root := strings.TrimPrefix(split[0], usedPrefix)

	args := make([]string, 0)
	if len(split) > 1 {
		for _, arg := range split[1:] {
			if arg != "" {
				args = append(args, arg)
			}
		}
	}

	ctx.Args = args
	ctx.Root = root

	return ctx, true
}
