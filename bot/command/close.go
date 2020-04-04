package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type CloseCommand struct {
}

func (CloseCommand) Name() string {
	return "close"
}

func (CloseCommand) Description() string {
	return "Closes the current ticket"
}

func (CloseCommand) Aliases() []string {
	return []string{}
}

func (CloseCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (CloseCommand) Execute(ctx utils.CommandContext) {
	// Verify this is a ticket or modmail channel
	isTicketChan := make(chan bool)
	go database.IsTicketChannel(ctx.ChannelId, isTicketChan)
	isTicket := <-isTicketChan

	if !isTicket {
		// Delegate to modmail executor if this is a modmail channel
		// TODO: Improve command executor to differentiate between modmail and ticket channels
		modmailSessionChan := make(chan *modmaildatabase.ModMailSession)
		go modmaildatabase.GetModMailSessionByStaffChannel(ctx.ChannelId, modmailSessionChan)
		modmailSession := <-modmailSessionChan

		if modmailSession != nil {
			modmail.HandleClose(modmailSession, ctx)
		} else {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", "This is not a ticket channel")
		}

		return
	}

	// Create reason
	var reason string
	silentClose := false
	for _, arg := range ctx.Args {
		if arg == "--silent" {
			silentClose = true
		} else {
			reason += fmt.Sprintf("%s ", arg)
		}
	}
	reason = strings.TrimSuffix(reason, " ")

	// Check the user is permitted to close the ticket
	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Shard, ctx.Member, ctx.Guild.Id, permissionLevelChan)
	permissionLevel := <-permissionLevelChan

	// Get ticket struct
	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(ctx.ChannelId, ticketChan)
	ticket := <-ticketChan

	usersCanCloseChan := make(chan bool)
	go database.IsUserCanClose(ctx.Guild.Id, usersCanCloseChan)
	usersCanClose := <-usersCanCloseChan

	if (permissionLevel == utils.Everyone && ticket.Owner != ctx.User.Id) || (permissionLevel == utils.Everyone && !usersCanClose) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You are not permitted to close this ticket")
		return
	}

	if !permission.HasPermissions(ctx.Shard, ctx.Guild.Id, ctx.Shard.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}

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

	// Set ticket state as closed and delete channel
	go database.Close(ctx.Guild.Id, ticket.Id)
	if _, err := ctx.Shard.DeleteChannel(ctx.ChannelId); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Send logs to archive channel
	archiveChannelChan := make(chan uint64)
	go database.GetArchiveChannel(ctx.Guild.Id, archiveChannelChan)
	archiveChannelId := <-archiveChannelChan

	channelExists := true
	if _, err := ctx.Shard.GetChannel(archiveChannelId); err != nil {
		channelExists = false
	}

	// Save space - delete the webhook
	go database.DeleteWebhookByUuid(ticket.Uuid)

	if channelExists {
		msg := fmt.Sprintf("Archive of `#ticket-%d` (closed by %s#%s)", ticket.Id, ctx.User.Username, utils.PadDiscriminator(ctx.User.Discriminator))
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
		m, err := ctx.Shard.CreateMessageComplex(archiveChannelId, data)
		if err == nil {
			// Add archive to DB
			uuidChan := make(chan string)
			go database.GetTicketUuid(ctx.ChannelId, uuidChan)
			uuid := <-uuidChan

			userNameChan := make(chan string)
			go database.GetUsername(ticket.Owner, userNameChan)
			userName := <-userNameChan

			go database.AddArchive(uuid, ctx.Guild.Id, ticket.Owner, userName, ticket.Id, m.Attachments[0].Url)
		} else {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		// Notify user and send logs in DMs
		if !silentClose {
			var content string
			// Create message content
			if ctx.User.Id == ticket.Owner {
				content = fmt.Sprintf("You closed your ticket (`#ticket-%d`) in `%s`", ticket.Id, ctx.Guild.Name)
			} else if len(ctx.Args) == 0 {
				content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s", ticket.Id, ctx.Guild.Name, ctx.User.Mention())
			} else {
				content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s with reason `%s`", ticket.Id, ctx.Guild.Name, ctx.User.Mention(), reason)
			}

			privateMessage, err := ctx.Shard.CreateDM(ticket.Owner)
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
				_, _ = ctx.Shard.CreateMessageComplex(privateMessage.Id, data)
			}
		}
	}
}

func (CloseCommand) Parent() interface{} {
	return nil
}

func (CloseCommand) Children() []Command {
	return make([]Command, 0)
}

func (CloseCommand) PremiumOnly() bool {
	return false
}

func (CloseCommand) Category() Category {
	return Tickets
}

func (CloseCommand) AdminOnly() bool {
	return false
}

func (CloseCommand) HelperOnly() bool {
	return false
}
