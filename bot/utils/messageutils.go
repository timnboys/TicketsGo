package utils

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/rest"
	"time"
)

type Colour int

const (
	Green  Colour = 0x2ECC71
	Red    Colour = 0xFC3F35
	Orange Colour = 16740864
	Lime   Colour = 7658240
	Blue   Colour = 472219
)

type SentMessage struct {
	Shard   *gateway.Shard
	Message *message.Message
}

func SendEmbed(session *gateway.Shard, channel uint64, colour Colour, title, content string, fields []embed.EmbedField, deleteAfter int, isPremium bool) {
	_, _ = SendEmbedWithResponse(session, channel, colour, title, content, fields, deleteAfter, isPremium)
}

func SendEmbedWithResponse(shard *gateway.Shard, channel uint64, colour Colour, title, content string, fields []embed.EmbedField, deleteAfter int, isPremium bool) (message.Message, error) {
	msgEmbed := embed.NewEmbed().
		SetColor(int(colour)).
		SetTitle(title).
		SetDescription(content)

	for _, field := range fields {
		msgEmbed.AddField(field.Name, field.Value, field.Inline)
	}

	if !isPremium {
		msgEmbed.SetFooter("Powered by ticketsbot.net", shard.SelfAvatar(256))
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

		return msg, err
	}

	if deleteAfter > 0 {
		DeleteAfter(SentMessage{shard, &msg}, deleteAfter)
	}

	return msg, err
}

func DeleteAfter(msg SentMessage, secs int) {
	go func() {
		time.Sleep(time.Duration(secs) * time.Second)

		// Fix a panic
		if msg.Message != nil && msg.Shard != nil && msg.Shard.ShardManager != nil {
			// Explicitly ignore error, pretty much always a 404
			_ = msg.Shard.DeleteMessage(msg.Message.ChannelId, msg.Message.Id)
		}
	}()
}

func ReactWithCheck(shard *gateway.Shard, msg message.MessageReference) {
	if err := shard.CreateReaction(msg.ChannelId, msg.MessageId, "✅"); err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildId,
			Channel: msg.ChannelId,
			Shard:   shard.ShardId,
		})
	}
}

func ReactWithCross(shard *gateway.Shard, msg message.MessageReference) {
	if err := shard.CreateReaction(msg.ChannelId, msg.MessageId, "❌"); err != nil {
		sentry.LogWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildId,
			Channel: msg.ChannelId,
			Shard:   shard.ShardId,
		})
	}
}

func PadDiscriminator(discrim uint16) string {
	return fmt.Sprintf("%04d", discrim)
}
