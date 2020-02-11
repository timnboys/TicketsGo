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

func (PanelCommand) Execute(ctx utils.CommandContext) {
	// Check the panel quota
	if !ctx.IsPremium {
		panels := make(chan []database.Panel)
		go database.GetPanelsByGuild(ctx.GuildId, panels)
		if len(<-panels) > 1 {
			ctx.SendEmbed(utils.Red, "Error", "You have hit your panel quota. Delete a panel on the web UI (<https://panel.ticketsbot.net>, or purchase premium at <https://ticketsbot.net/premium> to create unlimited panels")
			return
		}
	}

	settingsChan := make(chan database.PanelSettings)
	go database.GetPanelSettings(ctx.GuildId, settingsChan)
	settings := <-settingsChan

	embed := utils.NewEmbed().
		SetColor(settings.Colour).
		SetTitle(settings.Title).
		SetDescription(settings.Content)

	if !ctx.IsPremium {
		embed.SetFooter("Powered by ticketsbot.net", utils.AvatarUrl)
	}

	msg, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed.MessageEmbed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if err = ctx.Session.MessageReactionAdd(ctx.Channel, msg.ID, "📩"); err != nil {
		ctx.SendEmbed(utils.Red, "Error", "I don't have permission to react to the panel.")
		return
	}

	msgId, err := strconv.ParseInt(msg.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	defaultCategory := make(chan int64)
	go database.GetCategory(ctx.GuildId, defaultCategory)

	go database.AddPanel(msgId, ctx.ChannelId, ctx.GuildId, settings.Title, settings.Content, settings.Colour, <-defaultCategory, "📩")

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

