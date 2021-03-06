package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

type Command interface {
	Name() string
	Description() string
	Aliases() []string
	PermissionLevel() utils.PermissionLevel
	Execute(ctx utils.CommandContext)
	Parent() interface{}
	Children() []Command
	PremiumOnly() bool
	Category() Category
	AdminOnly() bool
	HelperOnly() bool
}
