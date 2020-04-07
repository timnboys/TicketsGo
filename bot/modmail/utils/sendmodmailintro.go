package utils

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strings"
)

var Emojis = map[int]string{
	1: "1️⃣",
	2: "2️⃣",
	3: "3️⃣",
	4: "4️⃣",
	5: "5️⃣",
	6: "6️⃣",
	7: "7️⃣",
	8: "8️⃣",
	9: "9️⃣",
}

func SendModMailIntro(ctx utils.CommandContext, dmChannelId uint64) {
	guilds := GetMutualGuilds(ctx.Shard, ctx.User.Id)

	message := "```fix\n"
	for i, guild := range guilds {
		message += fmt.Sprintf("%d) %s\n", i + 1, guild.Name)
	}

	if len(guilds) == 0 {
		message += "You do not have any mutual guilds with Tickets"
	}

	message = strings.TrimSuffix(message, "\n")
	message += "```\nRespond with the ID of the server you want to open a ticket in, or react to this message"

	// Create embed
	messageEmbed := embed.NewEmbed().
		SetColor(int(utils.Green)).
		SetTitle("Help").
		SetDescription(message)

	// Send message
	msg, err := ctx.Shard.CreateMessageEmbed(dmChannelId, messageEmbed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// Apply reactions
	max := len(guilds)
	if max > 9 {
		max = 9
	}

	if len(guilds) > 0 {
		for i := 1; i <= max; i++ {
			if err := ctx.Shard.CreateReaction(dmChannelId, msg.Id, Emojis[i]); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}
	}
}
