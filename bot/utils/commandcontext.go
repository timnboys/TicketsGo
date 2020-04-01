package utils

import (
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
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
		Permissions: ctx.GetCachedPermissions(ctx.ChannelId),
	}
}

func (ctx *CommandContext) SendEmbed(colour Colour, title, content string) {
	SendEmbed(ctx.Shard, ctx.Channel, colour, title, content, 30, ctx.IsPremium)
}

func (ctx *CommandContext) SendEmbedNoDelete(colour Colour, title, content string) {
	SendEmbed(ctx.Shard, ctx.Channel, colour, title, content, 0, ctx.IsPremium)
}

func (ctx *CommandContext) SendMessage(content string) {
	msg, err := ctx.Shard.ChannelMessageSend(ctx.Channel, content)
	if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
	} else {
		DeleteAfter(SentMessage{Shard: ctx.Shard, Message: msg}, 60)
	}
}

func (ctx *CommandContext) ReactWithCheck() {
	if ctx.ShouldReact {
		ReactWithCheck(ctx.Shard, &ctx.Message)
	}
}

func (ctx *CommandContext) ReactWithCross() {
	if ctx.ShouldReact {
		ReactWithCross(ctx.Shard, ctx.Message)
	}
}

func (ctx *CommandContext) GetPermissionLevel(ch chan PermissionLevel) {
	GetPermissionLevel(ctx.Shard, ctx.Member, ch)
}

func (ctx *CommandContext) ChannelMemberHasPermission(channel, user string, permission Permission, ch chan bool) {
	hasAdmin := make(chan bool)
	go ChannelMemberHasPermission(ctx.Shard, ctx.Guild.ID, channel, user, Administrator, hasAdmin)
	if <-hasAdmin {
		ch <- true
	} else {
		ChannelMemberHasPermission(ctx.Shard, ctx.Guild.ID, channel, user, permission, ch)
	}
}

func (ctx *CommandContext) MemberHasPermission(user string, permission Permission, ch chan bool) {
	hasAdmin := make(chan bool)
	go MemberHasPermission(ctx.Shard, ctx.Guild.ID, user, Administrator, hasAdmin)
	if <-hasAdmin {
		ch <- true
	} else {
		go MemberHasPermission(ctx.Shard, ctx.Guild.ID, user, permission, ch)
	}
}

func (ctx *CommandContext) GetCachedPermissions(ch string) []*discordgo.PermissionOverwrite {
	channel, err := ctx.Shard.State.Channel(ch)
	if err != nil {
		return make([]*discordgo.PermissionOverwrite, 0)
	}

	return channel.PermissionOverwrites
}
