package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnUserUpdate(s *gateway.Shard, e *events.UserUpdate) {
	go database.UpdateUser(e.Id, e.Username, e.Discriminator, e.Avatar)
}

func OnUserJoin(s *gateway.Shard, e *events.GuildMemberAdd) {
	go database.UpdateUser(e.GuildId, e.User.Username, e.User.Discriminator, e.User.Avatar)
}

func OnGuildCreateUserData(_ *gateway.Shard, e *events.GuildCreate) {
	data := make([]database.UserData, 0)

	for _, member := range e.Members {
		data = append(data, database.UserData{
			UserId:        member.User.Id,
			Username:      member.User.Username,
			Discriminator: member.User.Discriminator,
			Avatar:        member.User.Avatar,
		})
	}

	go database.InsertUsers(data)
}
