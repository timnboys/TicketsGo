package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/elliotchance/orderedmap"
	"github.com/rxdn/gdl/objects/channel/embed"
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
	commandCategories := orderedmap.NewOrderedMap()

	// initialise map with the correct order of categories
	for _, category := range categories {
		commandCategories.Set(category, make([]Command, 0))
	}

	for _, command := range Commands {
		// check bot admin / helper only commands
		if (command.AdminOnly() && !utils.IsBotAdmin(ctx.Author.Id)) || (command.HelperOnly() && !utils.IsBotHelper(ctx.Author.Id)) {
			continue
		}

		permissionLevel := make(chan utils.PermissionLevel)
		go utils.GetPermissionLevel(ctx.Shard, ctx.Member, ctx.GuildId, permissionLevel)
		if <-permissionLevel >= command.PermissionLevel() { // only send commands the user has permissions for
			var current []Command
			if commands, ok := commandCategories.Get(command.Category()); ok {
				current = commands.([]Command)
			}
			current = append(current, command)

			commandCategories.Set(command.Category(), current)
		}
	}

	// get prefix
	prefixChan := make(chan string)
	go getPrefix(ctx.GuildId, prefixChan)
	prefix := <-prefixChan

	embed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Help")

	for _, category := range commandCategories.Keys() {
		var commands []Command
		if retrieved, ok := commandCategories.Get(category.(Category)); ok {
			commands = retrieved.([]Command)
		}

		if len(commands) > 0 {
			formatted := make([]string, 0)
			for _, command := range commands {
				formatted = append(formatted, formatHelp(command, prefix))
			}

			embed.AddField(string(category.(Category)), strings.Join(formatted, "\n"), false)
		}
	}

	dmChannel, err := ctx.Shard.CreateDM(ctx.Author.Id); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	if !ctx.IsPremium {
		embed.SetFooter("Powered by ticketsbot.net", ctx.Shard.SelfAvatar(256))
	}

	// Explicitly ignore error to fix 403 (Cannot send messages to this user)
	_, err = ctx.Shard.CreateMessageEmbed(dmChannel.Id, embed)
	if err == nil {
		ctx.ReactWithCheck()
	} else {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I couldn't send you a direct message: make sure your privacy settings aren't too high")
	}
}

func formatHelp(c Command, prefix string) string {
	return fmt.Sprintf("**%s%s**: %s", prefix, c.Name(), c.Description())
}

func getPrefix(guildId uint64, res chan string) {
	ch := make(chan string)
	go database.GetPrefix(guildId, ch)
	customPrefix := <-ch

	if customPrefix != "" {
		res <- customPrefix
	} else { // return default prefix
		res <- config.Conf.Bot.Prefix
	}
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

func (HelpCommand) Category() Category {
	return General
}

func (HelpCommand) AdminOnly() bool {
	return false
}

func (HelpCommand) HelperOnly() bool {
	return false
}
