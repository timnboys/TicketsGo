package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/objects/channel/embed"
)

type BlacklistCommand struct {
}

func (BlacklistCommand) Name() string {
	return "blacklist"
}

func (BlacklistCommand) Description() string {
	return "Toggles whether users are allowed to interact with the bot"
}

func (BlacklistCommand) Aliases() []string {
	return []string{"unblacklist"}
}

func (BlacklistCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (BlacklistCommand) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!blacklist @User`",
		Inline: false,
	}

	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to toggle the blacklist state for", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]

	if ctx.Author.Id == user.Id {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist yourself")
		ctx.ReactWithCross()
		return
	}

	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Shard, ctx.Member, ctx.GuildId, permissionLevelChan)
	permissionLevel := <- permissionLevelChan

	if permissionLevel > 0 {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist staff")
		ctx.ReactWithCross()
		return
	}

	isBlacklistedChan := make(chan bool)
	go database.IsBlacklisted(ctx.GuildId, user.Id, isBlacklistedChan)
	isBlacklisted := <- isBlacklistedChan

	if isBlacklisted {
		go database.RemoveBlacklist(ctx.GuildId, user.Id)
	} else {
		go database.AddBlacklist(ctx.GuildId, user.Id)
	}

	ctx.ReactWithCheck()
}

func (BlacklistCommand) Parent() interface{} {
	return nil
}

func (BlacklistCommand) Children() []Command {
	return make([]Command, 0)
}

func (BlacklistCommand) PremiumOnly() bool {
	return false
}

func (BlacklistCommand) Category() Category {
	return Settings
}

func (BlacklistCommand) AdminOnly() bool {
	return false
}

func (BlacklistCommand) HelperOnly() bool {
	return false
}
