package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

// Reset
type AdminUsersCommand struct {
}

func (AdminUsersCommand) Name() string {
	return "users"
}

func (AdminUsersCommand) Description() string {
	return "Prints the instance's total user count"
}

func (AdminUsersCommand) Aliases() []string {
	return []string{}
}

func (AdminUsersCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminUsersCommand) Execute(ctx utils.CommandContext) {
	count := 0
	for _, guild := range ctx.Session.State.Guilds {
		count += guild.MemberCount
	}
	ctx.SendEmbed(utils.Green, "Admin", fmt.Sprintf("There are %d users on this instance", count))
}

func (AdminUsersCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminUsersCommand) Children() []Command {
	return []Command{}
}

func (AdminUsersCommand) PremiumOnly() bool {
	return false
}

func (AdminUsersCommand) AdminOnly() bool {
	return true
}

func (AdminUsersCommand) HelperOnly() bool {
	return false
}
