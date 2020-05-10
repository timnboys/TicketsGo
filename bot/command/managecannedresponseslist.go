package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strings"
)

type ManageCannedResponsesList struct {
}

func (ManageCannedResponsesList) Name() string {
	return "list"
}

func (ManageCannedResponsesList) Description() string {
	return "Lists all canned responses"
}

func (ManageCannedResponsesList) Aliases() []string {
	return []string{}
}

func (ManageCannedResponsesList) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (ManageCannedResponsesList) Execute(ctx utils.CommandContext) {
	ids, err := database.Client.Tag.GetTagIds(ctx.GuildId)
	if err != nil {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	var joined string
	for _, id := range ids {
		joined += fmt.Sprintf("• `%s`\n", id)
	}
	joined = strings.TrimSuffix(joined, "\n")

	ctx.SendEmbed(utils.Green, "Canned Responses", fmt.Sprintf("IDs for all tags:\n%s\nTo view the contents of a tag, run `%stag <ID>`", joined, config.Conf.Bot.Prefix))
}

func (ManageCannedResponsesList) Parent() interface{} {
	return ManageTags{}
}

func (ManageCannedResponsesList) Children() []Command {
	return make([]Command, 0)
}

func (ManageCannedResponsesList) PremiumOnly() bool {
	return false
}

func (ManageCannedResponsesList) Category() Category {
	return Tags
}

func (ManageCannedResponsesList) AdminOnly() bool {
	return false
}

func (ManageCannedResponsesList) HelperOnly() bool {
	return false
}
