package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/rest"
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
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!rename [ticket-name]`",
		Inline: false,
	}

	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(ctx.ChannelId, ticketChan)
	ticket := <-ticketChan

	// Check this is a ticket channel
	if ticket.Uuid == "" {
		ctx.SendEmbed(utils.Red, "Rename", "This command can only be ran in ticket channels", usageEmbed)
		return
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Rename", "You need to specify a new name for this ticket", usageEmbed)
		return
	}

	name := strings.Join(ctx.Args, " ")
	data := rest.ModifyChannelData{
		Name: name,
	}

	if _, err := ctx.Shard.ModifyChannel(ctx.ChannelId, data); err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext()) // Probably 403
		return
	}

	ctx.SendEmbed(utils.Green, "Rename", fmt.Sprintf("This ticket has been renamed to <#%d>", ctx.ChannelId))
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

func (RenameCommand) Category() Category {
	return Tickets
}

func (RenameCommand) AdminOnly() bool {
	return false
}

func (RenameCommand) HelperOnly() bool {
	return false
}
