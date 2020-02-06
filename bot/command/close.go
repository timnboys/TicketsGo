package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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
	channelId, err := strconv.ParseInt(ctx.Channel, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	isTicket := make(chan bool)
	go database.IsTicketChannel(channelId, isTicket)

	if !<-isTicket {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "This is not a ticket channel")
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
	go utils.GetPermissionLevel(ctx.Session, ctx.Member, permissionLevelChan)
	permissionLevel := <-permissionLevelChan

	idChan := make(chan int)
	go database.GetTicketId(channelId, idChan)
	id := <-idChan

	ownerChan := make(chan int64)
	go database.GetOwner(id, guildId, ownerChan)
	owner := <-ownerChan

	usersCanCloseChan := make(chan bool)
	go database.IsUserCanClose(guildId, usersCanCloseChan)
	usersCanClose := <-usersCanCloseChan

	if (permissionLevel == 0 && strconv.Itoa(int(owner)) != ctx.User.ID) || (permissionLevel == 0 && !usersCanClose) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You are not permitted to close this ticket")
		return
	}

	hasPerm := make(chan bool)
	go utils.MemberHasPermission(ctx.Session, ctx.Guild.ID, utils.Id, utils.ManageChannels, hasPerm)

	if !<-hasPerm {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}

	ctx.ReactWithCheck()

	// Archive
	msgs := make([]*discordgo.Message, 0)

	lastId := ""
	count := -1
	for count != 0 {
		array, err := ctx.Session.ChannelMessages(ctx.Channel, 100, lastId, "", "")

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

		logs += fmt.Sprintf("[%s][%s] %s: %s\n", date, msg.ID, msg.Author.Username, msg.Content)
	}

	archiveChannelChan := make(chan int64)
	go database.GetArchiveChannel(guildId, archiveChannelChan)
	archiveChannelId := strconv.Itoa(int(<-archiveChannelChan))

	channelExists := true
	_, err = ctx.Session.State.Channel(archiveChannelId); if err != nil {
		// Not cached
		_, err = ctx.Session.Channel(archiveChannelId); if err != nil {
			// Channel doesn't exist
			channelExists = false
		}
	}

	if channelExists {
		msg := fmt.Sprintf("Archive of `#ticket-%d` (closed by %s#%s)", id, ctx.User.Username, ctx.User.Discriminator)
		if reason != "" {
			msg += fmt.Sprintf(" with reason `%s`", reason)
		}

		data := discordgo.MessageSend{
			Content: msg,
			Files: []*discordgo.File{
				{
					Name: fmt.Sprintf("ticket-%d.txt", id),
					ContentType: "text/plain",
					Reader: strings.NewReader(logs),
				},
			},
		}

		userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		// Errors occur when the bot doesn't have permission
		m, err := ctx.Session.ChannelMessageSendComplex(archiveChannelId, &data)
		if err == nil {
			// Add archive to DB
			uuidChan := make(chan string)
			go database.GetTicketUuid(channelId, uuidChan)
			uuid := <-uuidChan

			userNameChan := make(chan string)
			go database.GetUsername(owner, userNameChan)
			userName := <-userNameChan

			go database.AddArchive(uuid, guildId, owner, userName, id, m.Attachments[0].URL)
		} else {
			sentry.LogWithContext(err, ctx.ToErrorContext())
		}

		// Notify user and send logs in DMs
		if !silentClose {
			var content string
			// Create message content
			if userId == owner {
				content = fmt.Sprintf("You closed your ticket (`#ticket-%d`) in `%s`", id, ctx.Guild.Name)
			} else if len(ctx.Args) == 0 {
				content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s", id, ctx.Guild.Name, ctx.User.Mention())
			} else {
				content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s with reason `%s`", id, ctx.Guild.Name, ctx.User.Mention(), reason)
			}

			privateMessage, err := ctx.Session.UserChannelCreate(strconv.Itoa(int(owner)))
			// Only send the msg if we could create the channel
			if err == nil {
				data := discordgo.MessageSend{
					Content: content,
					Files: []*discordgo.File{
						{
							Name:        fmt.Sprintf("ticket-%d.txt", id),
							ContentType: "text/plain",
							Reader:      strings.NewReader(logs),
						},
					},
				}

				// Errors occur when users have privacy settings high
				_, _ = ctx.Session.ChannelMessageSendComplex(privateMessage.ID, &data)
			}
		}

		// Set ticket state as closed and delete channel
		go database.Close(guildId, id)
		if _, err = ctx.Session.ChannelDelete(ctx.Channel); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
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

func (CloseCommand) AdminOnly() bool {
	return false
}

func (CloseCommand) HelperOnly() bool {
	return false
}
