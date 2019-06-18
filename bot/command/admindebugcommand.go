package command

import (
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
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

func (AdminDebugCommand) Execute(ctx CommandContext) {
	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	guild, err := ctx.Session.State.Guild(ctx.Guild); if err != nil {
		// Not cached
		guild, err = ctx.Session.Guild(ctx.Guild); if err != nil {
			log.Error(err.Error())
			return
		}
	}

	// Get if SQL is connected
	sqlConnected := make(chan bool)
	go database.IsConnected(sqlConnected)

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

	// Get archive channel
	//archiveChannelChan := make(chan int64)
	//go database.Get

	embed := utils.NewEmbed().
		SetTitle("Admin").
		SetColor(int(utils.Green)).

		AddField("Shard", strconv.Itoa(ctx.Session.ShardID), true).
		AddField("SQL Is Connected", strconv.FormatBool(<-sqlConnected), true).
		AddField("Redis Is Connected", "false", true).

		AddField("Ticket Category", ticketCategory, true).
		//AddField()

		MessageEmbed

	msg, err := ctx.Session.ChannelMessageSendEmbed(ctx.Channel, embed); if err != nil {
		log.Error(err.Error())
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
