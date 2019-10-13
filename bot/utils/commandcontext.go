package utils

import (
	"github.com/bwmarrin/discordgo"
)

type CommandContext struct {
	Session     *discordgo.Session
	User        discordgo.User
	UserID		int64
	Guild       string
	GuildId 	int64
	Channel     string
	Message     discordgo.Message
	Root        string
	Args        []string
	IsPremium   bool
	ShouldReact bool
}

func (ctx *CommandContext) SendEmbed(colour Colour, title, content string) {
	SendEmbed(ctx.Session, ctx.Channel, colour, title, content, 30, ctx.IsPremium)
}

func (ctx *CommandContext) SendEmbedNoDelete(colour Colour, title, content string) {
	SendEmbed(ctx.Session, ctx.Channel, colour, title, content, 0, ctx.IsPremium)
}

func (ctx *CommandContext) ReactWithCheck() {
	ReactWithCheck(ctx.Session, &ctx.Message)
}

func (ctx *CommandContext) ReactWithCross() {
	ReactWithCross(ctx.Session, ctx.Message)
}

func (ctx *CommandContext) GetPermissionLevel(ch chan PermissionLevel) {
	GetPermissionLevel(ctx.Session, ctx.Guild, ctx.User.ID, ch)
}

func (ctx *CommandContext) ChannelMemberHasPermission(channel, user string, permission Permission, ch chan bool) {
	hasAdmin := make(chan bool)
	go ChannelMemberHasPermission(ctx.Session, ctx.Guild, channel, user, Administrator, hasAdmin)
	if <-hasAdmin {
		ch <- true
	} else {
		ChannelMemberHasPermission(ctx.Session, ctx.Guild, channel, user, permission, ch)
	}
}

func (ctx *CommandContext) MemberHasPermission(user string, permission Permission, ch chan bool) {
	hasAdmin := make(chan bool)
	go MemberHasPermission(ctx.Session, ctx.Guild, user, Administrator, hasAdmin)
	if <-hasAdmin {
		ch <- true
	} else {
		MemberHasPermission(ctx.Session, ctx.Guild, user, permission, ch)
	}
}
