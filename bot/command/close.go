package command

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
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
	logic.CloseTicket(ctx.Shard, ctx.GuildId, ctx.ChannelId, ctx.Id, ctx.Member, ctx.Args, false, ctx.IsPremium)
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
