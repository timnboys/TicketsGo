package listeners

import (
	"github.com/bwmarrin/discordgo"
)

func OnReady(s *discordgo.Session, _ *discordgo.Ready) {
	s.State.TrackEmojis = false
	s.State.TrackPresences = false
	s.State.TrackVoice = false
	s.SyncEvents = false
}
