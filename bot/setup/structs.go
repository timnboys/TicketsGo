package setup

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type SetupUser struct {
	Guild   string
	User    string
	Channel string
	Session *discordgo.Session
}

type Stage interface {
	State() State
	Prompt() string
	Default() string
	Process(session *discordgo.Session, msg discordgo.Message)
}

func (s *SetupUser) ToString() string {
	return fmt.Sprintf("%s-%s-%s", s.Guild, s.User, s.Channel)
}
