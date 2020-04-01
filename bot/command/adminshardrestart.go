package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
)

// Reset
type AdminShardRestartCommand struct {
}

func (AdminShardRestartCommand) Name() string {
	return "shardrestart"
}

func (AdminShardRestartCommand) Description() string {
	return "Reconnects a shard to the websocket"
}

func (AdminShardRestartCommand) Aliases() []string {
	return []string{"sr", "restartshard", "rs"}
}

func (AdminShardRestartCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminShardRestartCommand) Execute(ctx utils.CommandContext) {
	if err := ctx.Shard.Close(); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if err := ctx.Shard.Open(); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	ctx.SendEmbed(utils.Green, "Admin", "Restarted")
}

func (AdminShardRestartCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminShardRestartCommand) Children() []Command {
	return []Command{}
}

func (AdminShardRestartCommand) PremiumOnly() bool {
	return false
}

func (AdminShardRestartCommand) AdminOnly() bool {
	return false
}

func (AdminShardRestartCommand) HelperOnly() bool {
	return true
}
