package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"runtime"
)

// Reset
type AdminGCCommand struct {
}

func (AdminGCCommand) Name() string {
	return "gc"
}

func (AdminGCCommand) Description() string {
	return "Forces a GC sweep"
}

func (AdminGCCommand) Aliases() []string {
	return []string{}
}

func (AdminGCCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminGCCommand) Execute(ctx utils.CommandContext) {
	runtime.GC()
	ctx.ReactWithCheck()
}

func (AdminGCCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminGCCommand) Children() []Command {
	return []Command{}
}

func (AdminGCCommand) PremiumOnly() bool {
	return false
}

func (AdminGCCommand) Category() Category {
	return Settings
}

func (AdminGCCommand) AdminOnly() bool {
	return true
}

func (AdminGCCommand) HelperOnly() bool {
	return false
}
