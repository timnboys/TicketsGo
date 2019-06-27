package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/bwmarrin/discordgo"
)

// Fires when we receive a guild
func OnGuildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	println(s.ShardID)
	servercounter.UpdateCache(s.ShardID, len(s.State.Guilds))
}
