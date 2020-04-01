package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel"
)

func storeChannel(channel *channel.Channel) {
	go database.StoreChannel(channel.Id, channel.GuildId, channel.Name, int(channel.Type))
}

func OnChannelCreate(s *gateway.Shard, e *events.ChannelCreate) {
	if e.GuildId == 0 {
		return
	}

	storeChannel(e.Channel)
}

func OnChannelUpdate(_ *gateway.Shard, e *events.ChannelUpdate) {
	storeChannel(e.Channel)
}
