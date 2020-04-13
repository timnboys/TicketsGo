package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strconv"
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

	id := strings.ToLower(ctx.Args[0])

	contentChan := make(chan string)
	go database.GetCannedResponse(ctx.GuildId, id, contentChan)
	content := <-contentChan

	if content == "" {
		ctx.SendEmbed(utils.Red, "Error", "Invalid tag. For more help with tags, visit <https://ticketsbot.net/cannedresponses>.", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	isTicket := make(chan bool)
	go database.IsTicketChannel(ctx.ChannelId, isTicket)
	if <-isTicket {
		ticketOwnerChan := make(chan uint64)
		go database.GetOwnerByChannel(ctx.ChannelId, ticketOwnerChan)
		mention := fmt.Sprintf("<@%s>", strconv.FormatUint(<-ticketOwnerChan, 10))
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
