package cache

import (
	"encoding/json"
	"github.com/TicketsBot/TicketsGo/sentry"
)

type TicketCloseMessage struct {
	Guild    uint64
	TicketId int
	User     uint64
	Reason   string
}

func (c *RedisClient) ListenTicketClose(message chan TicketCloseMessage) {
	pubsub := c.Subscribe("tickets:close")

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			sentry.Error(err)
			continue
		}

		var decoded TicketCloseMessage
		if err := json.Unmarshal([]byte(msg.Payload), &decoded); err != nil {
			sentry.Error(err)
			continue
		}

		if decoded.Reason == "" {
			decoded.Reason = "No reason specified"
		}

		message <- decoded
	}
}
