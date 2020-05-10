package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
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
	user.Member.User = user.User

	if ctx.Author.Id == user.Id {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist yourself")
		ctx.ReactWithCross()
		return
	}

	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Shard, user.Member, ctx.GuildId, permissionLevelChan)
	permissionLevel := <- permissionLevelChan

	if permissionLevel > utils.Everyone {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist staff")
		ctx.ReactWithCross()
		return
	}

	isBlacklisted, err := database.Client.Blacklist.IsBlacklisted(ctx.GuildId, user.Id)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.ReactWithCross()
		return
	}

	if isBlacklisted {
		if err := database.Client.Blacklist.Remove(ctx.GuildId, user.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}
	} else {
		if err := database.Client.Blacklist.Add(ctx.GuildId, user.Id); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ctx.ReactWithCross()
			return
		}
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
