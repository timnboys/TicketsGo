package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
	"strings"
)

type AddCommand struct {
}

func (AddCommand) Name() string {
	return "add"
}

func (AddCommand) Description() string {
	return "Adds a user to a ticket"
}

func (AddCommand) Aliases() []string {
	return []string{}
}

func (AddCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AddCommand) Execute(ctx utils.CommandContext) {
	// Check users are mentioned
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention members to add to the ticket")
		ctx.ReactWithCross()
		return
	}

	// Check channel is mentioned
	found := utils.ChannelMentionRegex.FindStringSubmatch(strings.Join(ctx.Args, " "))
	if len(found) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a ticket channel to add the user(s) in")
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
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Verify that the user is allowed to modify the ticket
	permLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Session, ctx.Guild, ctx.User.ID, permLevelChan)
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
		// Add user to ticket in DB
		go database.AddMember(ticketId, guildId, ctx.User.ID)

		// Add user to ticket
		err = ctx.Session.ChannelPermissionSet(
			strconv.Itoa(int(ticket)),
			user.ID,
			"member",
			utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
			0)

		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}

	ctx.ReactWithCheck()
}

func (AddCommand) Parent() interface{} {
	return nil
}

func (AddCommand) Children() []Command {
	return make([]Command, 0)
}

func (AddCommand) PremiumOnly() bool {
	return false
}

func (AddCommand) AdminOnly() bool {
	return false
}

func (AddCommand) HelperOnly() bool {
	return false
}
