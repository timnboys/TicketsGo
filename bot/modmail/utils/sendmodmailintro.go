package utils

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strings"
)

func SendModMailIntro(ctx utils.CommandContext) {
	guildsChan := make(chan []UserGuild)
	go GetMutualGuilds(ctx.UserID, guildsChan)
	guilds := <-guildsChan

	message := "```css\n"
	for i, guild := range guilds {
		message += fmt.Sprintf("**%d)** %s\n", i + 1, guild.Name)
	}

	message = strings.TrimSuffix(message, "\n")
	message += "```\nRespond with the ID of the server you want to open a ticket in, or react to this message"

	// Create embed
	embed := utils.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Help").
		SetDescription(message)

	// Send message
	_, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed.MessageEmbed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}
}
