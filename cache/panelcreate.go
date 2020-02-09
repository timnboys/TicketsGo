package cache

import (
	"encoding/json"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
)

func (c *RedisClient) ListenPanelCreate(message chan database.Panel) {
	pubsub := c.Subscribe("tickets:panel:create")

	for {
		msg, err := pubsub.ReceiveMessage(); if err != nil {
			sentry.Error(err)
			continue
		}

		var decoded database.Panel
		if err := json.Unmarshal([]byte(msg.Payload), &decoded); err != nil {
			sentry.Error(err)
			continue
		}

		message<-decoded
	}
}
