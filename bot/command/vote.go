package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
)

type VoteCommand struct {
}

func (VoteCommand) Name() string {
	return "vote"
}

func (VoteCommand) Description() string {
	return "Gives you a link to vote for free premium"
}

func (VoteCommand) Aliases() []string {
	return []string{}
}

func (VoteCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (VoteCommand) Execute(ctx utils.CommandContext) {
	ctx.ReactWithCheck()
	ctx.SendEmbed(utils.Green, "Vote", "Click here to vote for 24 hours of free premium:\nhttps://vote.ticketsbot.net")
}

func (VoteCommand) Parent() interface{} {
	return nil
}

func (VoteCommand) Children() []Command {
	return make([]Command, 0)
}

func (VoteCommand) PremiumOnly() bool {
	return false
}

func (VoteCommand) Category() Category {
	return General
}

func (VoteCommand) AdminOnly() bool {
	return false
}

func (VoteCommand) HelperOnly() bool {
	return false
}
