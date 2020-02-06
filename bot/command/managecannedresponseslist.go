package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
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
	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	idsChan := make(chan []string)
	go database.GetCannedResponses(guildId, idsChan)
	
	var joined string
	for _, id := range <-idsChan {
		joined += fmt.Sprintf("- `%s`\n", id)
	}
	joined = strings.TrimSuffix(joined, "\n")

	ctx.SendEmbed(utils.Green, "Canned Responses", fmt.Sprintf("IDs for all canned responses:\n%s\nTo view the contents of a canned response, run `%sc <ID>`", joined, config.Conf.Bot.Prefix))
}

func (ManageCannedResponsesList) Parent() interface{} {
	return ManageCannedResponses{}
}

func (ManageCannedResponsesList) Children() []Command {
	return make([]Command, 0)
}

func (ManageCannedResponsesList) PremiumOnly() bool {
	return false
}

func (ManageCannedResponsesList) AdminOnly() bool {
	return false
}

func (ManageCannedResponsesList) HelperOnly() bool {
	return false
}
