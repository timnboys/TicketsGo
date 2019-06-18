package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/bwmarrin/discordgo"
)

type CommandContext struct {
	Session *discordgo.Session
	User discordgo.User
	Guild string
	Channel string
	Message discordgo.Message
	Root string
	Args []string
	IsPremium bool
}

func (cc *CommandContext) SendEmbed(colour utils.Colour, title, content string) {
	utils.SendEmbed(cc.Session, cc.Channel, colour, title, content, 30, cc.IsPremium)
}

func (cc *CommandContext) SendEmbedNoDelete(colour utils.Colour, title, content string) {
	utils.SendEmbed(cc.Session, cc.Channel, colour, title, content, 0, cc.IsPremium)
}

func (cc *CommandContext) ReactWithCheck() {
	utils.ReactWithCheck(cc.Session, &cc.Message)
}

func (cc *CommandContext) ReactWithCross() {
	utils.ReactWithCross(cc.Session, cc.Message)
}

func (cc *CommandContext) GetPermissionLevel(ch chan utils.PermissionLevel) {
	utils.GetPermissionLevel(cc.Guild, cc.User.ID, ch)
}
