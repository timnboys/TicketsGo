package listeners

import (
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"sync"
)

var(
	// We need to keep track of guilds that we're in so we can determine whether GuildCreate is a join or a lazy load
	JoinedGuilds     = make(map[int][]string, 0)
	JoinedGuildsLock sync.Mutex
)

func OnReady(s *discordgo.Session, e *discordgo.Ready) {
	s.State.TrackEmojis = false
	s.State.TrackPresences = false
	s.State.TrackVoice = false
	s.SyncEvents = false

	if err := s.UpdateStatus(0, config.Conf.Bot.Game); err != nil {
		sentry.Error(err)
	}

	ids := make([]string, 0)
	for _, guild := range e.Guilds {
		ids = append(ids, guild.ID)
	}

	JoinedGuildsLock.Lock()
	JoinedGuilds[s.ShardID] = ids
	JoinedGuildsLock.Unlock()
}

func trackCachedGuild(shard int, id string) {
	JoinedGuildsLock.Lock()
	JoinedGuilds[shard] = append(JoinedGuilds[shard], id)
	JoinedGuildsLock.Unlock()
}
