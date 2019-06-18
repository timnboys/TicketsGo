package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
)

type AddAdminCommand struct {
}

func (AddAdminCommand) Name() string {
	return "addadmin"
}

func (AddAdminCommand) Description() string {
	return "Grants a user admin privileges"
}

func (AddAdminCommand) Aliases() []string {
	return []string{}
}

func (AddAdminCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (AddAdminCommand) Execute(ctx CommandContext) {
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to grant admin privileges to")
		ctx.ReactWithCross()
		return
	}

	for _, mention := range ctx.Message.Mentions {
		go database.AddAdmin(ctx.Guild, mention.ID)
	}
	ctx.ReactWithCheck()
}

func (AddAdminCommand) Parent() interface{} {
	return nil
}

func (AddAdminCommand) Children() []Command {
	return make([]Command, 0)
}

func (AddAdminCommand) PremiumOnly() bool {
	return false
}

func (AddAdminCommand) AdminOnly() bool {
	return false
}

func (AddAdminCommand) HelperOnly() bool {
	return false
}
