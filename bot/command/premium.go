package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	uuid "github.com/satori/go.uuid"
	"strconv"
	"time"
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

func (PremiumCommand) Execute(ctx utils.CommandContext) {
	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if len(ctx.Args) == 0 {
		if ctx.IsPremium {
			expiryChan := make(chan int64)
			go database.GetExpiry(guildId, expiryChan)
			expiry := <-expiryChan // millis

			parsed := time.Unix(0, expiry * int64(time.Millisecond))
			formatted := parsed.UTC().String()

			ctx.SendEmbed(utils.Red, "Premium", fmt.Sprintf("This guild already has premium. It expires on %s", formatted))
		} else {
			ctx.SendEmbed(utils.Red, "Premium", utils.PREMIUM_MESSAGE)
		}
	} else {
		key := uuid.FromStringOrNil(ctx.Args[0])

		if key == uuid.Nil {
			ctx.SendEmbed(utils.Red, "Premium", "Invalid key. Ensure that you have copied it correctly.")
			ctx.ReactWithCross()
			return
		}

		keyExistsChan := make(chan bool)
		go database.KeyExists(key, keyExistsChan)
		exists := <-keyExistsChan

		if !exists {
			ctx.SendEmbed(utils.Red, "Premium", "Invalid key. Ensure that you have copied it correctly.")
			ctx.ReactWithCross()
			return
		}

		lengthChan := make(chan int64)
		go database.PopKey(key, lengthChan)
		length := <-lengthChan

		userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		go database.AddPremium(key.String(), guildId, userId, length, userId)
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

func (PremiumCommand) AdminOnly() bool {
	return false
}

func (PremiumCommand) HelperOnly() bool {
	return false
}
