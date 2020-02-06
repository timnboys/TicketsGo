package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"strings"
)

type RemoveSupportCommand struct {
}

func (RemoveSupportCommand) Name() string {
	return "removesupport"
}

func (RemoveSupportCommand) Description() string {
	return "Revokes a user's or role's support representative privileges"
}

func (RemoveSupportCommand) Aliases() []string {
	return []string{}
}

func (RemoveSupportCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (RemoveSupportCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to revoke support representative privileges from")
		ctx.ReactWithCross()
		return
	}

	var roleId string
	if len(ctx.Message.Mentions) > 0 { // Individual users
		for _, mention := range ctx.Message.Mentions {
			// Verify that we're allowed to perform the remove operation
			if ctx.Guild.OwnerID == mention.ID {
				ctx.SendEmbed(utils.Red, "Error", "The guild owner must be an admin")
				continue
			}

			if ctx.User.ID == mention.ID {
				ctx.SendEmbed(utils.Red, "Error", "You cannot revoke your own privileges")
				continue
			}

			go database.RemoveSupport(ctx.Guild.ID, mention.ID)
		}
	} else { // Role
		roleName := strings.ToLower(ctx.Args[0])

		// Get role ID from name
		for _, role := range ctx.Guild.Roles {
			if strings.ToLower(role.Name) == roleName {
				roleId = role.ID
				break
			}
		}

		// Verify a valid role was mentioned
		if roleId == "" {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to revoke support representative privileges from")
			ctx.ReactWithCross()
			return
		}

		go database.RemoveSupportRole(ctx.Guild.ID, roleId)
	}

	ctx.ReactWithCheck()
}

func (RemoveSupportCommand) Parent() interface{} {
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
