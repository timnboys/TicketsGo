package command

import (
	"context"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"golang.org/x/sync/errgroup"
	"strconv"
	"time"
)

type StatsServerCommand struct {
}

func (StatsServerCommand) Name() string {
	return "server"
}

func (StatsServerCommand) Description() string {
	return "Shows you statistics about the server"
}

func (StatsServerCommand) Aliases() []string {
	return []string{"s"}
}

func (StatsServerCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Support
}

func (StatsServerCommand) Execute(ctx utils.CommandContext) {
	var totalTickets, openTickets int

	group, _ := errgroup.WithContext(context.Background())

	// totalTickets
	group.Go(func() (err error) {
		totalTickets, err = database.Client.Tickets.GetTotalTicketCount(ctx.GuildId)
		return
	})

	// openTickets
	group.Go(func() error {
		tickets, err := database.Client.Tickets.GetGuildOpenTickets(ctx.GuildId)
		openTickets = len(tickets)
		return err
	})

	// first response times
	var weekly, monthly, total *time.Duration

	// total
	group.Go(func() (err error) {
		total, err = database.Client.FirstResponseTime.GetAverageAllTime(ctx.GuildId)
		return
	})

	// monthly
	group.Go(func() (err error) {
		monthly, err = database.Client.FirstResponseTime.GetAverage(ctx.GuildId, time.Hour * 24 * 28)
		return
	})

	// weekly
	group.Go(func() (err error) {
		weekly, err = database.Client.FirstResponseTime.GetAverage(ctx.GuildId, time.Hour * 24 * 7)
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

		AddField("Total Tickets", strconv.Itoa(totalTickets), true).
		AddField("Open Tickets", strconv.Itoa(openTickets), true).

		AddBlankField(false).

		AddField("Average First Response Time (Total)", totalFormatted, true).
		AddField("Average First Response Time (Monthly)", monthlyFormatted, true).
		AddField("Average First Response Time (Weekly)", weeklyFormatted, true)

	if m, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed); err == nil {
		utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: &m}, 60)
	}
}

func (StatsServerCommand) Parent() interface{} {
	return StatsCommand{}
}

func (StatsServerCommand) Children() []Command {
	return make([]Command, 0)
}

func (StatsServerCommand) PremiumOnly() bool {
	return true
}

func (StatsServerCommand) Category() Category {
	return Statistics
}

func (StatsServerCommand) AdminOnly() bool {
	return false
}

func (StatsServerCommand) HelperOnly() bool {
	return false
}
