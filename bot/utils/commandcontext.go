package utils

import (
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/guild"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/objects/user"
	"strings"
)

type CommandContext struct {
	Shard       *gateway.Shard
	User        *user.User
	Guild       *guild.Guild
	ChannelId   uint64
	Message     *message.Message
	Root        string
	Args        []string
	IsPremium   bool
	ShouldReact bool
	Member      *member.Member
	IsFromPanel bool
}

func (ctx *CommandContext) ToErrorContext() sentry.ErrorContext {
	var guildId uint64
	if ctx.Guild != nil {
		guildId = ctx.Guild.Id
	}

	return sentry.ErrorContext{
		Guild:       guildId,
		User:        ctx.User.Id,
		Channel:     ctx.ChannelId,
		Shard:       ctx.Shard.ShardId,
		Command:     ctx.Root + " " + strings.Join(ctx.Args, " "),
		Premium:     ctx.IsPremium,
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
		ReactWithCheck(ctx.Shard, ctx.Message)
	}
}

func (ctx *CommandContext) ReactWithCross() {
	if ctx.ShouldReact {
		ReactWithCross(ctx.Shard, ctx.Message)
	}
}

func (ctx *CommandContext) GetPermissionLevel(ch chan PermissionLevel) {
	GetPermissionLevel(ctx.Shard, ctx.Member, ctx.Guild.Id, ch)
}
