package listeners

import (
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

// Remove user permissions when they leave
func OnMemberLeave(s *gateway.Shard, e *events.GuildMemberRemove) {
	if err := database.Client.Permissions.RemoveSupport(e.GuildId, e.User.Id); err != nil {
		sentry.Error(err)
	}
}