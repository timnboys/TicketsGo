package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
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

func (CloseCommand) Execute(ctx CommandContext) {
	channelId, err := strconv.ParseInt(ctx.Channel, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
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
	//silentClose := false
	for _, arg := range ctx.Args {
		if arg == "--silent" {
		//	silentClose = true
		} else {
			reason += fmt.Sprintf("%s ", arg)
		}
	}
	reason = strings.TrimSuffix(reason, " ")

	// Check the user is permitted to close the ticket
	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Session, ctx.Guild, ctx.User.ID, permissionLevelChan)
	permissionlevel := <-permissionLevelChan

	idChan := make(chan int)
	go database.GetTicketId(channelId, idChan)
	id := <-idChan

	ownerChan := make(chan int64)
	go database.GetOwner(id, guildId, ownerChan)
	owner := <-ownerChan

	if permissionlevel == 0 && strconv.Itoa(int(owner)) != ctx.User.ID {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You are not permitted to close this ticket")
		return
	}

	hasPerm := make(chan bool)
	go utils.MemberHasPermission(ctx.Session, ctx.Guild, utils.Id, utils.ManageChannels, hasPerm)

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

	content := ""
	for _, msg := range msgs {
		var date string
		if t, err := msg.Timestamp.Parse(); err == nil {
			date = t.UTC().String()
		}

		content += fmt.Sprintf("[%s][%s] %s: %s\n", date, msg.ID, msg.Author.Username, msg.Content)
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
					Reader: strings.NewReader(content),
				},
			},
		}

		// Errors occur when the bot doesn't have permission
		m, err := ctx.Session.ChannelMessageSendComplex(archiveChannelId, &data)

		// Add archive to DB
		userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
			log.Error(err.Error())
			return
		}

		if err == nil {
			uuidChan := make(chan string)
			go database.GetTicketUuid(channelId, uuidChan)
			uuid := <-uuidChan

			go database.AddArchive(uuid, guildId, userId, ctx.User.Username, id, m.Attachments[0].URL)
		}

		// Notify user and send logs in DMs
		guild, err := ctx.Session.State.Guild(ctx.Guild); if err != nil {
			// Not cached
			guild, err = ctx.Session.Guild(ctx.Guild); if err != nil {
				log.Error(err.Error())
				return
			}
		}

		// Create message content
		if userId == owner {
			content = fmt.Sprintf("You closed your ticket (`#ticket-%d`) in `%s`", id, guild.Name)
		} else if len(ctx.Args) == 0 {
			content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s", id, guild.Name, ctx.User.Mention())
		} else {
			content = fmt.Sprintf("Your ticket (`#ticket-%d`) in `%s` was closed by %s with reason `%s`", id, guild.Name, ctx.User.Mention(), reason)
		}

		privateMessage, err := ctx.Session.UserChannelCreate(strconv.Itoa(int(owner)))
		// Only send the msg if we could create the channel
		if err == nil {
			data := discordgo.MessageSend{
				Content: content,
				Files: []*discordgo.File{
					{
						Name: fmt.Sprintf("ticket-%d.txt", id),
						ContentType: "text/plain",
						Reader: strings.NewReader(content),
					},
				},
			}

			// Errors occur when users have privacy settings high
			_, _ = ctx.Session.ChannelMessageSendComplex(privateMessage.ID, &data)
		}

		// Set ticket state as closed and delete chan
		go database.Close(guildId, id)
		if _, err = ctx.Session.ChannelDelete(ctx.Channel); err != nil {
			log.Error(err.Error())
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
