package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
)

type PrefixStage struct {
}

func (PrefixStage) State() State {
	return Prefix
}

func (PrefixStage) Prompt() string {
	return "Type the prefix that you would like to use for the bot" +
		"\nThe prefix is the characters that come *before* the command (excluding the actual command itself)" +
		"\nExample: `t!`"
}

func (PrefixStage) Default() string {
	return config.Conf.Bot.Prefix
}

func (PrefixStage) Process(shard *gateway.Shard, msg message.Message) {
	if len(msg.Content) > 8 {
		utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("The maxium prefix langeth is 8 characters\nDefaulting to `%s`", PrefixStage{}.Default()), nil, 15, true)
		return
	}

	ref := message.MessageReference{
		MessageId: msg.Id,
		ChannelId: msg.ChannelId,
		GuildId:   msg.GuildId,
	}

	if err := database.Client.Prefix.Set(msg.GuildId, msg.Content); err == nil {
		utils.ReactWithCheck(shard, ref)
	} else {
		utils.ReactWithCross(shard, ref)
		sentry.Error(err)
	}
}
