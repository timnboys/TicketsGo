package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"strconv"
)

type SyncCommand struct {
}

func (SyncCommand) Name() string {
	return "sync"
}

func (SyncCommand) Description() string {
	return "Syncs the bot's database to the channels - useful if you a Discord outage has taken place"
}

func (SyncCommand) Aliases() []string {
	return []string{}
}

func (SyncCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Admin
}

func (SyncCommand) Execute(ctx utils.CommandContext) {
	ctx.SendMessage("Scanning for deleted channels...")
	updated := 0

	tickets := make(chan []*int64)
	go database.GetOpenTicketChannelIds(ctx.GuildId, tickets)
	for _, channel := range <-tickets {
		if channel == nil {
			continue
		}

		_, err := ctx.Session.Channel(strconv.Itoa(int(*channel)))
		if err != nil { // An admin has deleted the channel manually
			updated++
			go database.CloseByChannel(*channel)
		}
	}

	ctx.SendMessage(fmt.Sprintf("Updated **%d** channels", updated))
}

func (SyncCommand) Parent() interface{} {
	return nil
}

func (SyncCommand) Children() []Command {
	return make([]Command, 0)
}

func (SyncCommand) PremiumOnly() bool {
	return false
}

func (SyncCommand) AdminOnly() bool {
	return false
}

func (SyncCommand) HelperOnly() bool {
	return false
}
