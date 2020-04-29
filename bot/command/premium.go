package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	uuid "github.com/satori/go.uuid"
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

// TODO: Retire this?
func (PremiumCommand) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		if ctx.IsPremium {
			expiryChan := make(chan int64)
			go database.GetExpiry(ctx.GuildId, expiryChan)
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

		go database.AddPremium(key.String(), ctx.GuildId, ctx.Author.Id, length, ctx.Author.Id)
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
