package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
)

type RemoveSupportCommand struct {
}

func (RemoveSupportCommand) Name() string {
	return "removesupport"
}

func (RemoveSupportCommand) Description() string {
	return "Revokes a user's support representative privileges"
}

func (RemoveSupportCommand) Aliases() []string {
	return []string{}
}

func (RemoveSupportCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (RemoveSupportCommand) Execute(ctx CommandContext) {
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to revoke support representative privileges from")
		ctx.ReactWithCross()
		return
	}

	// Get guild obj
	guild, err := ctx.Session.Guild(ctx.Guild); if err != nil {
		log.Error(err.Error())
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

		go database.RemoveSupport(ctx.Guild, mention.ID)
	}
	ctx.ReactWithCheck()
}

func (RemoveSupportCommand) Parent() *Command {
	return nil
}

func (RemoveSupportCommand) Children() []Command {
	return make([]Command, 0)
}

func (RemoveSupportCommand) PremiumOnly() bool {
	return false
}

func (RemoveSupportCommand) AdminOnly() bool {
	return false
}

func (RemoveSupportCommand) HelperOnly() bool {
	return false
}
