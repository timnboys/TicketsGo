package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

// Fires when we receive a guild
func OnGuildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	servercounter.UpdateCache(s.ShardID, len(s.State.Guilds))

	guildId, err := strconv.ParseInt(e.Guild.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	// Determine whether this is a join or lazy load
	JoinedGuildsLock.Lock()
	cachedGuilds := JoinedGuilds[s.ShardID]
	JoinedGuildsLock.Unlock()

	isJoin := true
	for _, cachedId := range cachedGuilds {
		if cachedId == e.Guild.ID {
			isJoin = false
			break
		}
	}

	trackCachedGuild(s.ShardID, e.Guild.ID)

	if isJoin {
		go statsd.IncrementKey(statsd.JOINS)

		channels := make([]database.Channel, 0)
		for _, channel := range e.Channels {
			channelId, err := strconv.ParseInt(channel.ID, 10, 64); if err != nil {
				sentry.Error(err)
				return
			}

			channels = append(channels, database.Channel{
				ChannelId: channelId,
				GuildId:   guildId,
				Name:      channel.Name,
				Type:      int(channel.Type),
			})
		}

		go database.InsertChannels(channels)
	}
}