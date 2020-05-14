package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
)

type ClaimCommand struct {
}

func (ClaimCommand) Name() string {
	return "claim"
}

func (ClaimCommand) Description() string {
	return "Assigns a single staff member to a ticket"
}

func (ClaimCommand) Aliases() []string {
	return []string{}
}

func (ClaimCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (c ClaimCommand) Execute(ctx utils.CommandContext) {
	// Get ticket struct
	ticket, err := database.Client.Tickets.GetByChannel(ctx.ChannelId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", "This is not a ticket channel")
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Shard, ticket, ctx.Author.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", fmt.Sprintf("Your ticket will be handled by %s", ctx.Author.Mention()))
	ctx.ReactWithCheck()
}


func (ClaimCommand) Parent() interface{} {
	return nil
}

func (ClaimCommand) Children() []Command {
	return make([]Command, 0)
}

func (ClaimCommand) PremiumOnly() bool {
	return false
}

func (ClaimCommand) Category() Category {
	return Tickets
}

func (ClaimCommand) AdminOnly() bool {
	return false
}

func (ClaimCommand) HelperOnly() bool {
	return false
}
