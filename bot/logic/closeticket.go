package logic

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"strings"
)

// TODO: MODMAIL CLOSE REACT
func CloseTicket(s *gateway.Shard, guildId, channelId, messageId uint64, member member.Member, args []string, fromReaction, isPremium bool) {
	reference := message.MessageReference{
		MessageId: messageId,
		ChannelId: channelId,
		GuildId:   guildId,
	}

	// Verify this is a ticket or modmail channel
	isTicketChan := make(chan bool)
	go database.IsTicketChannel(channelId, isTicketChan)
	isTicket := <-isTicketChan

	// Cannot happen if fromReaction
	if !isTicket {
		utils.ReactWithCross(s, reference)
		utils.SendEmbed(s, channelId, utils.Red, "Error", "This is not a ticket channel", 30, isPremium)

		return
	}

	// Create reason
	var reason string
	silentClose := false
	for _, arg := range args {
		if arg == "--silent" {
			silentClose = true
		} else {
			reason += fmt.Sprintf("%s ", arg)
		}
	}
	reason = strings.TrimSuffix(reason, " ")

	// Check the user is permitted to close the ticket
	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(s, member, guildId, permissionLevelChan)
	permissionLevel := <-permissionLevelChan

	// Get ticket struct
	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(channelId, ticketChan)
	ticket := <-ticketChan

	usersCanCloseChan := make(chan bool)
	go database.IsUserCanClose(guildId, usersCanCloseChan)
	usersCanClose := <-usersCanCloseChan

	if (permissionLevel == utils.Everyone && ticket.Owner != member.User.Id) || (permissionLevel == utils.Everyone && !usersCanClose) {
		if !fromReaction {
			utils.ReactWithCross(s, reference)
			utils.SendEmbed(s, channelId, utils.Red, "Error", "You are not permitted to close this ticket", 30, isPremium)
		}
		return
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(s, guildId, s.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}*/

	if !fromReaction {
		utils.ReactWithCheck(s, reference)
	}

	// Archive
	msgs := make([]message.Message, 0)

	lastId := uint64(0)
	count := -1
	for count != 0 {
		array, err := s.GetChannelMessages(channelId, rest.GetChannelMessagesData{
			Before: lastId,
			Limit:  100,
		})

		count = len(array)
		if err != nil {
			count = 0
			sentry.Error(err)
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

	// Set ticket state as closed and delete channel
	go database.Close(guildId, ticket.Id)
	if _, err := s.DeleteChannel(channelId); err != nil {
		sentry.Error(err)
	}

	// Send logs to archive channel
	archiveChannelChan := make(chan uint64)
	go database.GetArchiveChannel(guildId, archiveChannelChan)
	archiveChannelId := <-archiveChannelChan

	channelExists := true
	if _, err := s.GetChannel(archiveChannelId); err != nil {
		channelExists = false
	}

	// Save space - delete the webhook
	go database.DeleteWebhookByUuid(ticket.Uuid)

	if channelExists {
		msg := fmt.Sprintf("Archive of `#ticket-%d` (closed by %s#%s)", ticket.Id, member.User.Username, utils.PadDiscriminator(member.User.Discriminator))
		if reason != "" {
			msg += fmt.Sprintf(" with reason `%s`", reason)
		}

		data := rest.CreateMessageData{
			Content: msg,
			File: &rest.File{
				Name:        fmt.Sprintf("ticket-%d.txt", ticket.Id),
				ContentType: "text/plain",
				Reader:      strings.NewReader(logs),
			},
		}

		// Errors occur when the bot doesn't have permission
		m, err := s.CreateMessageComplex(archiveChannelId, data)
		if err == nil {
			// Add archive to DB
			uuidChan := make(chan string)
			go database.GetTicketUuid(channelId, uuidChan)
			uuid := <-uuidChan

			userNameChan := make(chan string)
			go database.GetUsername(ticket.Owner, userNameChan)
			userName := <-userNameChan

			go database.AddArchive(uuid, guildId, ticket.Owner, userName, ticket.Id, m.Attachments[0].Url)
		} else {
			sentry.Error(err)
		}

		// Notify user and send logs in DMs
		if !silentClose {
			// get guild object
			guild, err := s.GetGuild(guildId)
			if err != nil {
				sentry.Error(err)
				return
			}

			var content string
			// Create message content
			if member.User.Id == ticket.Owner {
				content = fmt.Sprintf("You closed your ticket (`#ticket-%d`) in `%s`", ticket.Id, guild.Name)
			} else if len(args) == 0 {
				content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s", ticket.Id, guild.Name, member.User.Mention())
			} else {
				content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s with reason `%s`", ticket.Id, guild.Name, member.User.Mention(), reason)
			}

			privateMessage, err := s.CreateDM(ticket.Owner)
			// Only send the msg if we could create the channel
			if err == nil {
				data := rest.CreateMessageData{
					Content: content,
					File: &rest.File{
						Name:        fmt.Sprintf("ticket-%d.txt", ticket.Id),
						ContentType: "text/plain",
						Reader:      strings.NewReader(logs),
					},
				}

				// Errors occur when users have privacy settings high
				_, _ = s.CreateMessageComplex(privateMessage.Id, data)
			}
		}
	}
}
