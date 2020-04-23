package main

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot"
	"github.com/TicketsBot/TicketsGo/bot/archive"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/TicketsBot/archiverclient"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6080", nil))
	}()

	config.Load()

	archive.ArchiverClient = archiverclient.NewArchiverClient(config.Conf.Bot.ObjectStore)

	sentry.Connect()

	database.Connect()
	database.Setup()
	modmaildatabase.Setup()

	cache.Client = cache.NewRedisClient()

	if config.Conf.Metrics.Statsd.Enabled {
		var err error
		statsd.Client, err = statsd.NewClient(); if err != nil {
			sentry.Error(err)
		}
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	bot.Start(ch)
}
