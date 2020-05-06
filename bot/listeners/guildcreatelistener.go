package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/guild"
)

// Fires when we receive a guild
func OnGuildCreate(s *gateway.Shard, e *events.GuildCreate) {
	// Determine whether this is a join or lazy load
	_, exists := s.Cache.GetGuild(e.Id, false)

	fmt.Println(exists)

	if !exists {
		go statsd.IncrementKey(statsd.JOINS)

		//sendOwnerMessage(s, &e.Guild)
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

	utils.SendEmbed(shard, channel.Id, utils.Green, "Tickets", message, nil, 0, false)
}
