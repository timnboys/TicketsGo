package listeners

import (
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"sync"
)

var (
	// We need to keep track of guilds that we're in so we can determine whether GuildCreate is a join or a lazy load
	joinedGuilds     = make(map[int][]uint64, 0)
	joinedGuildsLock sync.RWMutex
)

func OnReady(s *gateway.Shard, e *events.Ready) {
	joinedGuildsLock.Lock()
	ids := make([]uint64, 0)
	for _, guild := range e.Guilds {
		ids = append(ids, guild.Id)
	}

	joinedGuilds[s.ShardId] = ids
	joinedGuildsLock.Unlock()
}

func trackCachedGuild(shard int, id uint64) {
	joinedGuildsLock.Lock()
	joinedGuilds[shard] = append(joinedGuilds[shard], id)
	joinedGuildsLock.Unlock()
}
