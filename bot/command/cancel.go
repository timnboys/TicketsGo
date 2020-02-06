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
		Guild: ctx.Guild.ID,
		User: ctx.User.ID,
		Channel: ctx.Channel,
		Session: ctx.Session,
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

func (CancelCommand) AdminOnly() bool {
	return false
}

func (CancelCommand) HelperOnly() bool {
	return false
}
