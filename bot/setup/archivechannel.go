package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
	"strings"
)

type ArchiveChannelStage struct {
}

func (ArchiveChannelStage) State() State {
	return ArchiveChannel
}

func (ArchiveChannelStage) Prompt() string {
	return "Please specify you wish ticket logs to be sent to after tickets have been closed" +
		"\nExample: `#logs`"
}

func (ArchiveChannelStage) Default() string {
	return ""
}

func (ArchiveChannelStage) Process(shard *gateway.Shard, msg message.Message) {
	guild, err := shard.GetGuild(msg.GuildId); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildId,
			User:    msg.Author.Id,
			Channel: msg.ChannelId,
			Shard:   shard.ShardId,
		})
		return
	}

	var archiveChannelId uint64

	// Prefer channel mention
	if len(msg.MentionChannels) > 0 {
		archiveChannelId = msg.MentionChannels[0].Id

		// Verify that the channel exists
		exists := false
		for _, guildChannel := range guild.Channels {
			if guildChannel.Id == archiveChannelId {
				exists = true
				break
			}
		}

		if !exists {
			utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", "Invalid channel, disabling archiving", 15, true)
			utils.ReactWithCross(shard, &msg)
			return
		}
	} else {
		// Try to match channel name
		split := strings.Split(msg.Content, " ")
		name := split[0]

		// Get channels from discord
		channels, err := shard.GetGuildChannels(msg.GuildId); if err != nil {
			utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()), 15, true)
			return
		}

		found := false
		for _, channel := range channels {
			if channel.Name == name {
				found = true
				archiveChannelId = channel.Id
				break
			}
		}

		if !found {
			utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", "Invalid channel, disabling archiving", 15, true)
			utils.ReactWithCross(shard, &msg)
			return
		}
	}

	go database.SetArchiveChannel(msg.GuildId, archiveChannelId)
	utils.ReactWithCheck(shard, &msg)
}
