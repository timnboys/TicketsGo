package listeners

import (
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
	"github.com/rxdn/gdl/objects/user"
)

func OnReady(s *gateway.Shard, e *events.Ready) {
	_ = s.UpdateStatus(user.BuildStatus(user.ActivityTypePlaying, "DM for help | t!help"))
}
