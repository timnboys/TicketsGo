package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
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

func (ManageCannedResponsesDelete) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "You must specify a canned response ID to delete")
		return
	}

	id := ctx.Args[0]

	idsChan := make(chan []string)
	go database.GetCannedResponses(ctx.GuildId, idsChan)
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

	go database.DeleteCannedResponse(ctx.GuildId, id)
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

func (ManageCannedResponsesDelete) Category() Category {
	return CannedResponses
}

func (ManageCannedResponsesDelete) AdminOnly() bool {
	return false
}

func (ManageCannedResponsesDelete) HelperOnly() bool {
	return false
}
