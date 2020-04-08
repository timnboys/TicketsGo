package command

import (
	"github.com/TicketsBot/TicketsGo/bot/setup"
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

type CancelCommand struct {
}

func (CancelCommand) Name() string {
	return "cancel"
}

func (CancelCommand) Description() string {
	return "Cancels the setup process"
}

func (CancelCommand) Aliases() []string {
	return []string{}
}

func (CancelCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (CancelCommand) Execute(ctx utils.CommandContext) {
	u := setup.SetupUser{
		Guild: ctx.GuildId,
		User: ctx.Author.Id,
		Channel: ctx.ChannelId,
		Session: ctx.Shard,
	}

	// Check if the user is in the setup process
	if !u.InSetup() {
		ctx.ReactWithCross()
		return
	}

	u.Cancel()
	ctx.ReactWithCheck()
}

func (CancelCommand) Parent() interface{} {
	return nil
}

func (CancelCommand) Children() []Command {
	return make([]Command, 0)
}

func (CancelCommand) PremiumOnly() bool {
	return false
}

func (CancelCommand) Category() Category {
	return Settings
}

func (CancelCommand) AdminOnly() bool {
	return false
}

func (CancelCommand) HelperOnly() bool {
	return false
}
