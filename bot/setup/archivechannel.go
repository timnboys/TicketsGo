package setup

import (
	"errors"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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

func (ArchiveChannelStage) Process(session *discordgo.Session, msg discordgo.Message) {
	guildId, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildID,
			User:    msg.Author.ID,
			Channel: msg.ChannelID,
			Shard:   session.ShardID,
		})
		return
	}

	guild, err := session.State.Guild(msg.GuildID); if err != nil {
		// Not cached
		guild, err = session.Guild(msg.GuildID); if err != nil {
			sentry.ErrorWithContext(err, sentry.ErrorContext{
				Guild:   msg.GuildID,
				User:    msg.Author.ID,
				Channel: msg.ChannelID,
				Shard:   session.ShardID,
			})
			return
		}
	}

	var id string

	// Prefer channel mention
	if len(msg.MentionChannels) > 0 {
		channel := msg.MentionChannels[0]
		if channel == nil { // Shouldn't ever happen, but best to be safe
			sentry.Error(errors.New("channel is nil"))
			return
		}

		// Verify that the channel exists
		exists := false
		for _, channel := range guild.Channels {
			if channel.ID == id {
				exists = true
				break
			}
		}

		if !exists {
			utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", "Invalid channel, disabling archiving", 15, true)
			utils.ReactWithCross(session, msg)
			return
		}

		id = channel.ID
	} else {
		// Try to match channel name
		split := strings.Split(msg.Content, " ")
		name := split[0]

		// Get channels from discord
		channels, err := session.GuildChannels(msg.GuildID); if err != nil {
			utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()), 15, true)
			return
		}

		found := false
		for _, channel := range channels {
			if channel == nil { // Best to be safe
				continue
			}

			if channel.Name == name {
				found = true
				id = channel.ID
				break
			}
		}

		if !found {
			utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", "Invalid channel, disabling archiving", 15, true)
			utils.ReactWithCross(session, msg)
			return
		}
	}

	archiveChannel, err := strconv.ParseInt(id, 10, 64); if err != nil { // Shouldn't ever happen
		sentry.Error(err)
		return
	}

	go database.SetArchiveChannel(guildId, archiveChannel)
	utils.ReactWithCheck(session, &msg)
}
