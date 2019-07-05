package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
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

func (BlacklistCommand) Execute(ctx CommandContext) {
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to toggle the blacklist state for")
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]

	if ctx.User.ID == user.ID {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist yourself")
		ctx.ReactWithCross()
		return
	}

	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Session, ctx.Guild, user.ID, permissionLevelChan)
	permissionLevel := <- permissionLevelChan

	if permissionLevel > 0 {
		ctx.SendEmbed(utils.Red, "Error", "You cannot blacklist staff")
		ctx.ReactWithCross()
		return
	}

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	isBlacklistedChan := make(chan bool)
	go database.IsBlacklisted(guildId, userId, isBlacklistedChan)
	isBlacklisted := <- isBlacklistedChan

	if isBlacklisted {
		go database.RemoveBlacklist(guildId, userId)
	} else {
		go database.AddBlacklist(guildId, userId)
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

func (BlacklistCommand) AdminOnly() bool {
	return false
}

func (BlacklistCommand) HelperOnly() bool {
	return false
}
