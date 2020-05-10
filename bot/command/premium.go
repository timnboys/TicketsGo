package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/gofrs/uuid"
)

type PremiumCommand struct {
}

func (PremiumCommand) Name() string {
	return "premium"
}

func (PremiumCommand) Description() string {
	return "Used to register a premium key to your guild"
}

func (PremiumCommand) Aliases() []string {
	return []string{}
}

func (PremiumCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

// TODO: Retire this?
func (PremiumCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		if ctx.IsPremium {
			expiry, err := database.Client.PremiumGuilds.GetExpiry(ctx.GuildId)
			if err != nil {
				ctx.ReactWithCross()
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
				return
			}

			ctx.SendEmbed(utils.Red, "Premium", fmt.Sprintf("This guild already has premium. It expires on %s", expiry.UTC().String()))
		} else {
			ctx.SendEmbed(utils.Red, "Premium", utils.PREMIUM_MESSAGE)
		}
	} else {
		key, err := uuid.FromString(ctx.Args[0])

		if err != nil {
			ctx.SendEmbed(utils.Red, "Premium", "Invalid key. Ensure that you have copied it correctly.")
			ctx.ReactWithCross()
			return
		}

		length, err := database.Client.PremiumKeys.Delete(key)
		if err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		if length == 0 {
			ctx.SendEmbed(utils.Red, "Premium", "Invalid key. Ensure that you have copied it correctly.")
			ctx.ReactWithCross()
			return
		}

		if err := database.Client.UsedKeys.Set(key, ctx.GuildId, ctx.Author.Id); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		if err := database.Client.PremiumGuilds.Add(ctx.GuildId, length); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		go cache.Client.SetPremium(ctx.GuildId, true)
		ctx.ReactWithCheck()
	}
}

func (PremiumCommand) Parent() interface{} {
	return nil
}

func (PremiumCommand) Children() []Command {
	return make([]Command, 0)
}

func (PremiumCommand) PremiumOnly() bool {
	return false
}

func (PremiumCommand) Category() Category {
	return Settings
}

func (PremiumCommand) AdminOnly() bool {
	return false
}

func (PremiumCommand) HelperOnly() bool {
	return false
}
