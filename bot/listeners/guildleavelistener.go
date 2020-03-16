package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/bwmarrin/discordgo"
)

/*
 * Sent when a guild becomes unavailable during a guild outage, or when the user leaves or is removed from a guild.
 * The inner payload is an unavailable guild object.
* If the unavailable field is not set, the user was removed from the guild.
 */
func OnGuildLeave(s *discordgo.Session, e *discordgo.GuildDelete) {
	if !e.Unavailable {
		servercounter.UpdateCache(s.ShardID, len(s.State.Guilds))
		go statsd.IncrementKey(statsd.LEAVES)
	}
}