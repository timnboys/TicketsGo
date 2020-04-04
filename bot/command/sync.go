package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
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
	guildStr := strconv.FormatUint(ctx.Guild.Id, 10)

	if !utils.IsBotAdmin(ctx.User.Id) && !utils.IsBotHelper(ctx.User.Id) {
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

	// Process any deleted cached channels
	//ctx.SendMessage("Scanning for deleted cached channels...")
	//processDeletedCachedChannels(ctx)
	//ctx.SendMessage("Completed synchronisation with cache")

	// Process any new channels that must be cached
	ctx.SendMessage("Recaching channels...")
	recacheChannels(ctx)
	ctx.SendMessage("Completed synchronisation with cache")

	// Check any panels still exist
	ctx.SendMessage("Scanning for deleted panels...")
	processDeletedPanels(ctx)
	ctx.SendMessage("Completed panel state synchronisation")
}

func processDeletedTickets(ctx utils.CommandContext, res chan int) {
	updated := 0

	tickets := make(chan []*uint64)
	go database.GetOpenTicketChannelIds(ctx.Guild.Id, tickets)
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
	go database.GetPanelsByGuild(ctx.Guild.Id, panels)

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

func processDeletedCachedChannels(ctx utils.CommandContext) {
	// Get all cached channels for the guild
	cachedChannelsChan := make(chan []database.Channel)
	go database.GetCachedChannelsByGuild(ctx.Guild.Id, cachedChannelsChan)
	cachedChannels := <-cachedChannelsChan

	// Get current guild channels
	channels, err := ctx.Shard.GetGuildChannels(ctx.Guild.Id); if err != nil {
		sentry.LogWithContext(err, ctx.ToErrorContext())
		return
	}

	// Make a duplicate slice of cached channels which we will remove IDs from as we go - remaining IDs have been deleted
	toRemove := make([]database.Channel, 0)
	// Prevent panic
	if len(cachedChannels) > 0 {
		toRemove = append(cachedChannels[:0:0], cachedChannels...)
	}

	for _, existingChannel := range channels {
		// Remove from toRemove slice & find cached object
		var index = -1
		var cached *database.Channel = nil // Since this a pointer, we must perform operations before removing from slice
		for i, cachedChannel := range toRemove {
			if cachedChannel.ChannelId == existingChannel.Id {
				index = i
				cached = &cachedChannel
			}
		}

		// Check that the
		if cached != nil {
			if cached.Name != existingChannel.Name { // Name is the only property that can be updated
				go database.StoreChannel(cached.ChannelId, cached.GuildId, existingChannel.Name, cached.Type)
			}
		}

		// If index = -1, we haven't cached the channel before
		if index != -1 {
			go database.StoreChannel(existingChannel.Id, ctx.Guild.Id, existingChannel.Name, int(existingChannel.Type))
		} else { // Else, we can remove the channel from the toRemove array
			toRemove = append(toRemove[:index], toRemove[index+1:]...)
		}
	}

	// Now we must remove from the cache any deleted channels
	for _, channel := range toRemove {
		go database.DeleteChannel(channel.ChannelId)
	}
}

func recacheChannels(ctx utils.CommandContext) {
	// Delete current cache, sync
	database.DeleteAllChannelsByGuild(ctx.Guild.Id)

	// Get refreshed channel objects from Discord
	raw, err := ctx.Shard.GetGuildChannels(ctx.Guild.Id); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	channels := make([]database.Channel, 0)
	for _, channel := range raw {
		channels = append(channels, database.Channel{
			ChannelId: channel.Id,
			GuildId:   ctx.Guild.Id,
			Name:      channel.Name,
			Type:      int(channel.Type),
		})
	}

	go database.InsertChannels(channels)
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
