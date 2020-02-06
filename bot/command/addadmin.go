package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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

	byMention := false
	var roleId string

	if len(ctx.Message.Mentions) > 0 {
		byMention = true
		for _, mention := range ctx.Message.Mentions {
			go database.AddAdmin(ctx.Guild.ID, mention.ID)
		}
	} else {
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
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant admin privileges to")
			ctx.ReactWithCross()
			return
		}

		go database.AddAdminRole(ctx.Guild.ID, roleId)
	}

	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		ctx.ReactWithCross()
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	openTicketsChan := make(chan []*int64)
	go database.GetOpenTicketChannelIds(guildId, openTicketsChan)

	// Update permissions for existing tickets
	for _, channelId := range <-openTicketsChan {
		var overwrites []*discordgo.PermissionOverwrite
		ch, err := ctx.Session.Channel(strconv.Itoa(int(*channelId))); if err != nil {
			continue
		}

		overwrites = ch.PermissionOverwrites

		if byMention {
			// If adding individual admins, apply each override individually
			for _, mention := range ctx.Message.Mentions {
				overwrites = append(overwrites, &discordgo.PermissionOverwrite{
					ID: mention.ID,
					Type: "member",
					Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
					Deny: 0,
				})
			}
		} else {
			// If adding a role as an admin, apply overrides to role
			overwrites = append(overwrites, &discordgo.PermissionOverwrite{
				ID:    roleId,
				Type:  "role",
				Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
				Deny: 0,
			})
		}

		data := discordgo.ChannelEdit{
			PermissionOverwrites: overwrites,
			Position: ch.Position,
		}

		if _, err = ctx.Session.ChannelEditComplex(strconv.Itoa(int(*channelId)), &data); err != nil {
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
