package listeners

import "github.com/TicketsBot/TicketsGo/config"

func isBotAdmin(id string) bool {
	for _, admin := range config.Conf.Bot.Admins {
		if admin == id {
			return true
		}
	}

	return false
}

func isBotHelper(id string) bool {
	for _, helper := range config.Conf.Bot.Helpers {
		if helper == id {
			return true
		}
	}

	return false
}
