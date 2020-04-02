package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"strings"
)

func OnCommand(s *gateway.Shard, e *events.MessageCreate) {
	if e.Author.Bot {
		return
	}

	// Ignore commands in DMs
	if e.GuildId == 0 {
		return
	}

	ch := make(chan string)
	go database.GetPrefix(e.GuildId, ch)

	customPrefix := <-ch
	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(e.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if strings.HasPrefix(e.Content, customPrefix) {
		usedPrefix = customPrefix
	} else { // Not a command
		return
	}

	split := strings.Split(e.Content, " ")
	root := strings.TrimPrefix(split[0], usedPrefix)

	args := make([]string, 0)
	if len(split) > 1 {
		for _, arg := range split[1:] {
			if arg != "" {
				args = append(args, arg)
			}
		}
	}

	var c command.Command
	for _, cmd := range command.Commands {
		if strings.ToLower(cmd.Name()) == strings.ToLower(root) || contains(cmd.Aliases(), strings.ToLower(root)) {
			parent := cmd
			index := 0

			for {
				if len(args) > index {
					childName := args[index]
					found := false

					for _, child := range parent.Children() {
						if strings.ToLower(child.Name()) == strings.ToLower(childName) || contains(child.Aliases(), strings.ToLower(childName)) {
							parent = child
							found = true
							index++
						}
					}

					if !found {
						break
					}
				} else {
					break
				}
			}

			var childArgs []string
			if len(args) > 0 {
				childArgs = args[index:]
			}

			args = childArgs
			c = parent
		}
	}

	errorContext := sentry.ErrorContext{
		Guild:   e.GuildId,
		User:    e.Author.Id,
		Channel: e.ChannelId,
		Shard:   s.ShardId,
		Command: root,
	}

	// Get guild obj
	guild, err := s.GetGuild(e.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	premiumChan := make(chan bool)
	go utils.IsPremiumGuild(utils.CommandContext{
		Shard: s,
		Guild: guild,
	}, premiumChan)
	premiumGuild := <-premiumChan

	e.Member.User = e.Author

	ctx := utils.CommandContext{
		Shard:       s,
		User:        e.Author,
		Guild:       guild,
		ChannelId:   e.ChannelId,
		Message:     e.Message,
		Root:        root,
		Args:        args,
		IsPremium:   premiumGuild,
		ShouldReact: true,
		Member:      e.Member,
		IsFromPanel: false,
	}

	// Ensure user isn't blacklisted
	blacklisted := make(chan bool)
	go database.IsBlacklisted(ctx.Guild.Id, ctx.User.Id, blacklisted)
	if <-blacklisted {
		ctx.ReactWithCross()
		return
	}

	if c != nil {
		permLevel := make(chan utils.PermissionLevel)
		go ctx.GetPermissionLevel(permLevel)
		if int(c.PermissionLevel()) > int(<-permLevel) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.AdminOnly() && !utils.IsBotAdmin(e.Author.Id) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", "This command is reserved for the bot owner only")
			return
		}

		if c.HelperOnly() && !utils.IsBotHelper(e.Author.Id) && !utils.IsBotAdmin(e.Author.Id) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.PremiumOnly() {
			if !premiumGuild {
				ctx.ReactWithCross()
				ctx.SendEmbed(utils.Red, "Premium Only Command", utils.PREMIUM_MESSAGE)
				return
			}
		}

		go c.Execute(ctx)
		go statsd.IncrementKey(statsd.COMMANDS)

		utils.DeleteAfter(utils.SentMessage{Shard: s, Message: e.Message}, 30)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
