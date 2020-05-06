package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/rxdn/gdl/rest"
)

type AdminSeedCommand struct {
}

func (AdminSeedCommand) Name() string {
	return "seed"
}

func (AdminSeedCommand) Description() string {
	return "Seeds the cache with members"
}

func (AdminSeedCommand) Aliases() []string {
	return []string{}
}

func (AdminSeedCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminSeedCommand) Execute(ctx utils.CommandContext) {
	var guilds []uint64
	guilds = []uint64{ctx.GuildId}

	ctx.SendEmbed(utils.Green, "Admin", fmt.Sprintf("Seeding %d guild(s)", len(guilds)))

	// retrieve all guild members
	var seeded int
	for _, guildId := range guilds {
		moreAvailable := true
		after := uint64(0)

		for moreAvailable {
			// calling this func will cache for us
			members, _ := ctx.Shard.ListGuildMembers(guildId, rest.ListGuildMembersData{
				Limit: 1000,
				After: after,
			})

			if len(members) < 1000 {
				moreAvailable = false
			}

			if len(members) > 0 {
				after = members[len(members) - 1].User.Id
			}
		}

		seeded++

		if seeded % 10 == 0 {
			ctx.SendEmbed(utils.Green, "Admin", fmt.Sprintf("Seeded %d / %d guilds", seeded, len(guilds)))
		}
	}

	ctx.SendEmbed(utils.Green, "Admin", "Seeding complete")
}

func (AdminSeedCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminSeedCommand) Children() []Command {
	return []Command{}
}

func (AdminSeedCommand) PremiumOnly() bool {
	return false
}

func (AdminSeedCommand) Category() Category {
	return Settings
}

func (AdminSeedCommand) AdminOnly() bool {
	return true
}

func (AdminSeedCommand) HelperOnly() bool {
	return false
}
