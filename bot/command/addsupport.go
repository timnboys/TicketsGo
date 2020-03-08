package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant support representative privileges to")
		ctx.ReactWithCross()
		return
	}

	user := false
	roles := make([]string, 0)

	if len(ctx.Message.Mentions) > 0 {
		user = true
		for _, mention := range ctx.Message.Mentions {
			go database.AddSupport(ctx.Guild.ID, mention.ID)
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
				roles = append(roles, role.ID)
				break
			}
		}

		// Verify a valid role was mentioned
		if !valid {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant support representative privileges to")
			ctx.ReactWithCross()
			return
		}
	}

	// Add roles to DB
	for _, role := range roles {
		go database.AddSupportRole(ctx.Guild.ID, role)
	}

	openTicketsChan := make(chan []*int64)
	go database.GetOpenTicketChannelIds(ctx.GuildId, openTicketsChan)

	// Update permissions for existing tickets
	for _, channelId := range <-openTicketsChan {
		// Mitigation for a very rare panic that occurs when this command is run whilst a ticket is being opened, but
		// the channel ID hasn't been set in the database yet, or if Discord is dead and won't let the channel
		// be created.
		if channelId == nil {
			continue
		}

		var overwrites []*discordgo.PermissionOverwrite
		ch, err := ctx.Session.Channel(strconv.Itoa(int(*channelId))); if err != nil {
			continue
		}

		overwrites = ch.PermissionOverwrites

		if user {
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
			for _, role := range roles {
				overwrites = append(overwrites, &discordgo.PermissionOverwrite{
					ID:    role,
					Type:  "role",
					Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
					Deny: 0,
				})
			}
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
