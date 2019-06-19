package setup

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

type ArchiveChannelStage struct {
}

func (ArchiveChannelStage) State() State {
	return ArchiveChannel
}

func (ArchiveChannelStage) Prompt() string {
	return "Type the channel that you would like ticket archives to be posted in"
}

func (ArchiveChannelStage) Default() string {
	return ""
}

func (ArchiveChannelStage) Process(session *discordgo.Session, msg discordgo.Message) {
	guildId, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	guild, err := session.State.Guild(msg.GuildID); if err != nil {
		// Not cached
		guild, err = session.Guild(msg.GuildID); if err != nil {
			log.Error(err.Error())
			return
		}
	}

	found := utils.ChannelMentionRegex.FindStringSubmatch(msg.Content)
	if len(found) == 0 {
		utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", "You need to mention a ticket channel to add the user(s) in", 15, true)
		utils.ReactWithCross(session, msg)
		return
	}

	id := found[1]
	exists := false
	for _, channel := range guild.Channels {
		if channel.ID == id {
			exists = true
			break
		}
	}

	archiveChannel, err := strconv.ParseInt(id, 10, 64)

	if !exists || err != nil {
		utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", "Invalid channel, disabling archiving", 15, true)
		utils.ReactWithCross(session, msg)
		return
	}

	go database.SetArchiveChannel(guildId, archiveChannel)
	utils.ReactWithCheck(session, &msg)
}
