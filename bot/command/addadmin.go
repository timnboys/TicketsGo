package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		ctx.ReactWithCross()
		sentry.Error(err)
		return
	}

	openTicketsChan := make(chan []*int64)
	go database.GetOpenTicketChannelIds(guildId, openTicketsChan)

	for _, channelId := range <-openTicketsChan {
		var overwrites []*discordgo.PermissionOverwrite
		ch, err := ctx.Session.Channel(ctx.Channel); if err != nil {
			continue
		}

		overwrites = ch.PermissionOverwrites

		for _, mention := range ctx.Message.Mentions {
			overwrites = append(overwrites, &discordgo.PermissionOverwrite{
				ID: mention.ID,
				Type: "member",
				Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
				Deny: 0,
			})
		}

		data := discordgo.ChannelEdit{
			PermissionOverwrites: overwrites,
			Position: ch.Position,
		}

		_, _ = ctx.Session.ChannelEditComplex(strconv.Itoa(int(*channelId)), &data)
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
