package sentry

import(
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/apex/log"
	"github.com/getsentry/raven-go"
	"os"
	"time"
)

func Connect() {
	if err := raven.SetDSN(config.Conf.Sentry.DSN); err != nil {
		log.Error(err.Error())
		return
	}
}

func ConstructPacket(e *log.Entry) *raven.Packet {
	hostname, err := os.Hostname(); if err != nil {
		hostname = "null"
		log.Error(err.Error())
	}

	return &raven.Packet{
		Message: e.Message,
		Extra: map[string]interface{}(e.Fields),
		Project: "tickets-bot",
		Timestamp: raven.Timestamp(time.Now()),
		Level: raven.ERROR,
		ServerName: hostname,
	}
}
