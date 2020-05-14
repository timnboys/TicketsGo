package command

import (
	"context"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/logic"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/cache"
	"strings"
)

type TransferCommand struct {
}

func (TransferCommand) Name() string {
	return "transfer"
}

func (TransferCommand) Description() string {
	return "Transfers a claimed ticket to another user"
}

func (TransferCommand) Aliases() []string {
	return []string{}
}

func (TransferCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (c TransferCommand) Execute(ctx utils.CommandContext) {
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

	target, found := c.getMentionedStaff(ctx)
	if !found {
		ctx.SendEmbed(utils.Red, "Error", "Couldn't find the target user")
		ctx.ReactWithCross()
		return
	}

	if err := logic.ClaimTicket(ctx.Shard, ticket, target); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbedNoDelete(utils.Green, "Ticket Claimed", fmt.Sprintf("Your ticket will be handled by <@%d>", target))
	ctx.ReactWithCheck()
}

func (TransferCommand) getMentionedStaff(ctx utils.CommandContext) (userId uint64, found bool) {
	if len(ctx.Mentions) > 0 {
		return ctx.Mentions[0].User.Id, true
	}

	if len(ctx.Args) == 0 {
		return
	}

	// get staff
	supportUsers, err := database.Client.Permissions.GetSupport(ctx.GuildId); if err != nil {
		return
	}

	supportRoles, err := database.Client.RolePermissions.GetSupportRoles(ctx.GuildId); if err != nil {
		return
	}

	query := `SELECT users.user_id FROM users WHERE LOWER("data"->>'Username') LIKE LOWER($1) AND EXISTS(SELECT FROM members WHERE members.guild_id=$2);`
	rows, err := ctx.Shard.Cache.(*cache.PgCache).Query(context.Background(), query, strings.Join(ctx.Args, " "), ctx.GuildId)
	defer rows.Close()
	if err != nil {
		return
	}

	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			continue
		}

		// Check if support rep
		for _, supportUser := range supportUsers {
			if supportUser == id {
				return id, true
			}
		}

		// Check if has support role
		// Get user object
		if member, err := ctx.Shard.GetGuildMember(ctx.GuildId, id); err == nil {
			for _, role := range member.Roles {
				for _, supportRole := range supportRoles {
					if role == supportRole {
						return id, true
					}
				}
			}
		}
	}

	return
}

func (TransferCommand) Parent() interface{} {
	return nil
}

func (TransferCommand) Children() []Command {
	return make([]Command, 0)
}

func (TransferCommand) PremiumOnly() bool {
	return false
}

func (TransferCommand) Category() Category {
	return Tickets
}

func (TransferCommand) AdminOnly() bool {
	return false
}

func (TransferCommand) HelperOnly() bool {
	return false
}
