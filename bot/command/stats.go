package command

import (
	"context"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type StatsCommand struct {
}

func (StatsCommand) Name() string {
	return "stats"
}

func (StatsCommand) Description() string {
	return "Shows you statistics about users, support staff and the server"
}

func (StatsCommand) Aliases() []string {
	return []string{"statistics"}
}

func (StatsCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (StatsCommand) Execute(ctx utils.CommandContext) {
	usageEmbed := embed.EmbedField{
		Name:   "Usage",
		Value:  "`t!stats server`\n`t!stats @User`",
		Inline: false,
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "Invalid argument: refer to usage", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	// server is handled as a subcommand, so a user has been pinged
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "Invalid argument: refer to usage", usageEmbed)
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]

	// User stats
	if ctx.UserPermissionLevel == 0 {
		var isBlacklisted bool
		var totalTickets int
		var openTickets int
		var ticketLimit uint8

		group, _ := errgroup.WithContext(context.Background())

		// load isBlacklisted
		group.Go(func() (err error) {
			isBlacklisted, err = database.Client.Blacklist.IsBlacklisted(ctx.GuildId, user.Id)
			return
		})

		// load totalTickets
		group.Go(func() error {
			tickets, err := database.Client.Tickets.GetAllByUser(ctx.GuildId, user.Id)
			totalTickets = len(tickets)
			return err
		})

		// load openTickets
		group.Go(func() error {
			tickets, err := database.Client.Tickets.GetOpenByUser(ctx.GuildId, user.Id)
			openTickets = len(tickets)
			return err
		})

		// load ticketLimit
		group.Go(func() (err error) {
			ticketLimit, err = database.Client.TicketLimit.Get(ctx.GuildId)
			return
		})

		if err := group.Wait(); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		embed := embed.NewEmbed().
			SetTitle("Statistics").
			SetColor(int(utils.Green)).

			AddField("Is Admin", "false", true).
			AddField("Is Support", "false", true).
			AddField("Is Blacklisted", strconv.FormatBool(isBlacklisted), true).

			AddField("Total Tickets", strconv.Itoa(totalTickets), true).
			AddField("Open Tickets", fmt.Sprintf("%d / %d", openTickets, ticketLimit), true)

		if m, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed); err == nil {
			utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: &m}, 60)
		}
	} else { // Support rep stats
		var weekly, monthly, total *time.Duration

		group, _ := errgroup.WithContext(context.Background())

		// total
		group.Go(func() (err error) {
			total, err = database.Client.FirstResponseTime.GetAverageAllTimeUser(ctx.GuildId, user.Id)
			return
		})

		// monthly
		group.Go(func() (err error) {
			monthly, err = database.Client.FirstResponseTime.GetAverageUser(ctx.GuildId, user.Id, time.Hour * 24 * 28)
			return
		})

		// weekly
		group.Go(func() (err error) {
			weekly, err = database.Client.FirstResponseTime.GetAverageUser(ctx.GuildId, user.Id, time.Hour * 24 * 7)
			return
		})

		if err := group.Wait(); err != nil {
			ctx.ReactWithCross()
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			return
		}

		var totalFormatted, monthlyFormatted, weeklyFormatted string

		if total == nil {
			totalFormatted = "No data"
		} else {
			totalFormatted = utils.FormatTime(*total)
		}

		if monthly == nil {
			monthlyFormatted = "No data"
		} else {
			monthlyFormatted = utils.FormatTime(*monthly)
		}

		if weekly == nil {
			weeklyFormatted = "No data"
		} else {
			weeklyFormatted = utils.FormatTime(*weekly)
		}

		embed := embed.NewEmbed().
			SetTitle("Statistics").
			SetColor(int(utils.Green)).

			AddField("Is Admin", strconv.FormatBool(permLevel == utils.Admin), true).
			AddField("Is Support", strconv.FormatBool(permLevel >= utils.Support), true).

			AddBlankField(false).

			AddField("Average First Response Time (Total)", totalFormatted, true).
			AddField("Average First Response Time (Monthly)", monthlyFormatted, true).
			AddField("Average First Response Time (Weekly)", weeklyFormatted, true)

		if m, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed); err == nil {
			utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: &m}, 60)
		}
	}
}

func (StatsCommand) Parent() interface{} {
	return nil
}

func (StatsCommand) Children() []Command {
	return []Command{
		StatsServerCommand{},
	}
}

func (StatsCommand) PremiumOnly() bool {
	return true
}

func (StatsCommand) Category() Category {
	return Statistics
}

func (StatsCommand) AdminOnly() bool {
	return false
}

func (StatsCommand) HelperOnly() bool {
	return false
}
