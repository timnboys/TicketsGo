package listeners

import (
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

/*
 * Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
 * The inner payload is an unavailable guild object.
 * If the unavailable field is not set, the user was removed from the guild.
 */
func OnGuildLeave(s *gateway.Shard, e *events.GuildDelete) {
	if e.Unavailable == nil {
		ExistingGuildsLock.Lock()

		for index, guildId := range ExistingGuilds {
			if guildId == e.Id {
				ExistingGuilds = append(ExistingGuilds[:index], ExistingGuilds[index+1:]...)
				break
			}
		}

		ExistingGuildsLock.Unlock()

		go statsd.IncrementKey(statsd.LEAVES)
	}
}