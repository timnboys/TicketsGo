package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
)

type AdminSeedMetrics struct {
}

func (AdminSeedMetrics) Name() string {
	return "seedmetrics"
}

func (AdminSeedMetrics) Description() string {
	return "Seeds statsd metrics"
}

func (AdminSeedMetrics) Aliases() []string {
	return []string{}
}

func (AdminSeedMetrics) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminSeedMetrics) Execute(ctx utils.CommandContext) {
	if statsd.IsClientNull() {
		ctx.SendEmbed(utils.Red, "Admin", "Statsd client is null")
		return
	}

	globalTicketsChan := make(chan int)
	go database.GetGlobalTicketCount(globalTicketsChan)
	globalTickets := <-globalTicketsChan

	go statsd.Client.Gauge(statsd.TICKETS.String(), globalTickets)

	ctx.SendEmbed(utils.Green, "Admin", "Seeded successfully")
}

func (AdminSeedMetrics) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminSeedMetrics) Children() []Command {
	return []Command{}
}

func (AdminSeedMetrics) PremiumOnly() bool {
	return false
}

func (AdminSeedMetrics) Category() Category {
	return Settings
}

func (AdminSeedMetrics) AdminOnly() bool {
	return true
}

func (AdminSeedMetrics) HelperOnly() bool {
	return false
}
