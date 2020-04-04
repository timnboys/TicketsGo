package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"runtime"
)

type AdminStatsCommand struct {
}

func (AdminStatsCommand) Name() string {
	return "stats"
}

func (AdminStatsCommand) Description() string {
	return "Gets hardware usage statistics"
}

func (AdminStatsCommand) Aliases() []string {
	return []string{}
}

func (AdminStatsCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminStatsCommand) Execute(ctx utils.CommandContext) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	embed := embed.NewEmbed().
		SetTitle("Admin").
		SetColor(int(utils.Green)).

		AddField("Heap", fmt.Sprintf("%dMB", m.Alloc / 1024 / 1024), true).
		AddField("Stack", fmt.Sprintf("%dMB", m.StackSys / 1024 / 1024), true).
		AddField("Total Reserved", fmt.Sprintf("%dMB", m.Sys / 1024 / 1024), true)

	msg, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: &msg}, 30)
}

func (AdminStatsCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminStatsCommand) Children() []Command {
	return []Command{}
}

func (AdminStatsCommand) PremiumOnly() bool {
	return false
}

func (AdminStatsCommand) Category() Category {
	return Settings
}

func (AdminStatsCommand) AdminOnly() bool {
	return false
}

func (AdminStatsCommand) HelperOnly() bool {
	return true
}
