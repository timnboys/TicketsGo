package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"strings"
)

type AdminCommand struct {
}

func (AdminCommand) Name() string {
	return "admin"
}

func (AdminCommand) Description() string {
	return "Bot management"
}

func (AdminCommand) Aliases() []string {
	return []string{"a"}
}

func (AdminCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminCommand) Execute(ctx utils.CommandContext) {
	msg := "Select a subcommand:\n"

	children := AdminCommand{}.Children()
	for _, child := range children {
		msg += fmt.Sprintf("`%sadmin %s` - %s\n", config.Conf.Bot.Prefix, child.Name(), child.Description())
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.SendEmbed(utils.Green, "Admin", msg)
}

func (AdminCommand) Parent() interface{} {
	return nil
}

func (AdminCommand) Children() []Command {
	return []Command{
		AdminCheckPermsCommand{},
		AdminDebugCommand{},
		AdminGeneratePremium{},
		AdminPingCommand{},
		AdminResetCommand{},
		AdminSeedMetrics{},
		AdminShardRestartCommand{},
		AdminStatsCommand{},
		AdminUsersCommand{},
	}
}

func (AdminCommand) PremiumOnly() bool {
	return false
}

func (AdminCommand) Category() Category {
	return Settings
}

func (AdminCommand) AdminOnly() bool {
	return false
}

func (AdminCommand) HelperOnly() bool {
	return true
}
