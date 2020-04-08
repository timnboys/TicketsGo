package utils

import (
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"strings"
)

type CommandContext struct {
	Shard *gateway.Shard
	message.Message
	Root        string
	Args        []string
	IsPremium   bool
	ShouldReact bool
	IsFromPanel bool
}

func (ctx *CommandContext) Guild() (guild.Guild, error) {
	return ctx.Shard.GetGuild(ctx.GuildId)
}

func (ctx *CommandContext) ToErrorContext() sentry.ErrorContext {
	return sentry.ErrorContext{
		Guild:   ctx.GuildId,
		User:    ctx.Author.Id,
		Channel: ctx.ChannelId,
		Shard:   ctx.Shard.ShardId,
		Command: ctx.Root + " " + strings.Join(ctx.Args, " "),
		Premium: ctx.IsPremium,
	}
}

func (ctx *CommandContext) SendEmbed(colour Colour, title, content string) {
	SendEmbed(ctx.Shard, ctx.ChannelId, colour, title, content, 30, ctx.IsPremium)
}

func (ctx *CommandContext) SendEmbedNoDelete(colour Colour, title, content string) {
	SendEmbed(ctx.Shard, ctx.ChannelId, colour, title, content, 0, ctx.IsPremium)
}

func (ctx *CommandContext) SendMessage(content string) {
	msg, err := ctx.Shard.CreateMessage(ctx.ChannelId, content)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		DeleteAfter(SentMessage{Shard: ctx.Shard, Message: &msg}, 60)
	}
}

func (ctx *CommandContext) ReactWithCheck() {
	if ctx.ShouldReact {
		ReactWithCheck(ctx.Shard, message.MessageReference{
			MessageId: ctx.Id,
			ChannelId: ctx.ChannelId,
			GuildId:   ctx.GuildId,
		})
	}
}

func (ctx *CommandContext) ReactWithCross() {
	if ctx.ShouldReact {
		ReactWithCross(ctx.Shard, message.MessageReference{
			MessageId: ctx.Id,
			ChannelId: ctx.ChannelId,
			GuildId:   ctx.GuildId,
		})
	}
}

func (ctx *CommandContext) GetPermissionLevel(ch chan PermissionLevel) {
	GetPermissionLevel(ctx.Shard, ctx.Member, ctx.GuildId, ch)
}
