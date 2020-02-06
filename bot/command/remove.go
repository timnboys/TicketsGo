package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
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

func (RemoveCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention members to remove from the ticket")
		ctx.ReactWithCross()
		return
	}

	// Verify that the current channel is a real ticket
	channelId, err := strconv.ParseInt(ctx.Channel, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	isTicketChan := make(chan bool)
	go database.IsTicketChannel(channelId, isTicketChan)
	isTicket := <- isTicketChan

	if !isTicket {
		ctx.SendEmbed(utils.Red, "Error", "The current channel is not a ticket")
		ctx.ReactWithCross()
		return
	}

	// Get ticket ID
	ticketIdChan := make(chan int)
	go database.GetTicketId(channelId, ticketIdChan)
	ticketId := <- ticketIdChan

	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Verify that the user is allowed to modify the ticket
	permLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Session, ctx.Member, permLevelChan)
	permLevel := <-permLevelChan

	ownerChan := make(chan int64)
	go database.GetOwner(ticketId, guildId, ownerChan)
	owner := <-ownerChan

	if permLevel == 0 && strconv.Itoa(int(owner)) != ctx.User.ID {
		ctx.SendEmbed(utils.Red, "Error", "You don't have permission to add people to this ticket")
		ctx.ReactWithCross()
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Remove user from ticket in DB
		go database.RemoveMember(ticketId, guildId, user.ID)

		// Remove user from ticket
		err = ctx.Session.ChannelPermissionSet(
			strconv.Itoa(int(channelId)),
			user.ID,
			"member",
			0,
			utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks))

		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
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
