package setup

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/robfig/go-cache"
	"time"
)

var setupCache = cache.New(2 * time.Minute, 5 * time.Minute)

func (u *SetupUser) InSetup() bool {
	_, ok := setupCache.Get(u.ToString())
	return ok
}

func (u *SetupUser) GetState() *State {
	raw, ok := setupCache.Get(u.ToString()); if !ok {
		return nil
	}

	state := raw.(State)
	return &state
}

func (u *SetupUser) Next() {
	state := u.GetState()

	var newState State
	if state == nil {
		newState = State(0)
	} else {
		id := int(*state) + 1

		if id > GetMaxStage() {
			u.Finish()
			return
		}

		newState = State(id)
	}

	setupCache.Set(u.ToString(), newState, 2 * time.Minute)
}

func (u *SetupUser) Finish() {
	setupCache.Delete(u.ToString())

	msg := "The setup has been complete!\n" +
		"You can add / remove support staff / admins using:\n" +
		"`t!addadmin @User`\n" +
		"`t!removadmin @User`\n" +
		"`t!addsupport @User`\n" +
		"`t!removesupprt @User`\n" +
		"You can access more settings on the web panel at <https://panel.ticketsbot.net>"

	// Psuedo-premium
	utils.SendEmbed(u.Session, u.Channel, utils.Green, "Setup", msg, 30, true)
}

func (u *SetupUser) Cancel() {
	setupCache.Delete(u.ToString())
}
