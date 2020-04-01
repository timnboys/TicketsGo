package setup

import (
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
)

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
	WelcomeMessageStage{},
	TicketLimitStage{},
	ChannelCategoryStage{},
	ArchiveChannelStage{},
}

func (s *State) GetStage() *Stage {
	for _, stage := range stages {
		if stage.State() == *s {
			return &stage
		}
	}
	return nil
}

func (s *State) Process(shard *gateway.Shard, msg message.Message) {
	stage := s.GetStage(); if stage == nil {
		return
	}

	(*stage).Process(shard, msg)
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
