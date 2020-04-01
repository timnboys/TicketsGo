package cache

import (
	"encoding/json"
	"github.com/TicketsBot/TicketsGo/sentry"
)

type TicketMessage struct {
	GuildId  uint64 `json:"guild,string"` // TODO: Refactor this on we bUI side
	TicketId int    `json:"ticket"`
	Username string `json:"username"`
	Content  string `json:"content"`
}

func (c *RedisClient) PublishMessage(msg TicketMessage) {
	encoded, err := json.Marshal(msg); if err != nil {
		sentry.Error(err)
		return
	}

	c.Publish("tickets:webchat:inboundmessage", string(encoded))
}
