package setup

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

type ChannelCategoryStage struct {
}

func (ChannelCategoryStage) State() State {
	return ChannelCategory
}

func (ChannelCategoryStage) Prompt() string {
	return "Type the **name** of the channel category that you would like tickets to be created under"
}

func (ChannelCategoryStage) Default() string {
	return ""
}

func (ChannelCategoryStage) Process(session *discordgo.Session, msg discordgo.Message) {
	guildId, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	name := msg.Content

	guild, err := session.State.Guild(msg.GuildID); if err != nil {
		// Not cached
		guild, err = session.Guild(msg.GuildID); if err != nil {
			sentry.Error(err)
			return
		}
	}

	var category string
	for _, channel := range guild.Channels {
		if strings.ToLower(channel.Name) == strings.ToLower(name) {
			category = channel.ID
			break
		}
	}

	categoryId, err := strconv.ParseInt(category, 10, 64)

	if category == "" || err != nil {
		utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", "Invalid category\nDefault to no category", 15, true)
		return
	}

	go database.SetCategory(guildId, categoryId)
	utils.ReactWithCheck(session, &msg)
}
