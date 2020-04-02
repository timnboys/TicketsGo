package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel/embed"
	"strconv"
)

// Reset
type AdminDebugCommand struct {
}

func (AdminDebugCommand) Name() string {
	return "debug"
}

func (AdminDebugCommand) Description() string {
	return "Provides debugging information"
}

func (AdminDebugCommand) Aliases() []string {
	return []string{}
}

func (AdminDebugCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

func (AdminDebugCommand) Execute(ctx utils.CommandContext) {
	// Get if SQL is connected
	sqlConnected := make(chan bool)
	go database.IsConnected(sqlConnected)

	// Get if Redis is connected
	redisConnected := make(chan bool)
	go cache.Client.IsConnected(redisConnected)

	// Get ticket category
	ticketCategoryChan := make(chan uint64)
	go database.GetCategory(ctx.Guild.Id, ticketCategoryChan)
	ticketCategoryId := <- ticketCategoryChan
	var ticketCategory string
	for _, channel := range ctx.Guild.Channels {
		if channel.Id == ticketCategoryId { // Don't need to compare channel types
			ticketCategory = channel.Name
		}
	}

	// Get owner
	invalidOwner := false
	owner, err := ctx.Shard.GetGuildMember(ctx.Guild.Id, ctx.Guild.OwnerId); if err != nil {
		invalidOwner = true
	}

	var ownerFormatted string
	if invalidOwner || owner == nil {
		ownerFormatted = strconv.FormatUint(ctx.Guild.OwnerId, 10)
	} else {
		ownerFormatted = fmt.Sprintf("%s#%d", owner.User.Username, owner.User.Discriminator)
	}

	// Get archive channel
	//archiveChannelChan := make(chan int64)
	//go database.GetArchiveChannel()

	embed := embed.NewEmbed().
		SetTitle("Admin").
		SetColor(int(utils.Green)).

		AddField("Shard", strconv.Itoa(ctx.Shard.ShardId), true).
		AddField("SQL Is Connected", strconv.FormatBool(<-sqlConnected), true).
		AddField("Redis Is Connected", strconv.FormatBool(<-redisConnected), true).

		AddField("Ticket Category", ticketCategory, true).
		AddField("Owner", ownerFormatted, true)

	msg, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: msg}, 30)
}

func (AdminDebugCommand) Parent() interface{} {
	return &AdminCommand{}
}

func (AdminDebugCommand) Children() []Command {
	return []Command{}
}

func (AdminDebugCommand) PremiumOnly() bool {
	return false
}

func (AdminDebugCommand) AdminOnly() bool {
	return false
}

func (AdminDebugCommand) HelperOnly() bool {
	return true
}
