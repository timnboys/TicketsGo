package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"strconv"
)

type ManageCannedResponsesDelete struct {
}

func (ManageCannedResponsesDelete) Name() string {
	return "delete"
}

func (ManageCannedResponsesDelete) Description() string {
	return "Deletes a canned response"
}

func (ManageCannedResponsesDelete) Aliases() []string {
	return []string{"del", "rm", "remove"}
}

func (ManageCannedResponsesDelete) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (ManageCannedResponsesDelete) Execute(ctx CommandContext) {
	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	if len(ctx.Args) == 0 {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You must specify a canned response ID to delete")
		return
	}

	id := ctx.Args[0]

	idsChan := make(chan []string)
	go database.GetCannedResponses(guildId, idsChan)
	ids := <-idsChan

	found := false
	for _, i := range ids {
		if i == id {
			found = true
			break
		}
	}

	if !found {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("A canned response with the ID `%s` could not be found", id))
		return
	}

	go database.DeleteCannedResponse(guildId, id)
	ctx.ReactWithCheck()
}

func (ManageCannedResponsesDelete) Parent() interface{} {
	return ManageCannedResponses{}
}

func (ManageCannedResponsesDelete) Children() []Command {
	return make([]Command, 0)
}

func (ManageCannedResponsesDelete) PremiumOnly() bool {
	return false
}

func (ManageCannedResponsesDelete) AdminOnly() bool {
	return false
}

func (ManageCannedResponsesDelete) HelperOnly() bool {
	return false
}
