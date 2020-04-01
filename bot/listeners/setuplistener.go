package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/setup"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnSetupProgress(s *gateway.Shard, e *events.MessageCreate) {
	u := setup.SetupUser{
		Guild:   e.GuildId,
		User:    e.Author.Id,
		Channel: e.ChannelId,
		Session: s,
	}

	if u.InSetup() {
		// Process current stage
		u.GetState().Process(s, *e.Message)

		// Start next stage
		u.Next()
		state := u.GetState()
		if state != nil {
			stage := state.GetStage()
			if stage != nil {
				// Psuedo-premium
				utils.SendEmbed(s, e.ChannelId, utils.Green, "Setup", (*stage).Prompt(), 120, true)
			}
		}
	}
}
