package listeners

import (
	"context"
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel/message"
	"golang.org/x/sync/errgroup"
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

	customPrefix, err := database.Client.Prefix.Get(e.GuildId)
	if err != nil {
		sentry.Error(err)
	}

	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(e.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if strings.HasPrefix(e.Content, customPrefix) && customPrefix != "" {
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

	var blacklisted, premium bool

	group, _ := errgroup.WithContext(context.Background())

	// get blacklisted
	group.Go(func() (err error) {
		blacklisted, err = database.Client.Blacklist.IsBlacklisted(e.GuildId, e.Author.Id)
		return
	})

	// get premium
	group.Go(func() error {
		ch := make(chan bool)
		go utils.IsPremiumGuild(s, e.GuildId, ch) // TODO: How long will this block for?
		premium = <-ch
		return nil
	})

	if err := group.Wait(); err != nil {
		sentry.Error(err)
		return
	}

	// Ensure user isn't blacklisted
	if blacklisted {
		utils.ReactWithCross(s, message.MessageReference{
			MessageId: e.Id,
			ChannelId: e.ChannelId,
			GuildId:   e.GuildId,
		})
		return
	}

	e.Member.User = e.Author

	ctx := utils.CommandContext{
		Shard:       s,
		Message:     e.Message,
		Root:        root,
		Args:        args,
		IsPremium:   premium,
		ShouldReact: true,
		IsFromPanel: false,
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

		if c.PremiumOnly() && !premium {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Premium Only Command", utils.PREMIUM_MESSAGE)
			return
		}

		go c.Execute(ctx)
		go statsd.IncrementKey(statsd.COMMANDS)

		utils.DeleteAfter(utils.SentMessage{Shard: s, Message: &e.Message}, 30)
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
