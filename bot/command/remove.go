package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"strconv"
	"strings"
)

type RemoveCommand struct {
}

func (RemoveCommand) Name() string {
	return "remove"
}

func (RemoveCommand) Description() string {
	return "Removes a user from a ticket"
}

func (RemoveCommand) Aliases() []string {
	return []string{}
}

func (RemoveCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (RemoveCommand) Execute(ctx CommandContext) {
	// Check users are mentioned
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention members to remove from the ticket")
		ctx.ReactWithCross()
		return
	}

	// Check channel is mentioned
	found := utils.ChannelMentionRegex.FindStringSubmatch(strings.Join(ctx.Args, " "))
	if len(found) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a ticket channel to remove the user(s) from")
		ctx.ReactWithCross()
		return
	}

	// Verify that the specified channel is a real ticket
	ticket, err := strconv.ParseInt(found[1], 10, 64); if err != nil {
		ctx.SendEmbed(utils.Red, "Error", "The specified channel is not a ticket")
		ctx.ReactWithCross()
		return
	}

	isTicketChan := make(chan bool)
	go database.IsTicketChannel(ticket, isTicketChan)
	isTicket := <- isTicketChan

	if !isTicket {
		ctx.SendEmbed(utils.Red, "Error", "The specified channel is not a ticket")
		ctx.ReactWithCross()
		return
	}

	// Get ticket ID
	ticketIdChan := make(chan int)
	go database.GetTicketId(ticket, ticketIdChan)
	ticketId := <- ticketIdChan

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Remove user from ticket in DB
		go database.RemoveMember(ticketId, guildId, user.ID)

		// Remove user from ticket
		err = ctx.Session.ChannelPermissionSet(
			strconv.Itoa(int(ticket)),
			user.ID,
			"member",
			0,
			utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks))

		if err != nil {
			log.Error(err.Error())
		}
	}

	ctx.ReactWithCheck()
}

func (RemoveCommand) Parent() interface{} {
	return nil
}

func (RemoveCommand) Children() []Command {
	return make([]Command, 0)
}

func (RemoveCommand) PremiumOnly() bool {
	return false
}

func (RemoveCommand) AdminOnly() bool {
	return false
}

func (RemoveCommand) HelperOnly() bool {
	return false
}
