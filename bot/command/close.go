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
	archiveChannel, err := ctx.Session.State.Channel(archiveChannelId); if err != nil {
		// Not cached
		archiveChannel, err = ctx.Session.Channel(archiveChannelId); if err != nil {
			// Channel doesn't exist
			channelExists = false
		}
	}

	if channelExists {
		requiredPerms := []utils.Permission{
			utils.ViewChannel,
			utils.SendMessages,
			utils.AttachFiles,
		}

		hasPermissions := true
		for _, perm := range requiredPerms {
			hasPermChannel := make(chan bool)
			go utils.ChannelMemberHasPermission(ctx.Session, ctx.Guild, archiveChannelId, utils.Id, perm, hasPermChannel)
			if !<-hasPermChannel {
				hasPermissions = false
				break
			}
		}

		if hasPermissions {
			data := discordgo.MessageSend{
				Content: fmt.Sprintf("Archive of `#%s` (closed by %s%d) with message `%s`", ),
			}

			if _, err := ctx.Session.ChannelMessageSendComplex(archiveChannelId, &data); err != nil {
				log.Error(err.Error())
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

func (CloseCommand) AdminOnly() bool {
	return false
}

func (CloseCommand) HelperOnly() bool {
	return false
}
