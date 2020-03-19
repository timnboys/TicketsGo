package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

type TicketLimitStage struct {
}

func (TicketLimitStage) State() State {
	return TicketLimit
}

func (TicketLimitStage) Prompt() string {
	return "Specify the maximum amount of tickets that a user should be able to have open at once"
}

// This is not used
func (TicketLimitStage) Default() string {
	return "5"
}

func (TicketLimitStage) Process(session *discordgo.Session, msg discordgo.Message) {
	guild, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildID,
			User:    msg.Author.ID,
			Channel: msg.ChannelID,
			Shard:   session.ShardID,
		})
		return
	}

	amountRaw := strings.Split(msg.Content, " ")[0]
	amount, err := strconv.Atoi(amountRaw)
	if err != nil {
		amount = 5
		utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", fmt.Sprintf("Error: `%s`\nDefault to `%d`", err.Error(), amount), 1, true)
		utils.ReactWithCross(session, msg)
	} else {
		utils.ReactWithCheck(session, &msg)
	}

	go database.SetTicketLimit(guild, amount)
}
