package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	dbclient "github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/TicketsBot/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/rest"
	"github.com/rxdn/gdl/rest/request"
	"strconv"
	"strings"
)

func OnDirectMessage(s *gateway.Shard, e *events.MessageCreate) {
	if e.Author.Bot {
		return
	}

	if e.GuildId != 0 { // DMs only
		return
	}

	ctx := utils.CommandContext{
		Shard:       s,
		Message:     e.Message,
		ShouldReact: true,
		IsFromPanel: false,
	}

	session, err := dbclient.Client.ModmailSession.GetByUser(utils.BOT_ID, e.Author.Id)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Create DM channel
	dmChannel, err := s.CreateDM(e.Author.Id)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext()) // User probably has DMs disabled
		return
	}

	// No active session
	if session.UserId == 0 {
		guilds := modmailutils.GetMutualGuilds(ctx.Shard, ctx.Author.Id)

		if len(e.Message.Content) == 0 {
			modmailutils.SendModMailIntro(ctx, dmChannel.Id)
			return
		}

		split := strings.Split(e.Message.Content, " ")

		targetGuildNumber, err := strconv.Atoi(split[0])
		if err != nil || targetGuildNumber < 1 || targetGuildNumber > len(guilds) {
			modmailutils.SendModMailIntro(ctx, dmChannel.Id)
			return
		}

		targetGuild := guilds[targetGuildNumber-1]

		// Check blacklist
		isBlacklisted, err := dbclient.Client.Blacklist.IsBlacklisted(targetGuild.Id, ctx.Author.Id)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}

		if isBlacklisted {
			utils.SendEmbed(s, dmChannel.Id, utils.Red, "Error", "You are blacklisted in this server!", nil, 30, true)
			return
		}

		utils.SendEmbed(s, dmChannel.Id, utils.Green, "Modmail", fmt.Sprintf("Your modmail ticket in %s has been opened! Use `t!close` to close the session.", targetGuild.Name), nil, 0, true)

		// Send guild's welcome message
		welcomeMessage, err := dbclient.Client.WelcomeMessages.Get(targetGuild.Id)
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
		}

		welcomeMessageId, err := utils.SendEmbedWithResponse(s, dmChannel.Id, utils.Green, "Modmail", welcomeMessage, nil, 0, true)
		if err != nil {
			utils.SendEmbed(s, dmChannel.Id, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), nil, 30, true)
			return
		}

		staffChannel, err := modmail.OpenModMailTicket(s, targetGuild, e.Author, welcomeMessageId.Id)
		if err != nil {
			utils.SendEmbed(s, dmChannel.Id, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), nil, 30, true)
			return
		}

		utils.SendEmbed(s, staffChannel, utils.Green, "Modmail", welcomeMessage, nil, 0, true)
	} else { // Forward message to guild or handle command
		// Determine whether premium guild
		premiumChan := make(chan bool)
		go utils.IsPremiumGuild(s, session.GuildId, premiumChan)
		isPremium := <-premiumChan

		// Update context
		ctx.IsPremium = isPremium
		ctx.ChannelId = dmChannel.Id

		// Parse DM channel ID
		ctx.ChannelId = dmChannel.Id

		var isCommand bool
		ctx, isCommand = handleCommand(ctx, session)

		if isCommand {
			switch ctx.Root {
			case "close":
				modmail.HandleClose(session, ctx)
			}
		} else {
			sendMessage(session, ctx, dmChannel.Id)
		}
	}
}

func sendMessage(session database.ModmailSession, ctx utils.CommandContext, dmChannel uint64) {
	// Preferably send via a webhook
	webhook, err := dbclient.Client.ModmailWebhook.Get(session.Uuid)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	success := false
	if webhook.WebhookId != 0 {
		success = executeWebhook(ctx.Shard, webhook, rest.WebhookBody{
			Content:         ctx.Message.Content,
			Username:        ctx.Message.Author.Username,
			AvatarUrl:       ctx.Author.AvatarUrl(256),
		})
	}

	if !success {
		if _, err := ctx.Shard.CreateMessage(session.StaffChannelId, ctx.Message.Content); err != nil {
			utils.SendEmbed(ctx.Shard, dmChannel, utils.Red, "Error", fmt.Sprintf("An error has occurred: `%s`", err.Error()), nil, 30, ctx.IsPremium)
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	// forward attachments
	// don't re-upload attachments incase user has uploaded TOS breaking attachment
	if len(ctx.Message.Attachments) > 0 {
		var content string
		if len(ctx.Message.Attachments) == 1 {
			content = fmt.Sprintf("%s attached a file:", ctx.Author.Mention())
		} else {
			content = fmt.Sprintf("%s attached files:", ctx.Author.Mention())
		}

		for _, attachment := range ctx.Message.Attachments {
			content += fmt.Sprintf("\n▶️ %s", attachment.ProxyUrl)
		}

		if _, err := ctx.Shard.CreateMessage(session.StaffChannelId, content); err != nil {
			utils.SendEmbed(ctx.Shard, dmChannel, utils.Red, "Error", fmt.Sprintf("An error has occurred: `%s`", err.Error()), nil, 30, ctx.IsPremium)
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}
}

func executeWebhook(shard *gateway.Shard, webhook database.ModmailWebhook, data rest.WebhookBody) bool {
	_, err := shard.ExecuteWebhook(webhook.WebhookId, webhook.WebhookToken, true, data)

	if err == request.ErrForbidden || err == request.ErrNotFound {
		go dbclient.Client.ModmailWebhook.Delete(webhook.Uuid)
		return false
	} else {
		return true
	}
}

// TODO: Make this less hacky
func handleCommand(ctx utils.CommandContext, session database.ModmailSession) (utils.CommandContext, bool) {
	customPrefix, err := dbclient.Client.Prefix.Get(session.GuildId); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(ctx.Message.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if customPrefix != "" && strings.HasPrefix(ctx.Message.Content, customPrefix) {
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
