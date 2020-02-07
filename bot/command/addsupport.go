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

	byMention := false
	var roleId string

	if len(ctx.Message.Mentions) > 0 {
		byMention = true
		for _, mention := range ctx.Message.Mentions {
			go database.AddSupport(ctx.Guild.ID, mention.ID)
		}
	} else {
		roleName := strings.ToLower(strings.Join(ctx.Args, " "))

		// Get role ID from name
		for _, role := range ctx.Guild.Roles {
			if strings.ToLower(role.Name) == roleName {
				roleId = role.ID
				break
			}
		}

		// Verify a valid role was mentioned
		if roleId == "" {
			ctx.SendEmbed(utils.Red, "Error", "You need to mention a user or name a role to grant support representative to")
			ctx.ReactWithCross()
			return
		}

		go database.AddSupportRole(ctx.Guild.ID, roleId)
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
			// If adding individual support representative, apply each override individually
			for _, mention := range ctx.Message.Mentions {
				overwrites = append(overwrites, &discordgo.PermissionOverwrite{
					ID: mention.ID,
					Type: "member",
					Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
					Deny: 0,
				})
			}
		} else {
			// If adding a role as an support representative, apply overrides to role
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
