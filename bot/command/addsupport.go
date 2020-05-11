package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"strings"
)

type AddSupportCommand struct {
}

func (AddSupportCommand) Name() string {
	return "addsupport"
}

func (AddSupportCommand) Description() string {
	return "Adds a user or role as a support representative"
}

func (AddSupportCommand) Aliases() []string {
	return []string{}
}

func (AddSupportCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (AddSupportCommand) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!addadmin @User`\n`t!addadmin @Role`\n`t!addadmin role name`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant support representative privileges to", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	user := false
	roles := make([]uint64, 0)

	if len(ctx.Message.Mentions) > 0 {
		user = true
		for _, mention := range ctx.Message.Mentions {
			go func() {
				if err := database.Client.Permissions.AddSupport(ctx.GuildId, mention.Id); err != nil {
					sentry.ErrorWithContext(err, ctx.ToErrorContext())
				}
			}()
		}
	} else if len(ctx.Message.MentionRoles) > 0 {
		for _, mention := range ctx.Message.MentionRoles {
			roles = append(roles, mention)
		}
	} else {
		guildRoles, err := ctx.Shard.GetGuildRoles(ctx.GuildId); if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		roleName := strings.ToLower(strings.Join(ctx.Args, " "))

		// Get role ID from name
		valid := false
		for _, role := range guildRoles {
			if strings.ToLower(role.Name) == roleName {
				valid = true
				roles = append(roles, role.Id)
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant support representative privileges to", usageEmbed)
			ctx.ReactWithCross()
			return
		}
	}

	// Add roles to DB
	for _, role := range roles {
		if err := database.Client.RolePermissions.AddSupport(ctx.GuildId, role); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}

	openTickets, err := database.Client.Tickets.GetGuildOpenTickets(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Update permissions for existing tickets
	for _, ticket := range openTickets {
		if ticket.ChannelId == nil {
			continue
		}

		ch, err := ctx.Shard.GetChannel(*ticket.ChannelId); if err != nil {
			continue
		}

		overwrites := ch.PermissionOverwrites

		if user {
			// If adding individual admins, apply each override individually
			for _, mention := range ctx.Message.Mentions {
				overwrites = append(overwrites, channel.PermissionOverwrite{
					Id: mention.Id,
					Type: channel.PermissionTypeMember,
					Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
					Deny: 0,
				})
			}
		} else {
			// If adding a role as an admin, apply overrides to role
			for _, role := range roles {
				overwrites = append(overwrites, channel.PermissionOverwrite{
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

		if _, err = ctx.Shard.ModifyChannel(*ticket.ChannelId, data); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
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

func (AddSupportCommand) Category() Category {
	return Settings
}

func (AddSupportCommand) AdminOnly() bool {
	return false
}

func (AddSupportCommand) HelperOnly() bool {
	return false
}
