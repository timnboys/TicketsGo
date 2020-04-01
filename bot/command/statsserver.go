package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
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
	guildId, err := strconv.ParseInt(ctx.Guild.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	totalTickets := make(chan int)
	go database.GetTotalTicketCount(guildId, totalTickets)

	openTickets := make(chan []string)
	go database.GetOpenTickets(guildId, openTickets)

	responseTimesChan := make(chan map[string]int64)
	go database.GetGuildResponseTimes(guildId, responseTimesChan)
	responseTimes := <-responseTimesChan

	openTimesChan := make(chan map[string]*int64)
	go database.GetOpenTimes(guildId, openTimesChan)
	openTimes := <-openTimesChan

	// total average response
	var averageResponse int64
	for _, t := range responseTimes {
		averageResponse += t
	}
	if len(responseTimes) > 0 { // Note: If responseTimes is empty, averageResponse must = 0
		averageResponse = averageResponse / int64(len(responseTimes))
	}

	current := time.Now().UnixNano() / int64(time.Millisecond)

	// monthly average response
	var monthly int64
	var monthlyCounter int
	for uuid, t := range responseTimes {
		openTime := openTimes[uuid]
		if openTime == nil {
			continue
		}

		if current - *openTime < 60 * 60 * 24 * 7 * 4 * 1000 {
			monthly += t
			monthlyCounter++
		}
	}
	if monthlyCounter > 0 {
		monthly = monthly / int64(monthlyCounter)
	}

	// weekly average response
	var weekly int64
	var weeklyCounter int
	for uuid, t := range responseTimes {
		openTime := openTimes[uuid]
		if openTime == nil {
			continue
		}

		if current - *openTime < 60 * 60 * 24 * 7 * 1000 {
			weekly += t
			weeklyCounter++
		}
	}
	if weeklyCounter > 0 {
		weekly = weekly / int64(weeklyCounter)
	}

	embed := utils.NewEmbed().
		SetTitle("Statistics").
		SetColor(int(utils.Green)).

		AddField("Total Tickets", strconv.Itoa(<-totalTickets), true).
		AddField("Open Tickets", strconv.Itoa(len(<-openTickets)), true).

		AddBlankField(false).

		AddField("Average First Response Time (Total)", utils.FormatTime(averageResponse), true).
		AddField("Average First Response Time (Weekly)", utils.FormatTime(weekly), true).
		AddField("Average First Response Time (Monthly)", utils.FormatTime(monthly), true).

		MessageEmbed

	if m, err := ctx.Shard.ChannelMessageSendEmbed(ctx.Channel, embed); err == nil {
		utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: m}, 60)
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

func (StatsServerCommand) AdminOnly() bool {
	return false
}

func (StatsServerCommand) HelperOnly() bool {
	return false
}
