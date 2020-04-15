package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
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

var cooldown = cache.New(time.Minute * 5, time.Minute * 1)

func (SyncCommand) Execute(ctx utils.CommandContext) {
	guildStr := strconv.FormatUint(ctx.GuildId, 10)

	if !utils.IsBotAdmin(ctx.Author.Id) && !utils.IsBotHelper(ctx.Author.Id) {
		cooldownEnd, ok := cooldown.Get(guildStr)
		if ok && cooldownEnd.(int64) > time.Now().UnixNano() { // Expiry search only runs once a minute
			ctx.SendEmbed(utils.Red, "Sync", "This command is currently in cooldown")
			return
		}

		cooldown.Set(guildStr, time.Now().Add(time.Minute*5).UnixNano(), time.Minute*5)
	}

	// Process deleted tickets
	ctx.SendMessage("Scanning for deleted ticket channels...")
	updated := make(chan int)
	go processDeletedTickets(ctx, updated)
	ctx.SendMessage(fmt.Sprintf("Completed **%d** ticket state synchronisation", <-updated))

	// Check any panels still exist
	ctx.SendMessage("Scanning for deleted panels...")
	processDeletedPanels(ctx)
	ctx.SendMessage("Completed panel state synchronisation")
}

func processDeletedTickets(ctx utils.CommandContext, res chan int) {
	updated := 0

	tickets := make(chan []*uint64)
	go database.GetOpenTicketChannelIds(ctx.GuildId, tickets)
	for _, channel := range <-tickets {
		if channel == nil {
			continue
		}

		_, err := ctx.Shard.GetChannel(*channel)
		if err != nil { // An admin has deleted the channel manually
			updated++
			go database.CloseByChannel(*channel)
		}
	}

	res <-updated
}

func processDeletedPanels(ctx utils.CommandContext) {
	panels := make(chan []database.Panel)
	go database.GetPanelsByGuild(ctx.GuildId, panels)

	for _, panel := range <-panels {
		// Pre-channel ID logging panel - we'll just leave it for now.
		if panel.ChannelId == 0 {
			continue
		}

		// Check cache first to prevent extra requests to discord
		if _, err := ctx.Shard.GetChannelMessage(panel.ChannelId, panel.MessageId); err != nil {
			// Message no longer exists
			go database.DeletePanel(panel.MessageId)
		}
	}
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

func (SyncCommand) Category() Category {
	return Settings
}

func (SyncCommand) AdminOnly() bool {
	return false
}

func (SyncCommand) HelperOnly() bool {
	return false
}
