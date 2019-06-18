package utils

import (
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
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

var AvatarUrl string

type SentMessage struct {
	Session *discordgo.Session
	Message *discordgo.Message
}

func SendEmbed(session *discordgo.Session, ch string, colour Colour, title, content string, deleteAfter int, isPremium bool) {
	embed := NewEmbed().
		SetColor(int(colour)).
		AddField(title, content, false)

	if !isPremium {
		embed.SetFooter("Powered by ticketsbot.net", AvatarUrl)
	}

	msg, err := session.ChannelMessageSendEmbed(ch, embed.MessageEmbed); if err != nil {
		log.Error(err.Error())
		return
	}

	if deleteAfter > 0 {
		DeleteAfter(SentMessage{session, msg}, deleteAfter)
	}
}

func DeleteAfter(msg SentMessage, secs int) {
	go func() {
		time.Sleep(time.Duration(secs) * time.Second)
		if err := msg.Session.ChannelMessageDelete(msg.Message.ChannelID, msg.Message.ID); err != nil {
			log.Error(err.Error())
		}
	}()
}

func ReactWithCheck(session *discordgo.Session, msg *discordgo.Message) {
	if err := session.MessageReactionAdd(msg.ChannelID, msg.ID, "✅"); err != nil {
		log.Error(err.Error())
	}
}

func ReactWithCross(session *discordgo.Session, msg discordgo.Message) {
	if err := session.MessageReactionAdd(msg.ChannelID, msg.ID, "❌"); err != nil {
		log.Error(err.Error())
	}
}
