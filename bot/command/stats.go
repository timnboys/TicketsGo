package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

type StatsCommand struct {
}

func (StatsCommand) Name() string {
	return "stats"
}

func (StatsCommand) Description() string {
	return "Shows you statistics about users, support staff and the server"
}

func (StatsCommand) Aliases() []string {
	return []string{"statistics"}
}

func (StatsCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (StatsCommand) Execute(ctx CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must specify `server` to view server statistics, or tag a user to view their statistics")
		ctx.ReactWithCross()
		return
	}

	// server is handled as a subcommand, so a user has been pinged
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must specify `server` to view server statistics, or tag a user to view their statistics")
		ctx.ReactWithCross()
		return
	}

	// Get user permission level
	user := ctx.Message.Mentions[0]
	permLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Session, ctx.Guild, user.ID, permLevelChan)
	permLevel := <-permLevelChan

	// User stats
	if permLevel == 0 {

	} else { // Support rep stats

	}
}

func (StatsCommand) Parent() interface{} {
	return nil
}

func (StatsCommand) Children() []Command {
	return []Command{
		StatsServerCommand{},
	}
}

func (StatsCommand) PremiumOnly() bool {
	return true
}

func (StatsCommand) AdminOnly() bool {
	return false
}

func (StatsCommand) HelperOnly() bool {
	return false
}
