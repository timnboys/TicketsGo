package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// Remove user permissions when they leave
func OnMemberLeave(s *gateway.Shard, e *events.GuildMemberRemove) {
	go database.RemoveSupport(e.GuildId, e.User.Id)
}