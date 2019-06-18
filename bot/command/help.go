package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/apex/log"
	"strings"
)

type HelpCommand struct {
}

func (HelpCommand) Name() string {
	return "help"
}

func (HelpCommand) Description() string {
	return "Shows you a list of commands"
}

func (HelpCommand) Aliases() []string {
	return []string{"h"}
}

func (HelpCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (HelpCommand) Execute(ctx CommandContext) {
	msg := ""
	for _, cmd := range Commands {
		msg += fmt.Sprintf("`%s%s` - %s\n", config.Conf.Bot.Prefix, cmd.Name(), cmd.Description())
	}
	msg = strings.Trim(msg, "\n")

	ch, err := ctx.Session.UserChannelCreate(ctx.User.ID); if err != nil {
		log.Error(err.Error())
		return
	}

	if ch != nil {
		utils.SendEmbed(ctx.Session, ch.ID, utils.Green, "Help", msg, 0, ctx.IsPremium)
	}

	ctx.ReactWithCheck()
}

func (HelpCommand) Parent() interface{} {
	return nil
}

func (HelpCommand) Children() []Command {
	return make([]Command, 0)
}

func (HelpCommand) PremiumOnly() bool {
	return false
}

func (HelpCommand) AdminOnly() bool {
	return false
}

func (HelpCommand) HelperOnly() bool {
	return false
}
