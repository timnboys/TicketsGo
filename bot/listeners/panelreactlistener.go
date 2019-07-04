package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnPanelReact(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	msgId, err := strconv.ParseInt(e.MessageID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	isPanel := make(chan bool)
	go database.IsPanel(msgId, isPanel)
	if <-isPanel {
		user, err := s.User(e.UserID); if err != nil {
			log.Error(err.Error())
			return
		}

		if user.Bot {
			return
		}

		if err = s.MessageReactionRemove(e.ChannelID, e.MessageID, "📩", e.UserID); err != nil {
			log.Error(err.Error())
		}

		msg, err := s.ChannelMessage(e.ChannelID, e.MessageID); if err != nil {
			log.Error(err.Error())
			return
		}

		isPremium := make(chan bool)
		go utils.IsPremiumGuild(e.GuildID, isPremium)

		ctx := command.CommandContext{
			Session: s,
			User: *user,
			Guild: e.GuildID,
			Channel: e.ChannelID,
			Message: *msg,
			Root: "new",
			Args: make([]string, 0),
			IsPremium: <-isPremium,
			ShouldReact: false,
		}

		command.OpenCommand{}.Execute(ctx)
	}
}
