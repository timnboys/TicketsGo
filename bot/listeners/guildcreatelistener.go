package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot"
	"github.com/bwmarrin/discordgo"
)

// Fires when we receive a guild
func OnGuildCreate(s *discordgo.Session, e *discordgo.GuildCreate) {
	bot.UpdateCache(s.ShardID, len(s.State.Guilds))
}
