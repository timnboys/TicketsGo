package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/bwmarrin/discordgo"
)

type CommandContext struct {
	Session     *discordgo.Session
	User        discordgo.User
	Guild       string
	Channel     string
	Message     discordgo.Message
	Root        string
	Args        []string
	IsPremium   bool
	ShouldReact bool
}

func (ctx *CommandContext) SendEmbed(colour utils.Colour, title, content string) {
	utils.SendEmbed(ctx.Session, ctx.Channel, colour, title, content, 30, ctx.IsPremium)
}

func (ctx *CommandContext) SendEmbedNoDelete(colour utils.Colour, title, content string) {
	utils.SendEmbed(ctx.Session, ctx.Channel, colour, title, content, 0, ctx.IsPremium)
}

func (ctx *CommandContext) ReactWithCheck() {
	utils.ReactWithCheck(ctx.Session, &ctx.Message)
}

func (ctx *CommandContext) ReactWithCross() {
	utils.ReactWithCross(ctx.Session, ctx.Message)
}

func (ctx *CommandContext) GetPermissionLevel(ch chan utils.PermissionLevel) {
	utils.GetPermissionLevel(ctx.Session, ctx.Guild, ctx.User.ID, ch)
}

func (ctx *CommandContext) ChannelMemberHasPermission(channel, user string, permission utils.Permission, ch chan bool) {
	hasAdmin := make(chan bool)
	go utils.ChannelMemberHasPermission(ctx.Session, ctx.Guild, channel, user, utils.Administrator, hasAdmin)
	if <-hasAdmin {
		ch <- true
	} else {
		utils.ChannelMemberHasPermission(ctx.Session, ctx.Guild, channel, user, permission, ch)
	}
}

func (ctx *CommandContext) MemberHasPermission(user string, permission utils.Permission, ch chan bool) {
	hasAdmin := make(chan bool)
	go utils.MemberHasPermission(ctx.Session, ctx.Guild, user, utils.Administrator, hasAdmin)
	if <-hasAdmin {
		ch <- true
	} else {
		utils.MemberHasPermission(ctx.Session, ctx.Guild, user, permission, ch)
	}
}
