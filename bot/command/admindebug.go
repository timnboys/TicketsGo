package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
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
	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		sentry.Error(err)
		return
	}

	guild, err := ctx.Session.State.Guild(ctx.Guild); if err != nil {
		// Not cached
		guild, err = ctx.Session.Guild(ctx.Guild); if err != nil {
			sentry.Error(err)
			return
		}
	}

	// Get if SQL is connected
	sqlConnected := make(chan bool)
	go database.IsConnected(sqlConnected)

	// Get if Redis is connected
	redisConnected := make(chan bool)
	go cache.Client.IsConnected(redisConnected)

	// Get ticket category
	ticketCategoryChan := make(chan int64)
	go database.GetCategory(guildId, ticketCategoryChan)
	ticketCategoryId := <- ticketCategoryChan
	var ticketCategory string
	for _, channel := range guild.Channels {
		if channel.ID == strconv.Itoa(int(ticketCategoryId)) { // Don't need to compare channel types
			ticketCategory = channel.Name
		}
	}

	// Get owner
	invalidOwner := false
	owner, err := ctx.Session.State.Member(guild.ID, guild.OwnerID); if err != nil {
		owner, err = ctx.Session.GuildMember(guild.ID, guild.OwnerID); if err != nil {
			invalidOwner = true
		}
	}

	var ownerFormatted string
	if invalidOwner {
		ownerFormatted = guild.OwnerID
	} else {
		ownerFormatted = fmt.Sprintf("%s#%s", owner.User.Username, owner.User.Discriminator)
	}

	// Get archive channel
	//archiveChannelChan := make(chan int64)
	//go database.GetArchiveChannel()

	embed := utils.NewEmbed().
		SetTitle("Admin").
		SetColor(int(utils.Green)).

		AddField("Shard", strconv.Itoa(ctx.Session.ShardID), true).
		AddField("SQL Is Connected", strconv.FormatBool(<-sqlConnected), true).
		AddField("Redis Is Connected", strconv.FormatBool(<-redisConnected), true).

		AddField("Ticket Category", ticketCategory, true).
		AddField("Owner", ownerFormatted, true).

		MessageEmbed

	msg, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed); if err != nil {
		sentry.Error(err)
		return
	}

	utils.DeleteAfter(utils.SentMessage{Session: ctx.Session, Message: msg}, 30)
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
