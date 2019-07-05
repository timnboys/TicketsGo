package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
)

type RemoveAdminCommand struct {
}

func (RemoveAdminCommand) Name() string {
	return "removeadmin"
}

func (RemoveAdminCommand) Description() string {
	return "Revokes a user's admin privileges"
}

func (RemoveAdminCommand) Aliases() []string {
	return []string{}
}

func (RemoveAdminCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (RemoveAdminCommand) Execute(ctx CommandContext) {
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to revoke admin privileges from")
		ctx.ReactWithCross()
		return
	}

	// Get guild obj
	guild, err := ctx.Session.Guild(ctx.Guild); if err != nil {
		sentry.Error(err)
		ctx.ReactWithCross()
		return
	}

	for _, mention := range ctx.Message.Mentions {
		if guild.OwnerID == mention.ID {
			ctx.SendEmbed(utils.Red, "Error", "The guild owner must be an admin")
			continue
		}

		if ctx.User.ID == mention.ID {
			ctx.SendEmbed(utils.Red, "Error", "You cannot revoke your own privileges")
			continue
		}

		go database.RemoveAdmin(ctx.Guild, mention.ID)
	}
	ctx.ReactWithCheck()
}

func (RemoveAdminCommand) Parent() interface{} {
	return nil
}

func (RemoveAdminCommand) Children() []Command {
	return make([]Command, 0)
}

func (RemoveAdminCommand) PremiumOnly() bool {
	return false
}

func (RemoveAdminCommand) AdminOnly() bool {
	return false
}

func (RemoveAdminCommand) HelperOnly() bool {
	return false
}
