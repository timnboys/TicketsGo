package command

import "github.com/TicketsBot/TicketsGo/bot/utils"

type PurgeCommand struct {
}

func (PurgeCommand) Name() string {
	return "purge"
}

func (PurgeCommand) Description() string {
	return "Automatically close tickets which have not received a response in a certain amount of time"
}

func (PurgeCommand) Aliases() []string {
	return []string{}
}

func (PurgeCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (PurgeCommand) Execute(ctx utils.CommandContext) {
}

func (PurgeCommand) Parent() interface{} {
	return nil
}

func (PurgeCommand) Children() []Command {
	return make([]Command, 0)
}

func (PurgeCommand) PremiumOnly() bool {
	return true
}

func (PurgeCommand) AdminOnly() bool {
	return false
}

func (PurgeCommand) HelperOnly() bool {
	return false
}
