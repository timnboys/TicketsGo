package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/setup"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/bwmarrin/discordgo"
)

func OnSetupProgress(s *discordgo.Session, e *discordgo.MessageCreate) {
	u := setup.SetupUser{
		Guild:   e.GuildID,
		User:    e.Author.ID,
		Channel: e.ChannelID,
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
				utils.SendEmbed(s, e.ChannelID, utils.Green, "Setup", (*stage).Prompt(), 120, true)
			}
		}
	}
}
