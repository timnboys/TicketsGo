package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
)

type ClaimCommand struct {
}

func (ClaimCommand) Name() string {
	return "claim"
}

func (ClaimCommand) Description() string {
	return "Assigns a single staff member to a ticket"
}

func (ClaimCommand) Aliases() []string {
	return []string{}
}

func (ClaimCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (c ClaimCommand) Execute(ctx utils.CommandContext) {
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

	// Set to claimed in DB
	if err := database.Client.TicketClaims.Set(ctx.GuildId, ticket.Id, ctx.Author.Id); err != nil {
		ctx.HandleError(err)
		return
	}

	// Get claim settings for guild
	claimSettings, err := database.Client.ClaimSettings.Get(ctx.GuildId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Get support users
	supportUsers, err := database.Client.Permissions.GetSupportOnly(ctx.GuildId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Get support roles
	supportRoles, err := database.Client.RolePermissions.GetSupportRoles(ctx.GuildId); if err != nil {
		ctx.HandleError(err)
		return
	}

	// Get existing overwrites
	var overwrites []channel.PermissionOverwrite
	{
		channel, err := ctx.Shard.GetChannel(ctx.ChannelId); if err != nil {
			ctx.HandleError(err)
			return
		}

		overwrites = channel.PermissionOverwrites
	}

	// TODO: Just delete from original slice
	var newOverwrites []channel.PermissionOverwrite
	if !claimSettings.SupportCanView {
		newOverwrites = c.overwritesCantView(overwrites, supportUsers, supportRoles)
	} else if !claimSettings.SupportCanType {
		newOverwrites = c.overwritesCantType(overwrites, supportRoles, supportRoles)
	}

	// Update channel
	data := rest.ModifyChannelData{
		PermissionOverwrites: newOverwrites,
	}
	if _, err := ctx.Shard.ModifyChannel(ctx.ChannelId, data); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Ticket Claimed", fmt.Sprintf("Your ticket will be handled by %s", ctx.Author.Mention()))
	ctx.ReactWithCheck()
}

func (ClaimCommand) overwritesCantView(existingOverwrites []channel.PermissionOverwrite, supportUsers, supportRoles []uint64) (newOverwrites []channel.PermissionOverwrite) {
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

func (ClaimCommand) overwritesCantType(existingOverwrites []channel.PermissionOverwrite, supportUsers, supportRoles []uint64) (newOverwrites []channel.PermissionOverwrite) {
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

func (ClaimCommand) Parent() interface{} {
	return nil
}

func (ClaimCommand) Children() []Command {
	return make([]Command, 0)
}

func (ClaimCommand) PremiumOnly() bool {
	return false
}

func (ClaimCommand) Category() Category {
	return Tickets
}

func (ClaimCommand) AdminOnly() bool {
	return false
}

func (ClaimCommand) HelperOnly() bool {
	return false
}
