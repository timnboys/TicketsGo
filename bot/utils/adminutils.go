package utils

import "github.com/TicketsBot/TicketsGo/config"

func IsBotAdmin(id uint64) bool {
	for _, admin := range config.Conf.Bot.Admins {
		if admin == id {
			return true
		}
	}

	return false
}

func IsBotHelper(id uint64) bool {
	for _, helper := range config.Conf.Bot.Helpers {
		if helper == id {
			return true
		}
	}

	return false
}
