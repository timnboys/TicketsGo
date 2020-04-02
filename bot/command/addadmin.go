package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type AddAdminCommand struct {
}

func (AddAdminCommand) Name() string {
	return "addadmin"
}

func (AddAdminCommand) Description() string {
	return "Grants a user or role admin privileges"
}

func (AddAdminCommand) Aliases() []string {
	return []string{}
}

func (AddAdminCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (AddAdminCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant admin privileges to")
		ctx.ReactWithCross()
		return
	}

	user := false
	roles := make([]uint64, 0)

	if len(ctx.Message.Mentions) > 0 {
		user = true
		for _, mention := range ctx.Message.Mentions {
			go database.AddAdmin(ctx.Guild.Id, mention.Id)
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
				valid = true
				roles = append(roles, role.Id)
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant admin privileges to")
			ctx.ReactWithCross()
			return
		}
	}

	// Add roles to DB
	for _, role := range roles {
		go database.AddAdminRole(ctx.Guild.Id, role)
	}

	openTicketsChan := make(chan []*uint64)
	go database.GetOpenTicketChannelIds(ctx.Guild.Id, openTicketsChan)

	// Update permissions for existing tickets
	for _, channelId := range <-openTicketsChan {
		// Mitigation for a very rare panic that occurs when this command is run whilst a ticket is being opened, but
		// the channel ID hasn't been set in the database yet, or if Discord is dead and won't let the channel
		// be created.
		if channelId == nil {
			continue
		}

		ch, err := ctx.Shard.GetChannel(*channelId); if err != nil {
			continue
		}

		overwrites := ch.PermissionOverwrites

		if user {
			// If adding individual admins, apply each override individually
			for _, mention := range ctx.Message.Mentions {
				overwrites = append(overwrites, &channel.PermissionOverwrite{
					Id: mention.Id,
					Type: channel.PermissionTypeMember,
					Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
					Deny: 0,
				})
			}
		} else {
			// If adding a role as an admin, apply overrides to role
			for _, role := range roles {
				overwrites = append(overwrites, &channel.PermissionOverwrite{
					Id:    role,
					Type:  channel.PermissionTypeRole,
					Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
					Deny: 0,
				})
			}
		}

		data := rest.ModifyChannelData{
			PermissionOverwrites: overwrites,
			Position: ch.Position,
		}

		if _, err = ctx.Shard.ModifyChannel(*channelId, data); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
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
