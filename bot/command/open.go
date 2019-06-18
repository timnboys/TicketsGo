package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

type OpenCommand struct {
}

func (OpenCommand) Name() string {
	return "open"
}

func (OpenCommand) Description() string {
	return "Opens a new ticket"
}

func (OpenCommand) Aliases() []string {
	return []string{"new"}
}

func (OpenCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (OpenCommand) Execute(ctx CommandContext) {

}

func (OpenCommand) Parent() *Command {
	return nil
}

func (OpenCommand) Children() []Command {
	return make([]Command, 0)
}

func (OpenCommand) PremiumOnly() bool {
	return false
}

func (OpenCommand) AdminOnly() bool {
	return false
}

func (OpenCommand) HelperOnly() bool {
	return false
}
