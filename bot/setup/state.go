package setup

import "github.com/bwmarrin/discordgo"

type State int

const (
	Prefix State = iota
	WelcomeMessage
	TicketLimit
	ChannelCategory
	ArchiveChannel
)

var stages = []Stage{
	PrefixStage{},
}

func (s *State) GetStage() *Stage {
	for _, stage := range stages {
		if stage.State() == *s {
			return &stage
		}
	}
	return nil
}

func (s *State) Process(session *discordgo.Session, msg discordgo.Message) {
	stage := s.GetStage(); if stage == nil {
		return
	}

	(*stage).Process(session, msg)
}

func GetMaxStage() int {
	max := 0

	for _, stage := range stages {
		if int(stage.State()) > max {
			max = int(stage.State())
		}
	}

	return max
}
