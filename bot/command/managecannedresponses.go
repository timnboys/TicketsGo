package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"strings"
)

type ManageCannedResponses struct {
}

func (ManageCannedResponses) Name() string {
	return "mcr"
}

func (ManageCannedResponses) Description() string {
	return "Command for managing canned responses"
}

func (ManageCannedResponses) Aliases() []string {
	return []string{"managecannedresponse", "managecannedresponses", "editcannedresponse", "editcannedresponses", "ecr"}
}

func (ManageCannedResponses) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (ManageCannedResponses) Execute(ctx utils.CommandContext) {
	msg := "Select a subcommand:\n"

	children := ManageCannedResponses{}.Children()
	for _, child := range children {
		msg += fmt.Sprintf("`%smcr %s` - %s\n", config.Conf.Bot.Prefix, child.Name(), child.Description())
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.SendEmbed(utils.Red, "Error", msg)
}

func (ManageCannedResponses) Parent() interface{} {
	return ManageCannedResponses{}
}

func (ManageCannedResponses) Children() []Command {
	return []Command{
		ManageCannedResponsesAdd{},
		ManageCannedResponsesDelete{},
		ManageCannedResponsesList{},
	}
}

func (ManageCannedResponses) PremiumOnly() bool {
	return false
}

func (ManageCannedResponses) AdminOnly() bool {
	return false
}

func (ManageCannedResponses) HelperOnly() bool {
	return false
}
