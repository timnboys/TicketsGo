package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
	"strings"
)

type CannedResponseCommand struct {
}

func (CannedResponseCommand) Name() string {
	return "c"
}

func (CannedResponseCommand) Description() string {
	return "Sends a predefined canned response"
}

func (CannedResponseCommand) Aliases() []string {
	return []string{"canned", "cannedresponse", "cr", "tags", "tag"}
}

func (CannedResponseCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (CannedResponseCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must provide the ID of the canned response. For more help with canned responses, visit <https://ticketsbot.net#canned>.")
		ctx.ReactWithCross()
		return
	}

	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	channelId, err := strconv.ParseInt(ctx.Channel, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	id := strings.ToLower(ctx.Args[0])

	contentChan := make(chan string)
	go database.GetCannedResponse(guildId, id, contentChan)
	content := <-contentChan

	if content == "" {
		ctx.SendEmbed(utils.Red, "Error", "Invalid canned response. For more help with canned responses, visit <https://ticketsbot.net#canned>.")
		ctx.ReactWithCross()
		return
	}

	isTicket := make(chan bool)
	go database.IsTicketChannel(channelId, isTicket)
	if <-isTicket {
		ticketOwnerChan := make(chan int64)
		go database.GetOwnerByChannel(channelId, ticketOwnerChan)
		mention := fmt.Sprintf("<@%s>", strconv.Itoa(int(<-ticketOwnerChan)))
		content = strings.Replace(content, "%user%", mention, -1)
	}

	ctx.ReactWithCheck()
	if _, err = ctx.Session.ChannelMessageSend(ctx.Channel, content); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}
}

func (CannedResponseCommand) Parent() interface{} {
	return nil
}

func (CannedResponseCommand) Children() []Command {
	return make([]Command, 0)
}

func (CannedResponseCommand) PremiumOnly() bool {
	return false
}

func (CannedResponseCommand) AdminOnly() bool {
	return false
}

func (CannedResponseCommand) HelperOnly() bool {
	return false
}
