package command

import (
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
)

type UnclaimCommand struct {
}

func (UnclaimCommand) Name() string {
	return "unclaim"
}

func (UnclaimCommand) Description() string {
	return "Removes the claim on the current ticket"
}

func (UnclaimCommand) Aliases() []string {
	return []string{}
}

func (UnclaimCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (c UnclaimCommand) Execute(ctx utils.CommandContext) {
	// Get ticket struct
	ticket, err := database.Client.Tickets.GetByChannel(ctx.ChannelId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Verify this is a ticket channel
	if ticket.UserId == 0 {
		ctx.SendEmbed(utils.Red, "Error", "This is not a ticket channel")
		ctx.ReactWithCross()
		return
	}

	// Get who claimed
	whoClaimed, err := database.Client.TicketClaims.Get(ctx.GuildId, ticket.Id); if err != nil {
		ctx.HandleError(err)
		return
	}

	if whoClaimed == 0 {
		ctx.SendEmbed(utils.Red, "Error", "This ticket is not claimed")
		ctx.ReactWithCross()
		return
	}

	var permissionLevel utils.PermissionLevel
	{
		ch := make(chan utils.PermissionLevel)
		go ctx.GetPermissionLevel(ch)
		permissionLevel = <-ch
	}

	if permissionLevel < utils.Admin && ctx.Author.Id != whoClaimed {
		ctx.SendEmbed(utils.Red, "Error", "Only admins and the user who claimed the ticket can unclaim the ticket")
		ctx.ReactWithCross()
		return
	}

	// Set to unclaimed in DB
	if err := database.Client.TicketClaims.Delete(ctx.GuildId, ticket.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: logic.CreateOverwrites(ctx.GuildId, ticket.UserId, ctx.Shard.SelfId()),
	}
	if _, err := ctx.Shard.ModifyChannel(ctx.ChannelId, data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Ticket Unclaimed", "All support representatives can now respond to the ticket")
	ctx.ReactWithCheck()
}

func (UnclaimCommand) overwritesCantView(existingOverwrites []channel.PermissionOverwrite, supportUsers, supportRoles []uint64) (newOverwrites []channel.PermissionOverwrite) {
	outer:
	for _, overwrite := range existingOverwrites {
		// Remove members
		if overwrite.Type == channel.PermissionTypeMember {
			for _, userId := range supportUsers {
				if overwrite.Id == userId {
					continue outer
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		} else if overwrite.Type == channel.PermissionTypeRole { // Remove roles
			for _, roleId := range supportRoles {
				if overwrite.Id == roleId {
					continue outer
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		}
	}

	return
}

func (UnclaimCommand) overwritesCantType(existingOverwrites []channel.PermissionOverwrite, supportUsers, supportRoles []uint64) (newOverwrites []channel.PermissionOverwrite) {
	for _, overwrite := range existingOverwrites {
		// Update members
		if overwrite.Type == channel.PermissionTypeMember {
			for _, userId := range supportUsers {
				if overwrite.Id == userId {
					overwrite.Allow = permission.BuildPermissions(permission.ViewChannel, permission.ReadMessageHistory)
					overwrite.Deny = permission.BuildPermissions(permission.AddReactions, permission.SendMessages)
					break
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		} else if overwrite.Type == channel.PermissionTypeRole { // Update roles
			for _, roleId := range supportRoles {
				if overwrite.Id == roleId {
					overwrite.Allow = permission.BuildPermissions(permission.ViewChannel, permission.ReadMessageHistory)
					overwrite.Deny = permission.BuildPermissions(permission.AddReactions, permission.SendMessages)
					break
				}
			}

			newOverwrites = append(newOverwrites, overwrite)
		}
	}

	return
}

func (UnclaimCommand) Parent() interface{} {
	return nil
}

func (UnclaimCommand) Children() []Command {
	return make([]Command, 0)
}

func (UnclaimCommand) PremiumOnly() bool {
	return false
}

func (UnclaimCommand) Category() Category {
	return Tickets
}

func (UnclaimCommand) AdminOnly() bool {
	return false
}

func (UnclaimCommand) HelperOnly() bool {
	return false
}
