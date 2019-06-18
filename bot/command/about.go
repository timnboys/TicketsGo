package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
)

type AboutCommand struct {
}

func (AboutCommand) Name() string {
	return "about"
}

func (AboutCommand) Description() string {
	return "Tells you information about the bot"
}

func (AboutCommand) Aliases() []string {
	return []string{}
}

func (AboutCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AboutCommand) Execute(ctx CommandContext) {
	ctx.SendEmbed(utils.Green, "About", config.Conf.AboutMessage)
}

func (AboutCommand) Parent() *Command {
	return nil
}

func (AboutCommand) Children() []Command {
	return make([]Command, 0)
}

func (AboutCommand) PremiumOnly() bool {
	return false
}

func (AboutCommand) AdminOnly() bool {
	return false
}

func (AboutCommand) HelperOnly() bool {
	return false
}
