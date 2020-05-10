package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type TagCommand struct {
}

func (TagCommand) Name() string {
	return "tag"
}

func (TagCommand) Description() string {
	return "Sends a message snippet"
}

func (TagCommand) Aliases() []string {
	return []string{"canned", "cannedresponse", "cr", "tags", "tag", "snippet", "c"}
}

func (TagCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (TagCommand) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!tag [TagID]`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must provide the ID of the tag. For more help with tag, visit <https://ticketsbot.net/cannedresponses>.", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	tagId := strings.ToLower(ctx.Args[0])

	content, err := database.Client.Tag.Get(ctx.GuildId, tagId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		ctx.ReactWithCross()
		return
	}

	if content == "" {
		ctx.SendEmbed(utils.Red, "Error", "Invalid tag. For more help with tags, visit <https://ticketsbot.net/cannedresponses>.", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	ticket, err := database.Client.Tickets.GetByChannel(ctx.ChannelId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	if ticket.UserId != 0 {
		mention := fmt.Sprintf("<@%d>", ticket.UserId)
		content = strings.Replace(content, "%user%", mention, -1)
	}

	ctx.ReactWithCheck()
	if _, err := ctx.Shard.CreateMessage(ctx.ChannelId, content); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}

func (TagCommand) Parent() interface{} {
	return nil
}

func (TagCommand) Children() []Command {
	return make([]Command, 0)
}

func (TagCommand) PremiumOnly() bool {
	return false
}

func (TagCommand) Category() Category {
	return Tags
}

func (TagCommand) AdminOnly() bool {
	return false
}

func (TagCommand) HelperOnly() bool {
	return false
}
