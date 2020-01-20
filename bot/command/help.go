package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strings"
)

type HelpCommand struct {
}

func (HelpCommand) Name() string {
	return "help"
}

func (HelpCommand) Description() string {
	return "Shows you a list of commands"
}

func (HelpCommand) Aliases() []string {
	return []string{"h"}
}

func (HelpCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (HelpCommand) Execute(ctx utils.CommandContext) {
	msg := ""
	for _, cmd := range Commands {
		msg += fmt.Sprintf("`%s%s` - %s\n", config.Conf.Bot.Prefix, cmd.Name(), cmd.Description())
	}
	msg = strings.Trim(msg, "\n")

	ch, err := ctx.Session.UserChannelCreate(ctx.User.ID); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if ch != nil {
		embed := utils.NewEmbed().
			SetColor(int(utils.Green)).
			SetTitle("Help").
			SetDescription(msg)

		if !ctx.IsPremium {
			embed.SetFooter("Powered by ticketsbot.net", utils.AvatarUrl)
		}

		// Explicitly ignore error to fix 403 (Cannot send messages to this user)
		_, _ = ctx.Session.ChannelMessageSendEmbed(ch.ID, embed.MessageEmbed); if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}
	}

	ctx.ReactWithCheck()
}

func (HelpCommand) Parent() interface{} {
	return nil
}

func (HelpCommand) Children() []Command {
	return make([]Command, 0)
}

func (HelpCommand) PremiumOnly() bool {
	return false
}

func (HelpCommand) AdminOnly() bool {
	return false
}

func (HelpCommand) HelperOnly() bool {
	return false
}
