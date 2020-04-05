package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
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

func (ChannelCategoryStage) Process(shard *gateway.Shard, msg message.Message) {
	errorContext := sentry.ErrorContext{
		Guild:   msg.GuildId,
		User:    msg.Author.Id,
		Channel: msg.ChannelId,
		Shard:   shard.ShardId,
	}

	name := msg.Content

	guild, err := shard.GetGuild(msg.GuildId); if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	var categoryId uint64
	for _, ch := range guild.Channels {
		if ch.Type == channel.ChannelTypeGuildCategory && strings.ToLower(ch.Name) == strings.ToLower(name) {
			categoryId = ch.Id
			break
		}
	}

	if categoryId == 0 {
		// Attempt to create categoryName
		data := rest.CreateChannelData{
			Name: name,
			Type: channel.ChannelTypeGuildCategory,
		}

		category, err := shard.CreateGuildChannel(guild.Id, data); if err != nil {
			// Likely no permission, default to having no category
			utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", "Invalid category\nDefaulting to using no category", 15, true)
			return
		}

		categoryId = category.Id

		utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("I have created the channel category %s for you, you may need to adjust permissions yourself", category.Name), 15, true)
	}

	go database.SetCategory(msg.GuildId, categoryId)
	utils.ReactWithCheck(shard, &msg)
}
