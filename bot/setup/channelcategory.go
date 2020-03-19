package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
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
	return "Type the **name** of the **channel category** that you would like tickets to be created under"
}

func (ChannelCategoryStage) Default() string {
	return ""
}

func (ChannelCategoryStage) Process(session *discordgo.Session, msg discordgo.Message) {
	errorContext := sentry.ErrorContext{
		Guild:   msg.GuildID,
		User:    msg.Author.ID,
		Channel: msg.ChannelID,
		Shard:   session.ShardID,
	}

	guildId, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	name := msg.Content

	guild, err := session.State.Guild(msg.GuildID); if err != nil {
		// Not cached
		guild, err = session.Guild(msg.GuildID); if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}
	}

	var categoryName string
	for _, channel := range guild.Channels {
		if strings.ToLower(channel.Name) == strings.ToLower(name) {
			categoryName = channel.ID
			break
		}
	}

	categoryId, err := strconv.ParseInt(categoryName, 10, 64)

	if categoryName == "" || err != nil {
		// Attempt to create categoryName
		data := discordgo.GuildChannelCreateData{
			Name: categoryName,
			Type: discordgo.ChannelTypeGuildCategory,
		}

		category, err := session.GuildChannelCreateComplex(guild.ID, data); if err != nil {
			// Likely no permission, default to having no category
			utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", "Invalid category\nDefaulting to using no category", 15, true)
			return
		}

		categoryId, err = strconv.ParseInt(category.ID, 10, 64); if err != nil {
			sentry.ErrorWithContext(err, errorContext)
			return
		}

		utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", fmt.Sprintf("I have created the channel category %s for you, you may need to adjust permissions yourself", category.Name), 15, true)
	}

	go database.SetCategory(guildId, categoryId)
	utils.ReactWithCheck(session, &msg)
}
