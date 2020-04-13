package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"strings"
)

type ManageTags struct {
}

func (ManageTags) Name() string {
	return "managetags"
}

func (ManageTags) Description() string {
	return "Command for managing tags"
}

func (ManageTags) Aliases() []string {
	return []string{"managecannedresponse", "managecannedresponses", "editcannedresponse", "editcannedresponses", "ecr", "managetags", "mcr", "managetag", "mt"}
}

func (ManageTags) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (ManageTags) Execute(ctx utils.CommandContext) {
	msg := "Select a subcommand:\n"

	children := ManageTags{}.Children()
	for _, child := range children {
		msg += fmt.Sprintf("`%smcr %s` - %s\n", config.Conf.Bot.Prefix, child.Name(), child.Description())
	}

	msg = strings.TrimSuffix(msg, "\n")

	ctx.SendEmbed(utils.Red, "Error", msg)
}

func (ManageTags) Parent() interface{} {
	return ManageTags{}
}

func (ManageTags) Children() []Command {
	return []Command{
		ManageTagsAdd{},
		ManageTagsDelete{},
		ManageCannedResponsesList{},
	}
}

func (ManageTags) PremiumOnly() bool {
	return false
}

func (ManageTags) Category() Category {
	return Tags
}

func (ManageTags) AdminOnly() bool {
	return false
}

func (ManageTags) HelperOnly() bool {
	return false
}
