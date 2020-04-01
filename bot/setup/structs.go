package setup

import (
	"fmt"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/message"
)

type SetupUser struct {
	Guild   uint64
	User    uint64
	Channel uint64
	Session *gateway.Shard
}

type Stage interface {
	State() State
	Prompt() string
	Default() string
	Process(shard *gateway.Shard, msg message.Message)
}

func (s *SetupUser) ToString() string {
	return fmt.Sprintf("%d-%d-%d", s.Guild, s.User, s.Channel)
}
