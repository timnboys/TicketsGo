package modmail

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

func HandleClose(session *modmaildatabase.ModMailSession, ctx utils.CommandContext) {
	reason := strings.Join(ctx.Args, " ")

	// Check the user is permitted to close the ticket
	permissionLevel := utils.Everyone
	if ctx.Member != nil {
		permissionLevelChan := make(chan utils.PermissionLevel)
		go utils.GetPermissionLevel(ctx.Shard, ctx.Member, permissionLevelChan)
		permissionLevel = <-permissionLevelChan
	}

	usersCanCloseChan := make(chan bool)
	go database.IsUserCanClose(session.Guild, usersCanCloseChan)
	usersCanClose := <-usersCanCloseChan

	if (permissionLevel == utils.Everyone && session.User != ctx.UserID) || (permissionLevel == utils.Everyone && !usersCanClose) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You are not permitted to close this ticket")
		return
	}

	hasPerm := make(chan bool)
	go utils.MemberHasPermission(ctx.Shard, ctx.Guild.ID, utils.Id, utils.ManageChannels, hasPerm)

	if !<-hasPerm {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}

	if ctx.ShouldReact {
		ctx.ReactWithCheck()
	}

	// Archive
	msgs := make([]*discordgo.Message, 0)

	lastId := ""
	count := -1
	for count != 0 {
		array, err := ctx.Shard.ChannelMessages(ctx.Channel, 100, lastId, "", "")

		count = len(array)
		if err != nil {
			count = 0
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		if count > 0 {
			lastId = array[len(array)-1].ID

			msgs = append(msgs, array...)
		}
	}

	// Reverse messages
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	logs := ""
	for _, msg := range msgs {
		var date string
		if t, err := msg.Timestamp.Parse(); err == nil {
			date = t.UTC().String()
		}

		content := msg.Content

		// append attachments
		for _, attachment := range msg.Attachments {
			content += fmt.Sprintf(" %s", attachment.ProxyURL)
		}

		logs += fmt.Sprintf("[%s][%s] %s: %s\n", date, msg.ID, msg.Author.Username, content)
	}

	// Get channel name
	var channelName string
	channel, err := ctx.Shard.Channel(strconv.Itoa(int(session.StaffChannel)))
	if err == nil {
		channelName = channel.Name
	} else {
		channelName = "invalid-channel"
	}

	// Set ticket state as closed and delete channel
	go modmaildatabase.CloseModMailSessions(session.User)
	if _, err := ctx.Shard.ChannelDelete(strconv.Itoa(int(session.StaffChannel))); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Send logs to archive channel
	archiveChannelChan := make(chan int64)
	go database.GetArchiveChannel(session.Guild, archiveChannelChan)
	archiveChannelId := strconv.Itoa(int(<-archiveChannelChan))

	channelExists := true
	_, err = ctx.Shard.State.Channel(archiveChannelId); if err != nil {
		// Not cached
		_, err = ctx.Shard.Channel(archiveChannelId); if err != nil {
			// Channel doesn't exist
			channelExists = false
		}
	}

	// Save space - delete the webhook
	go database.DeleteWebhookByUuid(session.Uuid)

	if channelExists {
		msg := fmt.Sprintf("Archive of `#%s` (closed by %s#%s)", channelName, ctx.User.Username, ctx.User.Discriminator)
		if reason != "" {
			msg += fmt.Sprintf(" with reason `%s`", reason)
		}

		data := discordgo.MessageSend{
			Content: msg,
			Files: []*discordgo.File{
				{
					Name: fmt.Sprintf("%s.txt", channelName),
					ContentType: "text/plain",
					Reader: strings.NewReader(logs),
				},
			},
		}

		// Errors occur when the bot doesn't have permission
		m, err := ctx.Shard.ChannelMessageSendComplex(archiveChannelId, &data)
		if err == nil {
			userNameChan := make(chan string)
			go database.GetUsername(session.User, userNameChan)
			userName := <-userNameChan

			archive := modmaildatabase.ModMailArchive{
				Uuid:     session.Uuid,
				Guild:    session.Guild,
				User:     session.User,
				Username: userName,
				CdnUrl:   m.Attachments[0].URL,
			}
			go archive.Store()
		} else {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}
	}

	// Notify user and send logs in DMs
	privateMessage, err := ctx.Shard.UserChannelCreate(strconv.Itoa(int(session.User)))
	if err == nil {
		var content string
		// Create message content
		if ctx.UserID == session.User {
			content = fmt.Sprintf("You closed your modmail session in `%s`", ctx.Guild.Name)
		} else if len(ctx.Args) == 0 {
			content = fmt.Sprintf("Your modmail session in `%s` was closed by %s", ctx.Guild.Name, ctx.User.Mention())
		} else {
			content = fmt.Sprintf("Your modmail session in `%s` was closed by %s with reason `%s`", ctx.Guild.Name, ctx.User.Mention(), reason)
		}

		// Errors occur when users have privacy settings high
		_, _ = ctx.Shard.ChannelMessageSend(privateMessage.ID, content)
	}
}
