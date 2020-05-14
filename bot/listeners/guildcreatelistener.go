package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/guild"
)

// Fires when we receive a guild
func OnGuildCreate(s *gateway.Shard, e *events.GuildCreate) {
	// Determine whether this is a join or lazy load
	var exists bool

	ExistingGuildsLock.RLock()

	for _, guildId := range ExistingGuilds {
		if e.Id == guildId {
			exists = true
			break
		}
	}

	ExistingGuildsLock.RUnlock()

	if !exists {
		go statsd.IncrementKey(statsd.JOINS)

		sendOwnerMessage(s, &e.Guild)
	}
}

func sendOwnerMessage(shard *gateway.Shard, guild *guild.Guild) {
	// Create DM channel
	channel, err := shard.CreateDM(guild.OwnerId)
	if err != nil { // User probably has DMs disabled
		return
	}

	msg := embed.NewEmbed().
		SetTitle("Tickets").
		SetDescription("Thank you for inviting Tickets to your server! Below is a quick guide on setting the bot up, please don't hesitate to contact us in our [support server](https://discord.gg/VtV3rSk) if you need any assistance!").
		SetColor(int(utils.Green)).
		AddField("Setup", "You can setup the bot using `t!setup`, or you can use the [dashboard](https://panel.ticketsbot.net) which has additional options", false).
		AddField("Reaction Panels", fmt.Sprintf("Reaction panels are a commonly used feature of the bot. You can read about them [here](https://ticketsbot.net/panels), or create one on [the dashboard](https://panel.ticketsbot.net/manage/%d/panels)", guild.Id), false).
		AddField("Adding Staff", "To make staff able to answer tickets, you must let the bot know about them first. You can do this through\n`t!addsupport [@User / @Role]` and `t!addadmin [@User / @Role]`. Administrators can change the settings of the bot and access the dashboard.", false).
		AddField("Tags", "Tags are predefined tickets of text which you can access through a simple command. You can learn more about them [here](https://ticketsbot.net/tags).", false).
		AddField("Claiming", "Tickets can be claimed by your staff such that other staff members cannot also reply to the ticket. You can learn more about claiming [here](https://ticketsbot.net/claiming).", false).
		AddField("Additional Support", "If you are still confused, we welcome you to our [support server](https://discord.gg/VtV3rSk). Cheers.", false)

	_, _ = shard.CreateMessageEmbed(channel.Id, msg)
}
