package listeners

import (
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"sync"
)

var (
	ExistingGuilds     []uint64
	ExistingGuildsLock sync.RWMutex
)

func OnReady(s *gateway.Shard, e *events.Ready) {
	ExistingGuildsLock.Lock()

	for _, guild := range e.Guilds {
		ExistingGuilds = append(ExistingGuilds, guild.Id)
	}

	ExistingGuildsLock.Unlock()
}
