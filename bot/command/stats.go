package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
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
	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	if len(ctx.Args) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must specify `server` to view server statistics, or tag a user to view their statistics")
		ctx.ReactWithCross()
		return
	}

	// server is handled as a subcommand, so a user has been pinged
	if len(ctx.Message.Mentions) == 0 {
		ctx.SendEmbed(utils.Red, "Error", "You must specify `server` to view server statistics, or tag a user to view their statistics")
		ctx.ReactWithCross()
		return
	}

	user := ctx.Message.Mentions[0]
	userId, err := strconv.ParseInt(user.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	// Get user permission level
	permLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(ctx.Session, ctx.Guild, user.ID, permLevelChan)
	permLevel := <-permLevelChan

	// User stats
	if permLevel == 0 {
		blacklisted := make(chan bool)
		go database.IsBlacklisted(guildId, userId, blacklisted)

		totalTickets := make(chan map[int64]int)
		go database.GetTicketsOpenedBy(guildId, userId, totalTickets)

		openTickets := make(chan []string)
		go database.GetOpenTicketsOpenedBy(guildId, userId, openTickets)

		ticketLimit := make(chan int)
		go database.GetTicketLimit(guildId, ticketLimit)

		embed := utils.NewEmbed().
			SetTitle("Statistics").
			SetColor(int(utils.Green)).

			AddField("Is Admin", "false", true).
			AddField("Is Support", "false", true).
			AddField("Is Blacklisted", strconv.FormatBool(<-blacklisted), true).

			AddField("Total Tickets", strconv.Itoa(len(<-totalTickets)), true).
			AddField("Open Tickets", fmt.Sprintf("%d / %d", len(<-openTickets), <-ticketLimit), true).

			MessageEmbed

		if m, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed); err == nil {
			utils.DeleteAfter(utils.SentMessage{Session: ctx.Session, Message: m}, 60)
		}
	} else { // Support rep stats
		responseTimesChan := make(chan map[string]int64)
		go database.GetUserResponseTimes(guildId, userId, responseTimesChan)
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

			AddField("Is Admin", strconv.FormatBool(permLevel == utils.Admin), true).
			AddField("Is Support", strconv.FormatBool(permLevel >= utils.Support), true).

			AddBlankField(false).

			AddField("Average First Response Time (Total)", utils.FormatTime(averageResponse), true).
			AddField("Average First Response Time (Weekly)", utils.FormatTime(weekly), true).
			AddField("Average First Response Time (Monthly)", utils.FormatTime(monthly), true).

			MessageEmbed

		if m, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed); err == nil {
			utils.DeleteAfter(utils.SentMessage{Session: ctx.Session, Message: m}, 60)
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

func (StatsCommand) AdminOnly() bool {
	return false
}

func (StatsCommand) HelperOnly() bool {
	return false
}
