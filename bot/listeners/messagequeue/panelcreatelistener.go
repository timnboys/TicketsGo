package messagequeue

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/jonas747/dshardmanager"
	"strconv"
)

func ListenPanelCreations(shardManager *dshardmanager.Manager) {
	creations := make(chan database.Panel)
	go cache.Client.ListenPanelCreate(creations)

	for panel := range creations {
		// Get shard
		shard := shardManager.SessionForGuild(panel.GuildId); if shard == nil {
			continue
		}

		guildIdStr := strconv.Itoa(int(panel.GuildId))
		channelIdStr := strconv.Itoa(int(panel.ChannelId))

		errorContext := sentry.ErrorContext{
			Guild:       guildIdStr,
			Channel:     channelIdStr,
			Shard:       shard.ShardID,
		}

		// Get guild object
		guild, err := shard.State.Guild(guildIdStr); if err != nil {
			guild, err = shard.Guild(guildIdStr); if err != nil {
				sentry.ErrorWithContext(err, errorContext)
				continue
			}
		}

		// Create embed object
		embed := utils.NewEmbed()

		// Get whether guild is premium
		isPremiumChan := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Shard:   shard,
			Guild:   guild,
			GuildId: panel.GuildId,
		}, isPremiumChan)
		isPremium := <-isPremiumChan

		if !isPremium {
			embed.SetFooter("Powered by ticketsbot.net", utils.AvatarUrl)
		}

		embed.SetTitle(panel.Title)
		embed.SetDescription(panel.Content)
		embed.SetColor(panel.Colour)

		msg, err := shard.ChannelMessageSendEmbed(channelIdStr, embed.MessageEmbed); if err != nil {
			sentry.LogWithContext(err, errorContext)
			continue
		}

		msgId, err := strconv.ParseInt(msg.ID, 10, 64); if err != nil {
			sentry.LogWithContext(err, errorContext)
			continue
		}

		// ReactionEmote is the unicode emoji
		if err = shard.MessageReactionAdd(channelIdStr, msg.ID, panel.ReactionEmote); err != nil {
			sentry.LogWithContext(err, sentry.ErrorContext{})
		}

		go database.AddPanel(msgId, panel.ChannelId, panel.GuildId, panel.Title, panel.Content, panel.Colour, panel.TargetCategory, panel.ReactionEmote)
	}
}
