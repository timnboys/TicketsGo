package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/patrickmn/go-cache"
	"github.com/rxdn/gdl/rest/request"
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
	ctx.SendMessage(fmt.Sprintf("Completed **%d** ticket state synchronisation(s)", processDeletedTickets(ctx)))

	// Check any panels still exist
	ctx.SendMessage("Scanning for deleted panels...")
	ctx.SendMessage(fmt.Sprintf("Completed **%d** panel state synchronisation(s)", processDeletedPanels(ctx)))

	ctx.SendMessage("Sync complete!")
}

func processDeletedTickets(ctx utils.CommandContext) (updated int) {
	tickets, err := database.Client.Tickets.GetGuildOpenTickets(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	for _, ticket := range tickets {
		if ticket.ChannelId == nil {
			continue
		}

		_, err := ctx.Shard.GetChannel(*ticket.ChannelId)
		if err != nil && err == request.ErrNotFound { // An admin has deleted the channel manually
			updated++

			go func() {
				if err := database.Client.Tickets.Close(ticket.Id, ticket.GuildId); err != nil {
					sentry.ErrorWithContext(err, ctx.ToErrorContext())
				}
			}()
		}
	}

	return
}

func processDeletedPanels(ctx utils.CommandContext) (removed int) {
	panels, err := database.Client.Panel.GetByGuild(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	for _, panel := range panels {
		// Pre-channel ID logging panel - we'll just leave it for now.
		if panel.ChannelId == 0 {
			continue
		}

		// Check cache first to prevent extra requests to discord
		if _, err := ctx.Shard.GetChannelMessage(panel.ChannelId, panel.MessageId); err != nil && err == request.ErrNotFound {
			removed++

			// Message no longer exists
			go func() {
				if err := database.Client.Panel.Delete(panel.MessageId); err != nil {
					sentry.ErrorWithContext(err, ctx.ToErrorContext())
				}
			}()
		}
	}

	return
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
