package sentry

import(
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/sentry"
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
	"os"
	"time"
)

func Connect() {
	if err := raven.SetDSN(config.Conf.Sentry.DSN); err != nil {
		sentry.Error(err)
		return
	}
}

func ConstructPacket(e *errors.Error) *raven.Packet {
	hostname, err := os.Hostname(); if err != nil {
		hostname = "null"
		sentry.Error(err)
	}

	extra := map[string]interface{}{
		"stack": e.ErrorStack(),
	}

	return &raven.Packet{
		Message: e.Error(),
		Extra: extra,
		Project: "tickets-bot",
		Timestamp: raven.Timestamp(time.Now()),
		Level: raven.ERROR,
		ServerName: hostname,
	}
}
