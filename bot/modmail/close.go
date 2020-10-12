package modmail

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/archive"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	dbclient "github.com/timnboys/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/timnboys/database"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"strings"
	"time"
)

func HandleClose(session database.ModmailSession, ctx utils.CommandContext) {
	reason := strings.Join(ctx.Args, " ")

	// Check the user is permitted to close the ticket
	usersCanClose, err :=  dbclient.Client.UsersCanClose.Get(session.GuildId)
	if err != nil {
		sentry.Error(err)
	}

	if (ctx.UserPermissionLevel == utils.Everyone && session.UserId != ctx.Author.Id) || (ctx.UserPermissionLevel == utils.Everyone && !usersCanClose) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You are not permitted to close this ticket")
		return
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}*/

	if ctx.ShouldReact {
		ctx.ReactWithCheck()
	}

	// Archive
	msgs := make([]message.Message, 0)

	lastId := uint64(0)
	count := -1
	for count != 0 {
		array, err := ctx.Shard.GetChannelMessages(ctx.ChannelId, rest.GetChannelMessagesData{
			Before: lastId,
			Limit:  100,
		})

		count = len(array)
		if err != nil {
			count = 0
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		if count > 0 {
			lastId = array[len(array)-1].Id

			for _, msg := range array {
				msgs = append(msgs, msg)
				if msg.Id == session.WelcomeMessageId {
					count = 0
					break
				}
			}
		}
	}

	// Reverse messages
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	logs := ""
	for _, msg := range msgs {
		date := msg.Timestamp.UTC().String()

		content := msg.Content

		// append attachments
		for _, attachment := range msg.Attachments {
			content += fmt.Sprintf(" %s", attachment.ProxyUrl)
		}

		logs += fmt.Sprintf("[%s][%d] %s: %s\n", date, msg.Id, msg.Author.Username, content)
	}

	// we don't use this yet so chuck it in a goroutine
	go func() {
		isPremium := make(chan bool)
		go utils.IsPremiumGuild(ctx.Shard, ctx.GuildId, isPremium)

		if err := archive.ArchiverClient.StoreModmail(msgs, session.GuildId, session.Uuid.String(), <-isPremium); err != nil {
			sentry.Error(err)
		}
	}()

	// Delete the webhook
	// We need to block for this
	if err := dbclient.Client.ModmailWebhook.Delete(session.Uuid); err != nil {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Set ticket state as closed and delete channel
	go dbclient.Client.ModmailSession.DeleteByUser(utils.BOT_ID, session.UserId)
	if _, err := ctx.Shard.DeleteChannel(session.StaffChannelId); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Send logs to archive channel
	archiveChannelId, err := dbclient.Client.ArchiveChannel.Get(session.GuildId); if err != nil {
		sentry.Error(err)
	}

	var channelExists bool
	if archiveChannelId != 0 {
		if _, err := ctx.Shard.GetChannel(archiveChannelId); err == nil {
			channelExists = true
		}
	}

	if channelExists {
		embed := embed.NewEmbed().
			SetTitle("Ticket Closed").
			SetColor(int(utils.Green)).
			AddField("Closed By", ctx.Author.Mention(), true).
			AddField("Archive", fmt.Sprintf("[Click here](https://panel.ticketsbot.net/manage/%d/logs/modmail/view/%s)", session.GuildId, session.Uuid), true)

		if reason == "" {
			embed.AddField("Reason", "No reason specified", false)
		} else {
			embed.AddField("Reason", reason, false)
		}

		if _, err := ctx.Shard.CreateMessageEmbed(archiveChannelId, embed); err != nil {
			sentry.Error(err)
		}
	}

	go dbclient.Client.ModmailArchive.Set(database.ModmailArchive{
		Uuid:      session.Uuid,
		GuildId:   session.GuildId,
		UserId:    session.UserId,
		CloseTime: time.Now(),
	})

	guild, err := ctx.Shard.GetGuild(session.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Notify user and send logs in DMs
	privateMessage, err := ctx.Shard.CreateDM(session.UserId)
	if err == nil {
		var content string
		// Create message content
		if ctx.Author.Id == session.UserId {
			content = fmt.Sprintf("You closed your modmail ticket in `%s`", guild.Name)
		} else if len(ctx.Args) == 0 {
			content = fmt.Sprintf("Your modmail ticket in `%s` was closed by %s", guild.Name, ctx.Author.Mention())
		} else {
			content = fmt.Sprintf("Your modmail ticket in `%s` was closed by %s with reason `%s`", guild.Name, ctx.Author.Mention(), reason)
		}

		// Errors occur when users have privacy settings high
		_, _ = ctx.Shard.CreateMessage(privateMessage.Id, content)
	}
}
