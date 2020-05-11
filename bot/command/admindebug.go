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
	// Get if Redis is connected
	redisConnected := make(chan bool)
	go cache.Client.IsConnected(redisConnected)

	// Get ticket category
	categoryId, err := database.Client.ChannelCategory.Get(ctx.GuildId)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// get guild channels
	channels, err := ctx.Shard.GetGuildChannels(ctx.GuildId); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	var categoryName string
	for _, channel := range channels {
		if channel.Id == categoryId { // Don't need to compare channel types
			categoryName = channel.Name
		}
	}

	if categoryName == "" {
		categoryName = "None"
	}

	// get guild object
	guild, err := ctx.Guild(); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
	}

	// Get owner
	invalidOwner := false
	owner, err := ctx.Shard.GetGuildMember(ctx.GuildId, guild.OwnerId); if err != nil {
		invalidOwner = true
	}

	var ownerFormatted string
	if invalidOwner {
		ownerFormatted = strconv.FormatUint(guild.OwnerId, 10)
	} else {
		ownerFormatted = fmt.Sprintf("%s#%s", owner.User.Username, utils.PadDiscriminator(owner.User.Discriminator))
	}

	// Get archive channel
	//archiveChannelChan := make(chan int64)
	//go database.GetArchiveChannel()

	embed := embed.NewEmbed().
		SetTitle("Admin").
		SetColor(int(utils.Green)).

		AddField("Shard", strconv.Itoa(ctx.Shard.ShardId), true).
		AddField("Redis Is Connected", strconv.FormatBool(<-redisConnected), true).
		AddBlankField(false).

		AddField("Ticket Category", categoryName, true).
		AddField("Owner", ownerFormatted, true)

	msg, err := ctx.Shard.CreateMessageEmbed(ctx.ChannelId, embed); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	utils.DeleteAfter(utils.SentMessage{Shard: ctx.Shard, Message: &msg}, 30)
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

func (AdminDebugCommand) Category() Category {
	return Settings
}

func (AdminDebugCommand) AdminOnly() bool {
	return false
}

func (AdminDebugCommand) HelperOnly() bool {
	return true
}
