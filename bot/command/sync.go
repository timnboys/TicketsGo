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
	cooldownEnd, ok := cooldown.Get(ctx.Guild.ID)
	if ok && cooldownEnd.(int64) > time.Now().UnixNano() { // Expiry search only runs once a minute
		ctx.SendEmbed(utils.Red, "Sync", "This command is currently in cooldown")
		return
	}

	cooldown.Set(ctx.Guild.ID, time.Now().Add(time.Minute * 5).UnixNano(), time.Minute * 5)

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
		channelId := strconv.Itoa(int(panel.ChannelId))
		msgId := strconv.Itoa(int(panel.MessageId))
		if _, err := ctx.Session.State.Message(channelId, msgId); err != nil {
			if _, err := ctx.Session.ChannelMessage(channelId, msgId); err != nil {
				// Message no longer exists
				go database.DeletePanel(panel.MessageId)
			}
		}
	}
}

func processDeletedCachedChannels(ctx utils.CommandContext) {
	// Get all cached channels for the guild
	cachedChannelsChan := make(chan []database.Channel)
	go database.GetCachedChannelsByGuild(ctx.GuildId, cachedChannelsChan)
	cachedChannels := <-cachedChannelsChan

	// Get current guild channels
	channels, err := ctx.Session.GuildChannels(ctx.Guild.ID); if err != nil {
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
			if strconv.Itoa(int(cachedChannel.ChannelId)) == existingChannel.ID {
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
			channelId, err := strconv.ParseInt(existingChannel.ID, 10, 64); if err != nil {
				sentry.LogWithContext(err, ctx.ToErrorContext())
				continue
			}

			go database.StoreChannel(channelId, ctx.GuildId, existingChannel.Name, int(existingChannel.Type))
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
	// Delete current cache
	go database.DeleteAllChannelsByGuild(ctx.GuildId)

	// Get refreshed channel objects from Discord
	raw, err := ctx.Session.GuildChannels(ctx.Guild.ID); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	channels := make([]database.Channel, 0)
	for _, channel := range raw {
		channelId, err := strconv.ParseInt(channel.ID, 10, 64); if err != nil {
			sentry.LogWithContext(err, ctx.ToErrorContext())
			continue
		}

		channels = append(channels, database.Channel{
			ChannelId: channelId,
			GuildId:   ctx.GuildId,
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

func (SyncCommand) AdminOnly() bool {
	return false
}

func (SyncCommand) HelperOnly() bool {
	return false
}
