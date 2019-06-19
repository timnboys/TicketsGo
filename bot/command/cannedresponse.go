package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
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

func (CannedResponseCommand) Execute(ctx CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must provide the ID of the canned response. For more help with canned responses, visit <https://ticketsbot.net#canned>.")
		ctx.ReactWithCross()
		return
	}

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
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

	ctx.ReactWithCheck()
	if _, err = ctx.Session.ChannelMessageSend(ctx.Channel, content); err != nil {
		log.Error(err.Error())
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
