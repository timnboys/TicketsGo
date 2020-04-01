package listeners

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func OnDirectMessage(s *gateway.Shard, e *events.MessageCreate) {
	go func() {
		if e.Author.Bot {
			return
		}

		if e.GuildId != 0 { // DMs only
			return
		}

		ctx := utils.CommandContext{
			Shard:       s,
			User:        e.Author,
			Message:     e.Message,
			IsPremium:   false,
			ShouldReact: true,
			Member:      e.Member,
		}

		sessionChan := make(chan *modmaildatabase.ModMailSession, 0)
		go modmaildatabase.GetModMailSession(e.Author.Id, sessionChan)
		session := <-sessionChan

		// Create DM channel
		dmChannel, err := s.CreateDM(e.Author.Id); if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext()) // User probably has DMs disabled
			return
		}

		// No active session
		if session == nil {
			guildsChan := make(chan []modmailutils.UserGuild)
			go modmailutils.GetMutualGuilds(ctx.User.Id, guildsChan)
			guilds := <-guildsChan

			if len(e.Message.Content) == 0 {
				modmailutils.SendModMailIntro(ctx, dmChannel.Id)
				return
			}

			split := strings.Split(e.Message.Content, " ")

			targetGuildId, err := strconv.Atoi(split[0])
			if err != nil || targetGuildId < 1 || targetGuildId > len(guilds) + 1 {
				modmailutils.SendModMailIntro(ctx, dmChannel.Id)
				return
			}

			targetGuild := guilds[targetGuildId - 1]
			staffChannel, err := modmail.OpenModMailTicket(s, targetGuild, e.Author, userId)
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
		} else { // Forward message to guild or handle command
			// Create fake guild object
			targetGuild := strconv.Itoa(int(session.Guild))
			guildChan := make(chan discordgo.Guild)
			go createFakeGuild(targetGuild, guildChan)
			guild := <-guildChan

			// Determine whether premium guild
			premiumChan := make(chan bool)
			go utils.IsPremiumGuild(utils.CommandContext{
				GuildId: session.Guild,
				Guild:   &guild,
			}, premiumChan)
			isPremium := <-premiumChan

			// Update context
			ctx.Guild = &guild
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
				sendMessage(session, ctx, dmChannel.ID)
			}
		}
	}()
}

func sendMessage(session *modmaildatabase.ModMailSession, ctx utils.CommandContext, dmChannel string) {
	channel := strconv.Itoa(int(session.StaffChannel))

	// Preferably send via a webhook
	webhookChan := make(chan *string)
	go database.GetWebhookByUuid(session.Uuid, webhookChan)
	webhook := <-webhookChan

	success := false
	if webhook != nil {
		success = executeWebhook(session.Uuid, *webhook, ctx.Message.ContentWithMentionsReplaced(), ctx.User.Username, ctx.User.AvatarURL("256"))
	}

	if !success {
		if _, err := ctx.Shard.ChannelMessageSend(channel, ctx.Message.ContentWithMentionsReplaced()); err != nil {
			utils.SendEmbed(ctx.Shard, dmChannel, utils.Red, "Error", fmt.Sprintf("An error has occurred: `%s`", err.Error()), 30, ctx.IsPremium)
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	// forward attachments
	// don't re-upload attachments incase user has uploaded TOS breaking attachment
	if len(ctx.Message.Attachments) > 0 {
		var content string
		if len(ctx.Message.Attachments) == 1 {
			content = fmt.Sprintf("%s attached a file:", ctx.User.Mention())
		} else {
			content = fmt.Sprintf("%s attached files:", ctx.User.Mention())
		}

		for _, attachment := range ctx.Message.Attachments {
			content += fmt.Sprintf("\n▶️ %s", attachment.ProxyURL)
		}

		if _, err := ctx.Shard.ChannelMessageSend(channel, content); err != nil {
			utils.SendEmbed(ctx.Shard, dmChannel, utils.Red, "Error", fmt.Sprintf("An error has occurred: `%s`", err.Error()), 30, ctx.IsPremium)
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}
}

func executeWebhook(uuid, webhook, content, username, avatarUrl string) bool {
	body := map[string]interface{}{
		"content":    content,
		"username":   username,
		"avatar_url": avatarUrl,
	}
	encoded, err := json.Marshal(&body); if err != nil {
		return false
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("https://canary.discordapp.com/api/webhooks/%s", webhook), bytes.NewBuffer(encoded)); if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	client.Timeout = 3 * time.Second

	res, err := client.Do(req); if err != nil {
		return false
	}

	if res.StatusCode == 404 || res.StatusCode == 403 {
		go database.DeleteWebhookByUuid(uuid)
	} else {
		return true
	}

	return false
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

func createFakeGuild(id string, res chan discordgo.Guild) {
	name := make(chan string)
	go cache.Client.GetGuildName(id, name)

	owner := make(chan string)
	go cache.Client.GetGuildOwner(id, owner)

	guild := discordgo.Guild{
		ID:      id,
		Name:    <-name,
		OwnerID: <-owner,
	}

	res <-guild
}
