package utils

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/sentry"
	"strconv"
)

func IsPremiumGuild(guild string, ch chan bool) {
	i, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		sentry.Error(err)
		ch <- false
		return
	}

	database.IsPremium(i, ch)
}
