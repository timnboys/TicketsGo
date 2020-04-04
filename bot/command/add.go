package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
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
		ctx.SendEmbed(utils.Red, "Error", "You need to mention members to add to the ticketChannel")
		ctx.ReactWithCross()
		return
	}

	// Check channel is mentioned
	if len(ctx.Message.MentionChannels) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a ticketChannel channel to add the user(s) in")
		ctx.ReactWithCross()
		return
	}

	// Verify that the specified channel is a real ticketChannel
	ticketChannel := ctx.Message.MentionChannels[0]

	isTicketChan := make(chan bool)
	go database.IsTicketChannel(ticketChannel.Id, isTicketChan)
	isTicket := <- isTicketChan

	if !isTicket {
		ctx.SendEmbed(utils.Red, "Error", "The specified channel is not a ticketChannel")
		ctx.ReactWithCross()
		return
	}

	// Get ticketChannel ID
	ticketIdChan := make(chan int)
	go database.GetTicketId(ticketChannel.Id, ticketIdChan)
	ticketId := <- ticketIdChan

	// Verify that the user is allowed to modify the ticketChannel
	permLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Shard, ctx.Member, ctx.Guild.Id, permLevelChan)
	permLevel := <-permLevelChan

	ownerChan := make(chan uint64)
	go database.GetOwner(ticketId, ctx.Guild.Id, ownerChan)
	owner := <-ownerChan

	if permLevel == 0 && owner != ctx.User.Id {
		ctx.SendEmbed(utils.Red, "Error", "You don't have permission to add people to this ticketChannel")
		ctx.ReactWithCross()
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Add user to ticketChannel in DB
		go database.AddMember(ticketId, ctx.Guild.Id, user.Id)

		if err := ctx.Shard.EditChannelPermissions(ticketChannel.Id, channel.PermissionOverwrite{
			Id:    user.Id,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
		}); err != nil {
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

func (AddCommand) Category() Category {
	return Tickets
}

func (AddCommand) AdminOnly() bool {
	return false
}

func (AddCommand) HelperOnly() bool {
	return false
}
