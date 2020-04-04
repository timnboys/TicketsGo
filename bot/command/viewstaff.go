package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

type ViewStaffCommand struct {
}

func (ViewStaffCommand) Name() string {
	return "viewstaff"
}

func (ViewStaffCommand) Description() string {
	return "Lists the staff members and roles"
}

func (ViewStaffCommand) Aliases() []string {
	return []string{}
}

func (ViewStaffCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (ViewStaffCommand) Execute(ctx utils.CommandContext) {
	embed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Staff")

	var fieldContent string // temp var

	// Add field for admin users
	adminUsers := make(chan []uint64)
	go database.GetAdmins(ctx.Guild.Id, adminUsers)
	for _, adminUserId := range <-adminUsers {
		fieldContent += fmt.Sprintf("• <@%d> (`%d`)\n", adminUserId, adminUserId)
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No admin users"
	}
	embed.AddField("Admin Users", fieldContent, true)
	fieldContent = ""

	// get existing guild roles
	allRoles, err := ctx.Shard.GetGuildRoles(ctx.Guild.Id); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Add field for admin roles
	adminRoles := make(chan []uint64)
	go database.GetAdminRoles(ctx.Guild.Id, adminRoles)
	for _, adminRoleId := range <-adminRoles {
		for _, guildRole := range allRoles {
			if guildRole.Id == adminRoleId {
				fieldContent += fmt.Sprintf("• %s (`%d`)\n", guildRole.Name, adminRoleId)
			}
		}
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No admin roles"
	}
	embed.AddField("Admin Roles", fieldContent, true)
	fieldContent = ""

	embed.AddBlankField(false) // Add spacer between admin & support reps

	// Add field for support representatives
	supportUsers := make(chan []uint64)
	go database.GetSupport(ctx.Guild.Id, supportUsers)
	for _, supportUserId := range <-supportUsers {
		fieldContent += fmt.Sprintf("• <@%d> (`%d`)\n", supportUserId, supportUserId)
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No support representatives"
	}
	embed.AddField("Support Representatives", fieldContent, true)
	fieldContent = ""

	// Add field for admin roles
	supportRoles := make(chan []uint64)
	go database.GetSupportRoles(ctx.Guild.Id, supportRoles)
	for _, supportRoleId := range <-supportRoles {
		for _, guildRole := range allRoles {
			if guildRole.Id == supportRoleId {
				fieldContent += fmt.Sprintf("• %s (`%d`)\n", guildRole.Name, supportRoleId)
			}
		}
	}
	fieldContent = strings.TrimSuffix(fieldContent, "\n")
	if fieldContent == "" {
		fieldContent = "No support representative roles"
	}
	embed.AddField("Support Roles", fieldContent, true)
	fieldContent = ""

	msg, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: &msg}, 60)
	}
}

func (ViewStaffCommand) Parent() interface{} {
	return nil
}

func (ViewStaffCommand) Children() []Command {
	return make([]Command, 0)
}

func (ViewStaffCommand) PremiumOnly() bool {
	return false
}

func (ViewStaffCommand) Category() Category {
	return Settings
}

func (ViewStaffCommand) AdminOnly() bool {
	return false
}

func (ViewStaffCommand) HelperOnly() bool {
	return false
}
