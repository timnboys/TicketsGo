package main

import (
	"context"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot"
	"github.com/TicketsBot/TicketsGo/bot/archive"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/TicketsBot/archiverclient"
	db "github.com/TicketsBot/database"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v4/log/logrusadapter"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
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

	initDatabase()

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

func initDatabase() {
	config, err := pgxpool.ParseConfig(config.Conf.Database.Uri); if err != nil {
		panic(err)
	}

	// TODO: Sentry
	config.ConnConfig.LogLevel = pgx.LogLevelWarn
	config.ConnConfig.Logger = logrusadapter.NewLogger(logrus.New())

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		panic(err)
	}

	database.Client = db.NewDatabase(pool)
	database.Client.CreateTables(pool)
}
