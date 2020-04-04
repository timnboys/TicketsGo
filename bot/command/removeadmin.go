package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"strings"
)

type RemoveAdminCommand struct {
}

func (RemoveAdminCommand) Name() string {
	return "removeadmin"
}

func (RemoveAdminCommand) Description() string {
	return "Revokes a user's or role's admin privileges"
}

func (RemoveAdminCommand) Aliases() []string {
	return []string{}
}

func (RemoveAdminCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (RemoveAdminCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to revoke admin privileges from")
		ctx.ReactWithCross()
		return
	}

	roles := make([]uint64, 0)
	if len(ctx.Message.Mentions) > 0 {
		for _, mention := range ctx.Message.Mentions {
			if ctx.Guild.OwnerId == mention.Id {
				ctx.SendEmbed(utils.Red, "Error", "The guild owner must be an admin")
				continue
			}

			if ctx.User.Id == mention.Id {
				ctx.SendEmbed(utils.Red, "Error", "You cannot revoke your own privileges")
				continue
			}

			go database.RemoveAdmin(ctx.Guild.Id, mention.Id)
		}
	} else if len(ctx.Message.MentionRoles) > 0 {
		for _, mention := range ctx.Message.MentionRoles {
			roles = append(roles, mention)
		}
	} else {
		roleName := strings.ToLower(strings.Join(ctx.Args, " "))

		// Get role ID from name
		valid := false
		for _, role := range ctx.Guild.Roles {
			if strings.ToLower(role.Name) == roleName {
				roles = append(roles, role.Id)
				valid = true
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to revoke admin privileges from")
			ctx.ReactWithCross()
			return
		}
	}

	// Remove roles from DB
	for _, role := range roles {
		go database.RemoveAdminRole(ctx.Guild.Id, role)
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

func (RemoveAdminCommand) Category() Category {
	return Settings
}

func (RemoveAdminCommand) AdminOnly() bool {
	return false
}

func (RemoveAdminCommand) HelperOnly() bool {
	return false
}
