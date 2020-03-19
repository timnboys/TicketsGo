package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

type PrefixStage struct {
}

func (PrefixStage) State() State {
	return Prefix
}

func (PrefixStage) Prompt() string {
	return "Type the prefix that you would like to use for the bot" +
		"\nThe prefix is the characters that come *before* the command (excluding the actual command itself)" +
		"\nExample: `t!`"
}

func (PrefixStage) Default() string {
	return config.Conf.Bot.Prefix
}

func (PrefixStage) Process(session *discordgo.Session, msg discordgo.Message) {
	if len(msg.Content) > 8 {
		utils.SendEmbed(session, msg.ChannelID, utils.Red, "Error", fmt.Sprintf("The maxium prefix langeth is 8 characters\nDefaulting to `%s`", PrefixStage{}.Default()), 15, true)
		return
	}

	guild, err := strconv.ParseInt(msg.GuildID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild:   msg.GuildID,
			User:    msg.Author.ID,
			Channel: msg.ChannelID,
			Shard:   session.ShardID,
		})
		return
	}

	go database.SetPrefix(guild, msg.Content)
	utils.ReactWithCheck(session, &msg)
}
