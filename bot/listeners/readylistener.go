package listeners

import (
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"sync"
)

var(
	// We need to keep track of guilds that we're in so we can determine whether GuildCreate is a join or a lazy load
	JoinedGuilds     = make(map[int][]uint64, 0)
	JoinedGuildsLock sync.Mutex
)

func OnReady(s *gateway.Shard, e *events.Ready) {
	ids := make([]uint64, 0)
	for _, guild := range e.Guilds {
		ids = append(ids, guild.Id)
	}

	JoinedGuildsLock.Lock()
	JoinedGuilds[s.ShardId] = ids
	JoinedGuildsLock.Unlock()
}

func trackCachedGuild(shard int, id uint64) {
	JoinedGuildsLock.Lock()
	JoinedGuilds[shard] = append(JoinedGuilds[shard], id)
	JoinedGuildsLock.Unlock()
}
