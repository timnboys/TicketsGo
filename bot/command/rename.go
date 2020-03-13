package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strings"
)

type RenameCommand struct {
}

func (RenameCommand) Name() string {
	return "rename"
}

func (RenameCommand) Description() string {
	return "Renames the current ticket"
}

func (RenameCommand) Aliases() []string {
	return []string{}
}

func (RenameCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (RenameCommand) Execute(ctx utils.CommandContext) {
	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(ctx.ChannelId, ticketChan)
	ticket := <-ticketChan

	// Check this is a ticket channel
	if ticket.Uuid == "" {
		ctx.SendEmbed(utils.Red, "Rename", "This command can only be ran in ticket channels")
		return
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Rename", "You need to specify a new name for this ticket")
		return
	}

	name := strings.Join(ctx.Args, " ")
	if _, err := ctx.Session.ChannelEdit(ctx.Channel, name); err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext()) // Probably 403
		return
	}

	ctx.SendEmbed(utils.Green, "Rename", fmt.Sprintf("This ticket has been renamed to <#%s>", ctx.Channel))
}

func (RenameCommand) Parent() interface{} {
	return nil
}

func (RenameCommand) Children() []Command {
	return make([]Command, 0)
}

func (RenameCommand) PremiumOnly() bool {
	return false
}

func (RenameCommand) AdminOnly() bool {
	return false
}

func (RenameCommand) HelperOnly() bool {
	return false
}
