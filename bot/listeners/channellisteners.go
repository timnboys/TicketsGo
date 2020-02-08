package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func storeChannel(channel *discordgo.Channel) {
	channelId, err := strconv.ParseInt(channel.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	guildId, err := strconv.ParseInt(channel.GuildID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	go database.StoreChannel(channelId, guildId, channel.Name, int(channel.Type))
}

func OnChannelCreate(s *discordgo.Session, e *discordgo.ChannelCreate) {
	storeChannel(e.Channel)
}

func OnChannelUpdate(s *discordgo.Session, e *discordgo.ChannelUpdate) {
	storeChannel(e.Channel)
}
