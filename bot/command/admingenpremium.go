package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	uuid "github.com/satori/go.uuid"
	"strconv"
	"strings"
)

type AdminGeneratePremium struct {
}

func (AdminGeneratePremium) Name() string {
	return "genpremium"
}

func (AdminGeneratePremium) Description() string {
	return "Generate premium keys"
}

func (AdminGeneratePremium) Aliases() []string {
	return []string{"gp", "gk", "generatepremium", "genkeys", "generatekeys"}
}

func (AdminGeneratePremium) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminGeneratePremium) Execute(ctx utils.CommandContext) {
	if len(ctx.Args) == 0 {
		ctx.ReactWithCross()
		return
	}

	days, err := strconv.Atoi(ctx.Args[0]); if err != nil {
		ctx.SendEmbed(utils.Red, "Admin", err.Error())
		ctx.ReactWithCross()
		return
	}
	millis := int64(days) * 24 * 60 * 60 * 1000

	amount := 1
	if len(ctx.Args) == 2 {
		if a, err := strconv.Atoi(ctx.Args[1]); err == nil {
			amount = a
		}
	}

	keys := make([]string, 0)
	for i := 0; i < amount; i++ {
		key := make(chan uuid.UUID)
		go database.AddKey(millis, key)
		keys = append(keys, (<-key).String())
	}

	ch, err := ctx.Shard.UserChannelCreate(ctx.User.ID); if err != nil {
		ctx.SendEmbed(utils.Red, "Admin", err.Error())
		ctx.ReactWithCross()
		return
	}

	content := "```"
	for _, key := range keys {
		content += fmt.Sprintf("%s\n", key)
	}
	content = strings.TrimSuffix(content, "\n")
	content += "```"

	_, err = ctx.Shard.ChannelMessageSend(ch.ID, content); if err != nil {
		ctx.SendEmbed(utils.Red, "Admin", err.Error())
		ctx.ReactWithCross()
		return
	}

	ctx.ReactWithCheck()
}

func (AdminGeneratePremium) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminGeneratePremium) Children() []Command {
	return []Command{}
}

func (AdminGeneratePremium) PremiumOnly() bool {
	return false
}

func (AdminGeneratePremium) AdminOnly() bool {
	return true
}

func (AdminGeneratePremium) HelperOnly() bool {
	return false
}
