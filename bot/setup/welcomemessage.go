package setup

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
)

type WelcomeMessageStage struct {
}

func (WelcomeMessageStage) State() State {
	return WelcomeMessage
}

func (WelcomeMessageStage) Prompt() string {
	return "Type the message that should be sent by the bot when a ticket is opened"
}

func (WelcomeMessageStage) Default() string {
	return "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
}

func (WelcomeMessageStage) Process(shard *gateway.Shard, msg message.Message) {
	ref := message.MessageReference{
		MessageId: msg.Id,
		ChannelId: msg.ChannelId,
		GuildId:   msg.GuildId,
	}

	if err := database.Client.WelcomeMessages.Set(msg.GuildId, msg.Content); err == nil {
		utils.ReactWithCheck(shard, ref)
	} else {
		utils.ReactWithCross(shard, ref)
		sentry.Error(err)
	}
}
