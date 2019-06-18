package utils

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"strconv"
)

func IsPremiumGuild(guild string, ch chan bool) {
	i, err := strconv.ParseInt(guild, 10, 64); if err != nil {
		log.Error(err.Error())
		ch <- false
		return
	}

	database.IsPremium(i, ch)
}
