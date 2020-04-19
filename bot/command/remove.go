package command

import (
	"context"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/permission"
	"golang.org/x/sync/errgroup"
	"sync"
)

type RemoveCommand struct {
}

func (RemoveCommand) Name() string {
	return "remove"
}

func (RemoveCommand) Description() string {
	return "Removes a user from a ticket"
}

func (RemoveCommand) Aliases() []string {
	return []string{}
}

func (RemoveCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (r RemoveCommand) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!remove @User`",
		Inline: false,
	}

	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You need to mention members to remove from the ticket", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// Verify that the current channel is a real ticket
	isTicketChan := make(chan bool)
	go database.IsTicketChannel(ctx.ChannelId, isTicketChan)
	isTicket := <-isTicketChan

	if !isTicket {
		ctx.SendEmbed(utils.Red, "Error", "The current channel is not a ticket")
		ctx.ReactWithCross()
		return
	}

	// Get ticket ID
	ticketIdChan := make(chan int)
	go database.GetTicketId(ctx.ChannelId, ticketIdChan)
	ticketId := <-ticketIdChan

	// Verify that the user is allowed to modify the ticket
	permLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Shard, ctx.Member, ctx.GuildId, permLevelChan)
	permLevel := <-permLevelChan

	ownerChan := make(chan uint64)
	go database.GetOwner(ticketId, ctx.GuildId, ownerChan)
	owner := <-ownerChan

	if permLevel == 0 && owner != ctx.Author.Id {
		ctx.SendEmbed(utils.Red, "Error", "You don't have permission to remove people from this ticket")
		ctx.ReactWithCross()
		return
	}

	// verify that the user isn't trying to remove staff
	if r.mentionsStaff(ctx) {
		ctx.SendEmbed(utils.Red, "Error", "You cannot remove staff from a ticket")
		ctx.ReactWithCross()
		return
	}

	for _, user := range ctx.Message.Mentions {
		// Remove user from ticket in DB
		go database.RemoveMember(ticketId, ctx.GuildId, user.Id)

		// Remove user from ticket
		if err := ctx.Shard.EditChannelPermissions(ctx.ChannelId, channel.PermissionOverwrite{
			Id:    user.Id,
			Type:  channel.PermissionTypeMember,
			Allow: 0,
			Deny:  permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
		}); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		}
	}

	ctx.ReactWithCheck()
}

func (r RemoveCommand) mentionsStaff(ctx utils.CommandContext) bool {
	permissionChan := make(chan utils.PermissionLevel)

	group, _ := errgroup.WithContext(context.Background())
	for _, user := range ctx.Message.Mentions {
		group.Go(func() error {
			utils.GetPermissionLevel(ctx.Shard, user.Member, ctx.GuildId, permissionChan)
			return nil
		})
	}

	var lock sync.Mutex
	var mentionsStaff bool

	group.Go(func() error {
		for level := range permissionChan {
			if level > utils.Everyone {
				lock.Lock()
				mentionsStaff = true
				lock.Unlock()
			}
		}
		return nil
	})

	if err := group.Wait(); err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return true
	}

	return mentionsStaff
}

func (RemoveCommand) Parent() interface{} {
	return nil
}

func (RemoveCommand) Children() []Command {
	return make([]Command, 0)
}

func (RemoveCommand) PremiumOnly() bool {
	return false
}

func (RemoveCommand) Category() Category {
	return Tickets
}

func (RemoveCommand) AdminOnly() bool {
	return false
}

func (RemoveCommand) HelperOnly() bool {
	return false
}
