package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
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

func (a AddCommand) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!add @User #ticket-channel`",
		Inline: false,
	}

	// Check users are mentioned
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention members to add to the ticket", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// Check channel is mentioned
	ticketChannel := ctx.GetChannelFromArgs()
	if ticketChannel == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a ticket channel to add the user(s) in", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	ticket, err := database.Client.Tickets.GetByChannel(ticketChannel)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// 2 in 1: verify guild is the same & the channel is valid
	if ticket.GuildId != ctx.GuildId {
		ctx.SendEmbed(utils.Red, "Error", "The mentioned channel is not a ticket", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// Get ticket ID
	owner := make(chan uint64)

	// Verify that the user is allowed to modify the ticket
	if ctx.UserPermissionLevel == 0 && <-owner != ctx.Author.Id {
		ctx.SendEmbed(utils.Red, "Error", "You don't have permission to add people to this ticket")
		ctx.ReactWithCross()
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Add user to ticket in DB
		go func() {
			if err := database.Client.TicketMembers.Add(ctx.GuildId, ticket.Id, user.Id); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}()

		if err := ctx.Shard.EditChannelPermissions(ticketChannel, channel.PermissionOverwrite{
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
