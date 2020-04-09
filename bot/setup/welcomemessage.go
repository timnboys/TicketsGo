package setup

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
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
	return "No message specified"
}

func (WelcomeMessageStage) Process(shard *gateway.Shard, msg message.Message) {
	go database.SetWelcomeMessage(msg.GuildId, msg.Content)
	utils.ReactWithCheck(shard, message.MessageReference{
		MessageId: msg.Id,
		ChannelId: msg.ChannelId,
		GuildId:   msg.GuildId,
	})
}
