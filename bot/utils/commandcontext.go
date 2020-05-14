package utils

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"strconv"
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

func (ctx *CommandContext) SendEmbed(colour Colour, title, content string, fields ...embed.EmbedField) {
	SendEmbed(ctx.Shard, ctx.ChannelId, colour, title, content, fields, 30, ctx.IsPremium)
}

func (ctx *CommandContext) SendEmbedNoDelete(colour Colour, title, content string, fields ...embed.EmbedField) {
	SendEmbed(ctx.Shard, ctx.ChannelId, colour, title, content, fields, 0, ctx.IsPremium)
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

func (ctx *CommandContext) GetChannelFromArgs() uint64 {
	mentions := ctx.ChannelMentions()
	if len(mentions) > 0 {
		return mentions[0]
	}

	for _, arg := range ctx.Args {
		if parsed, err := strconv.ParseUint(arg, 10, 64); err == nil {
			return parsed
		}
	}

	return 0
}

func (ctx *CommandContext) HandleError(err error) {
	sentry.ErrorWithContext(err, ctx.ToErrorContext())
	ctx.SendEmbed(Red, "Error", fmt.Sprintf("An error occurred: `%s`", err.Error()))
}
