package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
)

type AddSupportCommand struct {
}

func (AddSupportCommand) Name() string {
	return "addsupport"
}

func (AddSupportCommand) Description() string {
	return "Adds a user as a support representative"
}

func (AddSupportCommand) Aliases() []string {
	return []string{}
}

func (AddSupportCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (AddSupportCommand) Execute(ctx CommandContext) {
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user to grant support representative privileges to")
		ctx.ReactWithCross()
		return
	}

	for _, mention := range ctx.Message.Mentions {
		go database.AddSupport(ctx.Guild, mention.ID)
	}
	ctx.ReactWithCheck()
}

func (AddSupportCommand) Parent() interface{} {
	return nil
}

func (AddSupportCommand) Children() []Command {
	return make([]Command, 0)
}

func (AddSupportCommand) PremiumOnly() bool {
	return false
}

func (AddSupportCommand) AdminOnly() bool {
	return false
}

func (AddSupportCommand) HelperOnly() bool {
	return false
}
