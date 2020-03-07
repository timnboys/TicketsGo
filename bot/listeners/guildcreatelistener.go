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

	cached, err := s.State.Guild(e.Guild.ID)
	isJoin := false
	if err == nil || cached == nil || cached.Unavailable {
		isJoin = true
	}

	guildId, err := strconv.ParseInt(e.Guild.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

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