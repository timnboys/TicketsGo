package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

// Reset
type AdminResetCommand struct {
}

func (AdminResetCommand) Name() string {
	return "reset"
}

func (AdminResetCommand) Description() string {
	return "Purges the guild's cached objects"
}

func (AdminResetCommand) Aliases() []string {
	return []string{}
}

func (AdminResetCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminResetCommand) Execute(ctx utils.CommandContext) {
	ctx.SendEmbed(utils.Red, "Admin", "Not yet implemented")
}

func (AdminResetCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminResetCommand) Children() []Command {
	return []Command{}
}

func (AdminResetCommand) PremiumOnly() bool {
	return false
}

func (AdminResetCommand) AdminOnly() bool {
	return false
}

func (AdminResetCommand) HelperOnly() bool {
	return true
}
