package command

import (
	"context"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/rxdn/gdl/cache"
)

// Reset
type AdminUsersCommand struct {
}

func (AdminUsersCommand) Name() string {
	return "users"
}

func (AdminUsersCommand) Description() string {
	return "Prints the instance's total user count"
}

func (AdminUsersCommand) Aliases() []string {
	return []string{}
}

func (AdminUsersCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminUsersCommand) Execute(ctx utils.CommandContext) {
	var count int
	query := `SELECT COUNT(DISTINCT "user_id") FROM members;`

	if err := ctx.Shard.Cache.(*cache.PgCache).QueryRow(context.Background(), query).Scan(&count); err != nil {
		ctx.HandleError(err)
		return
	}

	ctx.SendEmbed(utils.Green, "Admin", fmt.Sprintf("There are %d users on this instance", count))
}

func (AdminUsersCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminUsersCommand) Children() []Command {
	return []Command{}
}

func (AdminUsersCommand) PremiumOnly() bool {
	return false
}

func (AdminUsersCommand) Category() Category {
	return Settings
}

func (AdminUsersCommand) AdminOnly() bool {
	return true
}

func (AdminUsersCommand) HelperOnly() bool {
	return false
}
