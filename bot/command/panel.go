package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
)

type PanelCommand struct {
}

func (PanelCommand) Name() string {
	return "panel"
}

func (PanelCommand) Description() string {
	return "Creates a panel to enable users to open a ticket with 1 click"
}

func (PanelCommand) Aliases() []string {
	return []string{}
}

func (PanelCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (PanelCommand) Execute(ctx CommandContext) {
	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	settingsChan := make(chan database.PanelSettings)
	go database.GetPanelSettings(guildId, settingsChan)
	settings := <-settingsChan

	embed := utils.NewEmbed().
		SetColor(settings.Colour).
		SetTitle(settings.Title).
		SetDescription(settings.Content)

	if !ctx.IsPremium {
		embed.SetFooter("Powered by ticketsbot.net", utils.AvatarUrl)
	}

	msg, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed.MessageEmbed); if err != nil {
		sentry.Error(err)
		return
	}

	if err = ctx.Session.MessageReactionAdd(ctx.Channel, msg.ID, "📩"); err != nil {
		ctx.SendEmbed(utils.Red, "Error", "I don't have permission to react to the panel.")
		return
	}

	msgId, err := strconv.ParseInt(msg.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	go database.AddPanel(msgId, guildId)

	ctx.ReactWithCheck()
}

func (PanelCommand) Parent() interface{} {
	return nil
}

func (PanelCommand) Children() []Command {
	return make([]Command, 0)
}

func (PanelCommand) PremiumOnly() bool {
	return false
}

func (PanelCommand) AdminOnly() bool {
	return false
}

func (PanelCommand) HelperOnly() bool {
	return false
}

