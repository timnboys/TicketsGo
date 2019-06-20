package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
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

func (StatsServerCommand) Execute(ctx CommandContext) {
	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	totalTickets := make(chan int)
	go database.GetTotalTicketCount(guildId, totalTickets)

	openTickets := make(chan []string)
	go database.GetOpenTickets(guildId, openTickets)

	responseTimesChan := make(chan map[string]int64)
	go database.GetGuildResponseTimes(guildId, responseTimesChan)
	responseTimes := <-responseTimesChan

	// total average response
	var averageResponse int64
	for _, t := range responseTimes {
		averageResponse += t
	}
	averageResponse = averageResponse / int64(len(responseTimes))

	current := time.Now().UnixNano() / int64(time.Millisecond)

	// monthly average response

	// weekly average response
	weekly := make([]int64, 0)
	for uuid, t := range responseTimes {
		openTimeChan := make(chan *int64)
		go database.GetOpenTime(uuid, openTimeChan)
		openTime := <-openTimeChan
	}

	embed := utils.NewEmbed().
		SetTitle("Statistics").
		SetColor(int(utils.Green)).

		AddField("Total Tickets", strconv.Itoa(<-totalTickets), true).
		AddField("Open Tickets", strconv.Itoa(len(<-openTickets)), true).

		AddBlankField(false).

		AddField("Average First Response Time (Total)", utils.FormatTime(averageResponse), true)

		MessageEmbed

	if m, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed); err == nil {
		utils.DeleteAfter(utils.SentMessage{Session: ctx.Session, Message: m}, 60)
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
