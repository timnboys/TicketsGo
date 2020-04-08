package setup

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
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
	return "Specify the maximum amount of tickets that a user should be able to have open at once"
}

// This is not used
func (TicketLimitStage) Default() string {
	return "5"
}

func (TicketLimitStage) Process(shard *gateway.Shard, msg message.Message) {
	amountRaw := strings.Split(msg.Content, " ")[0]
	amount, err := strconv.Atoi(amountRaw)
	if err != nil {
		amount = 5
		utils.SendEmbed(shard, msg.ChannelId, utils.Red, "Error", fmt.Sprintf("Error: `%s`\nDefault to `%d`", err.Error(), amount), 1, true)
		utils.ReactWithCross(shard, msg.MessageReference)
	} else {
		utils.ReactWithCheck(shard, msg.MessageReference)
	}

	go database.SetTicketLimit(msg.GuildId, amount)
}
