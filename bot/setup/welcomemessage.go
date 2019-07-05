package setup

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

type WelcomeMessageStage struct {
}

func (WelcomeMessageStage) State() State {
	return WelcomeMessage
}

func (WelcomeMessageStage) Prompt() string {
	return "Type the message that should be sent by the bot when a ticket channel is opened"
}

func (WelcomeMessageStage) Default() string {
	return "No message specified"
}

func (WelcomeMessageStage) Process(session *discordgo.Session, msg discordgo.Message) {
	guild, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	go database.SetWelcomeMessage(guild, msg.Content)
	utils.ReactWithCheck(session, &msg)
}
