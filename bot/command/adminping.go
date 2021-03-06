package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

type AdminPingCommand struct {
}

func (AdminPingCommand) Name() string {
	return "ping"
}

func (AdminPingCommand) Description() string {
	return "Measures WS latency to Discord"
}

func (AdminPingCommand) Aliases() []string {
	return []string{"latency"}
}

func (AdminPingCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminPingCommand) Execute(ctx utils.CommandContext) {
	latency := ctx.Shard.HeartbeatLatency()
	ctx.SendEmbed(utils.Green, "Admin", fmt.Sprintf("Shard %d latency: `%dms`", ctx.Shard.ShardId, latency))
}

func (AdminPingCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminPingCommand) Children() []Command {
	return []Command{}
}

func (AdminPingCommand) PremiumOnly() bool {
	return false
}

func (AdminPingCommand) Category() Category {
	return Settings
}

func (AdminPingCommand) AdminOnly() bool {
	return false
}

func (AdminPingCommand) HelperOnly() bool {
	return true
}
