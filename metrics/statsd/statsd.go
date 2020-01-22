package statsd

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	stats "gopkg.in/alexcesaro/statsd.v2"
)

type StatsdClient struct {
	*stats.Client
}

var Client StatsdClient

func NewClient() (StatsdClient, error) {
	addr := fmt.Sprintf("%s:%d", config.Conf.Metrics.Statsd.Host, config.Conf.Metrics.Statsd.Port)
	client, err := stats.New(stats.Address(addr), stats.Prefix(config.Conf.Metrics.Statsd.Prefix)); if err != nil {
		return StatsdClient{}, err
	}

	return StatsdClient{
		client,
	}, nil
}

func IsClientNull() bool {
	return Client.Client == nil
}

func IncrementKey(key Key) {
	if IsClientNull() {
		return
	}

	Client.Increment(key.String())
}