package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/setup"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
)

type SetupCommand struct {
}

func (SetupCommand) Name() string {
	return "setup"
}

func (SetupCommand) Description() string {
	return "Allows you to easily configure the bot"
}

func (SetupCommand) Aliases() []string {
	return []string{}
}

func (SetupCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (SetupCommand) Execute(ctx utils.CommandContext) {
	u := setup.SetupUser{
		Guild:   ctx.Guild.Id,
		User:    ctx.User.Id,
		Channel: ctx.ChannelId,
		Session: ctx.Shard,
	}

	if u.InSetup() {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("You are already in setup mode (use `%scancel` to exit)", config.Conf.Bot.Prefix))
	} else {
		ctx.ReactWithCheck()

		u.Next()
		state := u.GetState()
		stage := state.GetStage()
		if stage != nil {
			// Psuedo-premium
			utils.SendEmbed(ctx.Shard, ctx.ChannelId, utils.Green, "Setup", (*stage).Prompt(), 120, true)
		}
	}
}

func (SetupCommand) Parent() interface{} {
	return nil
}

func (SetupCommand) Children() []Command {
	return make([]Command, 0)
}

func (SetupCommand) PremiumOnly() bool {
	return false
}

func (SetupCommand) Category() Category {
	return Settings
}

func (SetupCommand) AdminOnly() bool {
	return false
}

func (SetupCommand) HelperOnly() bool {
	return false
}
