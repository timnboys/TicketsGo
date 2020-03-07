package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/bwmarrin/discordgo"
)

// Fires when we get kicked from a guild
func OnGuildLeave(s *discordgo.Session, e *discordgo.GuildDelete) {
	servercounter.UpdateCache(s.ShardID, len(s.State.Guilds))
	go statsd.IncrementKey(statsd.LEAVES)
}