package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/rxdn/gdl/rest"
)

type AdminCheckPermsCommand struct {
}

func (AdminCheckPermsCommand) Name() string {
	return "checkperms"
}

func (AdminCheckPermsCommand) Description() string {
	return "Checks permissions for the bot on the channel"
}

func (AdminCheckPermsCommand) Aliases() []string {
	return []string{"cp"}
}

func (AdminCheckPermsCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminCheckPermsCommand) Execute(ctx utils.CommandContext) {
	guild, found := ctx.Shard.Cache.GetGuild(ctx.GuildId, false)

	if !found {
		ctx.SendMessage("guild not cached")
	} else {
		g, err := rest.GetGuild(ctx.Shard.Token, ctx.Shard.ShardManager.RateLimiter, ctx.GuildId)
		if err != nil {
			ctx.SendMessage(err.Error())
			return
		}
		guild = g
	}

	ctx.SendMessage(fmt.Sprintf("roles: %d", len(guild.Roles)))

	for _, role := range guild.Roles {
		ctx.SendMessage(fmt.Sprintf("role %s: %d", role.Name, role.Permissions))
	}
}

func (AdminCheckPermsCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminCheckPermsCommand) Children() []Command {
	return []Command{}
}

func (AdminCheckPermsCommand) PremiumOnly() bool {
	return false
}

func (AdminCheckPermsCommand) Category() Category {
	return Settings
}

func (AdminCheckPermsCommand) AdminOnly() bool {
	return false
}

func (AdminCheckPermsCommand) HelperOnly() bool {
	return true
}
