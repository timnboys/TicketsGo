package main

import (
	"github.com/TicketsBot/TicketsGo/bot"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/apex/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	config.Load()

	sentry.Connect()
	log.SetHandler(sentry.Default)

	database.Connect()
	database.Setup()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	bot.Start(ch)
}
