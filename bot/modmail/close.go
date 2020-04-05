package modmail

import (
	"fmt"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/message"
	//"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"strings"
)

func HandleClose(shard *modmaildatabase.ModMailSession, ctx utils.CommandContext) {
	reason := strings.Join(ctx.Args, " ")

	// Check the user is permitted to close the ticket
	permissionLevel := utils.Everyone
	if ctx.Member != nil {
		permissionLevelChan := make(chan utils.PermissionLevel)
		go utils.GetPermissionLevel(ctx.Shard, ctx.Member, ctx.Guild.Id, permissionLevelChan)
		permissionLevel = <-permissionLevelChan
	}

	usersCanCloseChan := make(chan bool)
	go database.IsUserCanClose(shard.Guild, usersCanCloseChan)
	usersCanClose := <-usersCanCloseChan

	if (permissionLevel == utils.Everyone && shard.User != ctx.User.Id) || (permissionLevel == utils.Everyone && !usersCanClose) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You are not permitted to close this ticket")
		return
	}

	// TODO: Re-add perm checks
	/*if !permission.HasPermissions(ctx.Shard, ctx.Guild.Id, ctx.Shard.SelfId(), permission.ManageChannels) {
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

			msgs = append(msgs, array...)
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

	// Get channel name
	var channelName string
	channel, err := ctx.Shard.GetChannel(shard.StaffChannel)
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = "invalid-channel"
	}

	// Set ticket state as closed and delete channel
	go modmaildatabase.CloseModMailSessions(shard.User)
	if _, err := ctx.Shard.DeleteChannel(shard.StaffChannel); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Send logs to archive channel
	archiveChannelChan := make(chan uint64)
	go database.GetArchiveChannel(shard.Guild, archiveChannelChan)
	archiveChannelId := <-archiveChannelChan

	channelExists := true
	if _, err := ctx.Shard.GetChannel(archiveChannelId); err != nil {
		channelExists = false
	}

	// Save space - delete the webhook
	go database.DeleteWebhookByUuid(shard.Uuid)

	if channelExists {
		msg := fmt.Sprintf("Archive of `#%s` (closed by %s#%d)", channelName, ctx.User.Username, ctx.User.Discriminator)
		if reason != "" {
			msg += fmt.Sprintf(" with reason `%s`", reason)
		}

		data := rest.CreateMessageData{
			Content: msg,
			File: &rest.File{
				Name:        fmt.Sprintf("%s.txt", channelName),
				ContentType: "text/plain",
				Reader:      strings.NewReader(logs),
			},
		}

		// Errors occur when the bot doesn't have permission
		m, err := ctx.Shard.CreateMessageComplex(archiveChannelId, data)
		if err == nil {
			userNameChan := make(chan string)
			go database.GetUsername(shard.User, userNameChan)
			userName := <-userNameChan

			archive := modmaildatabase.ModMailArchive{
				Uuid:     shard.Uuid,
				Guild:    shard.Guild,
				User:     shard.User,
				Username: userName,
				CdnUrl:   m.Attachments[0].Url,
			}
			go archive.Store()
		} else {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	// Notify user and send logs in DMs
	privateMessage, err := ctx.Shard.CreateDM(shard.User)
	if err == nil {
		var content string
		// Create message content
		if ctx.User.Id == shard.User {
			content = fmt.Sprintf("You closed your modmail shard in `%s`", ctx.Guild.Name)
		} else if len(ctx.Args) == 0 {
			content = fmt.Sprintf("Your modmail shard in `%s` was closed by %s", ctx.Guild.Name, ctx.User.Mention())
		} else {
			content = fmt.Sprintf("Your modmail shard in `%s` was closed by %s with reason `%s`", ctx.Guild.Name, ctx.User.Mention(), reason)
		}

		// Errors occur when users have privacy settings high
		_, _ = ctx.Shard.CreateMessage(privateMessage.Id, content)
	}
}
