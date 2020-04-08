package command

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/rxdn/gdl/objects/channel/message"
)

type OpenCommand struct {
}

func (OpenCommand) Name() string {
	return "open"
}

func (OpenCommand) Description() string {
	return "Opens a new ticket"
}

func (OpenCommand) Aliases() []string {
	return []string{"new"}
}

func (OpenCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (OpenCommand) Execute(ctx utils.CommandContext) {
	logic.OpenTicket(ctx.Shard, ctx.Author, message.MessageReference{
		MessageId: ctx.Id,
		ChannelId: ctx.ChannelId,
		GuildId:   ctx.GuildId,
	}, ctx.IsPremium, ctx.Args, nil)
}


func (OpenCommand) Parent() interface{} {
	return nil
}

func (OpenCommand) Children() []Command {
	return make([]Command, 0)
}

func (OpenCommand) PremiumOnly() bool {
	return false
}

func (OpenCommand) Category() Category {
	return Tickets
}

func (OpenCommand) AdminOnly() bool {
	return false
}

func (OpenCommand) HelperOnly() bool {
	return false
}