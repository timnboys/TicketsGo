package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/bwmarrin/discordgo"
)

// Fires when we receive a guild
func OnGuildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	servercounter.UpdateCache(s.ShardID, len(s.State.Guilds))

	for _, channel := range e.Channels {
		storeChannel(channel)
	}
}