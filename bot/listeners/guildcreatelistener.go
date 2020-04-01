package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/guild"
)

// Fires when we receive a guild
func OnGuildCreate(s *gateway.Shard, e *events.GuildCreate) {
	servercounter.UpdateCache(s.ShardId, len((*s.Cache).GetGuilds()))

	// Determine whether this is a join or lazy load
	JoinedGuildsLock.Lock()
	cachedGuilds := JoinedGuilds[s.ShardId]
	JoinedGuildsLock.Unlock()

	isJoin := true
	for _, cachedId := range cachedGuilds {
		if cachedId == e.Guild.Id {
			isJoin = false
			break
		}
	}

	trackCachedGuild(s.ShardId, e.Guild.Id)

	if isJoin {
		go statsd.IncrementKey(statsd.JOINS)

		channels := make([]database.Channel, 0)
		for _, channel := range e.Channels {
			channels = append(channels, database.Channel{
				ChannelId: channel.Id,
				GuildId:   e.Guild.Id,
				Name:      channel.Name,
				Type:      int(channel.Type),
			})
		}

		go cache.Client.CacheGuildProperties(e.Guild)
		go database.InsertChannels(channels)

		sendOwnerMessage(s, e.Guild)
	}
}

func sendOwnerMessage(shard *gateway.Shard, guild *guild.Guild) {
	// Create DM channel
	channel, err := shard.CreateDM(guild.OwnerId)
	if err != nil { // User probably has DMs disabled
		return
	}

	message := fmt.Sprintf("Thanks for inviting Tickets to %s!\n"+
		"To get set up, start off by running `t!setup` to configure the bot. You may then wish to visit the [web UI](https://panel.ticketsbot.net/manage/%d/settings) to access further configurations, "+
		"as well as to create a [panel](https://ticketsbot.net/panels) (reactable embed that automatically opens a ticket).\n"+
		"If you require further assistance, you may wish to read the information section on our [website](https://ticketsbot.net), or if you prefer, feel free to join our [support server](https://discord.gg/VtV3rSk) to ask any questions you may have, "+
		"or to provide feedback to use (especially if you choose to switch to a competitor - we'd love to know how we can improve).",
		guild.Name, guild.Id)

	utils.SendEmbed(shard, channel.Id, utils.Green, "Tickets", message, 0, false)
}
