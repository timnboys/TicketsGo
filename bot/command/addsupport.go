package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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
