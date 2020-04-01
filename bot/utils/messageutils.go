package utils

import (
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"time"
)

type Colour int

const (
	Green  Colour = 2335514
	Red    Colour = 11010048
	Orange Colour = 16740864
	Lime   Colour = 7658240
	Blue   Colour = 472219
)

var (
	AvatarUrl           string
	Id                  string
)

type SentMessage struct {
	Shard   *gateway.Shard
	Message *message.Message
}

func SendEmbed(session *gateway.Shard, channel uint64, colour Colour, title, content string, deleteAfter int, isPremium bool) {
	_ = SendEmbedWithResponse(session, channel, colour, title, content, deleteAfter, isPremium)
}

func SendEmbedWithResponse(shard *gateway.Shard, channel uint64, colour Colour, title, content string, deleteAfter int, isPremium bool) *message.Message {
	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
		AddField(title, content, false)

	if !isPremium {
		msgEmbed.SetFooter("Powered by ticketsbot.net", AvatarUrl)
	}

	// Explicitly ignore error because it's usually a 403 (missing permissions)
	msg, err := shard.CreateMessageComplex(channel, rest.CreateMessageData{
		Embed: msgEmbed,
	})

	if err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Channel: channel,
			Shard:   shard.ShardId,
			Premium: isPremium,
		})
	}

	if deleteAfter > 0 {
		DeleteAfter(SentMessage{shard, msg}, deleteAfter)
	}

	return msg
}

func DeleteAfter(msg SentMessage, secs int) {
	go func() {
		time.Sleep(time.Duration(secs) * time.Second)

		// Fix a panic
		if msg.Message != nil && msg.Shard != nil {
			// Explicitly ignore error, pretty much always a 404
			_ = msg.Shard.DeleteMessage(msg.Message.ChannelId, msg.Message.Id)
		}
	}()
}

func ReactWithCheck(shard *gateway.Shard, msg *message.Message) {
	if err := shard.CreateReaction(msg.ChannelId, msg.Id, "✅"); err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildId,
			User:    msg.Author.Id,
			Channel: msg.ChannelId,
			Shard:   shard.ShardId,
		})
	}
}

func ReactWithCross(shard *gateway.Shard, msg message.Message) {
	if err := shard.CreateReaction(msg.ChannelId, msg.Id, "❌"); err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildId,
			User:    msg.Author.Id,
			Channel: msg.ChannelId,
			Shard:   shard.ShardId,
		})
	}
}
