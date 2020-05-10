package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
	"strconv"
	"strings"
)

type TicketLimitStage struct {
}

func (TicketLimitStage) State() State {
	return TicketLimit
}

func (TicketLimitStage) Prompt() string {
	return "Specify the maximum amount of tickets that a **single user** should be able to have open at once"
}

// This is not used
func (TicketLimitStage) Default() string {
	return "5"
}

func (TicketLimitStage) Process(shard *gateway.Shard, msg message.Message) {
	ref := message.MessageReference{
		MessageId: msg.Id,
		ChannelId: msg.ChannelId,
		GuildId:   msg.GuildId,
	}

	amountRaw := strings.Split(msg.Content, " ")[0]
	amount, err := strconv.Atoi(amountRaw)
	if err != nil {
		amount = 5
		utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("Error: `%s`\nDefaulting to `%d`", err.Error(), amount), nil, 30, true)
		utils.ReactWithCross(shard, ref)
	} else {
		utils.ReactWithCheck(shard, ref)
	}

	if err := database.Client.TicketLimit.Set(msg.GuildId, uint8(amount)); err == nil {
		utils.ReactWithCheck(shard, ref)
	} else {
		utils.ReactWithCross(shard, ref)
		sentry.Error(err)
	}
}
