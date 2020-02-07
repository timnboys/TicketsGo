package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
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
	embed := utils.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Staff")

	var fieldContent string // temp var

	// Add field for admin users
	adminUsers := make(chan []int64)
	go database.GetAdmins(ctx.Guild.ID, adminUsers)
	for _, adminUserId := range <-adminUsers {
		fieldContent += fmt.Sprintf("• <@%d> (`%d`)\n", adminUserId, adminUserId)
		fieldContent = strings.TrimSuffix(fieldContent, "\n")
	}
	if fieldContent == "" {
		fieldContent = "No admin users"
	}
	embed.AddField("Admin Users", fieldContent, true)
	fieldContent = ""

	// Add field for admin roles
	adminRoles := make(chan []int64)
	go database.GetAdminRoles(ctx.Guild.ID, adminRoles)
	for _, adminRoleId := range <-adminRoles {
		// Resolve role ID to name
		role, err := ctx.Session.State.Role(ctx.Guild.ID, strconv.Itoa(int(adminRoleId))); if err != nil {
			role, err = ctx.Session.State.Role(ctx.Guild.ID, strconv.Itoa(int(adminRoleId))); if err != nil {
				// Role has likely been deleted
				continue
			}
		}

		fieldContent += fmt.Sprintf("• %s (`%d`)\n", role.Name, adminRoleId)
		fieldContent = strings.TrimSuffix(fieldContent, "\n")
	}
	if fieldContent == "" {
		fieldContent = "No admin roles"
	}
	embed.AddField("Admin Roles", fieldContent, true)
	fieldContent = ""

	embed.AddBlankField(false) // Add spacer between admin & support reps

	// Add field for support representatives
	supportUsers := make(chan []int64)
	go database.GetSupport(ctx.Guild.ID, supportUsers)
	for _, supportUserId := range <-supportUsers {
		fieldContent += fmt.Sprintf("• <@%d> (`%d`)\n", supportUserId, supportUserId)
		fieldContent = strings.TrimSuffix(fieldContent, "\n")
	}
	if fieldContent == "" {
		fieldContent = "No support representatives"
	}
	embed.AddField("Support Representatives", fieldContent, true)
	fieldContent = ""

	// Add field for admin roles
	supportRoles := make(chan []int64)
	go database.GetSupportRoles(ctx.Guild.ID, supportRoles)
	for _, supportRoleId := range <-supportRoles {
		// Resolve role ID to name
		role, err := ctx.Session.State.Role(ctx.Guild.ID, strconv.Itoa(int(supportRoleId))); if err != nil {
			role, err = ctx.Session.State.Role(ctx.Guild.ID, strconv.Itoa(int(supportRoleId))); if err != nil {
				// Role has likely been deleted
				continue
			}
		}

		fieldContent += fmt.Sprintf("• %s (`%d`)\n", role.Name, supportRoleId)
		fieldContent = strings.TrimSuffix(fieldContent, "\n")
	}
	if fieldContent == "" {
		fieldContent = "No support representative roles"
	}
	embed.AddField("Support Roles", fieldContent, true)
	fieldContent = ""

	msg, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed.MessageEmbed)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		utils.DeleteAfter(utils.SentMessage{Session: ctx.Session, Message: msg}, 60)
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

func (ViewStaffCommand) AdminOnly() bool {
	return false
}

func (ViewStaffCommand) HelperOnly() bool {
	return false
}
