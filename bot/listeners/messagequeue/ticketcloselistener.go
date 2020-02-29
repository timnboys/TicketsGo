package messagequeue

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/jonas747/dshardmanager"
	"strconv"
	"strings"
)

func ListenTicketClose(shardManager *dshardmanager.Manager) {
	closes := make(chan cache.TicketCloseMessage)
	go cache.Client.ListenTicketClose(closes)

	for payload := range closes {
		// Get the ticket properties
		ticketChan := make(chan database.Ticket)
		go database.GetTicketByUuid(payload.Uuid, ticketChan)
		ticket := <-ticketChan

		// Check that this is a valid ticket
		if ticket.Uuid == "" {
			return
		}

		// Get session
		s := shardManager.SessionForGuild(ticket.Guild)
		if s == nil { // Not on this cluster
			continue
		}

		guildIdStr := strconv.Itoa(int(ticket.Guild))

		// Create error context for later
		errorContext := sentry.ErrorContext{
			Guild:   guildIdStr,
			User:    strconv.Itoa(int(payload.User)),
			Shard:   s.ShardID,
		}

		// Get guild obj
		guild, err := s.State.Guild(guildIdStr)
		if err != nil {
			guild, err = s.Guild(guildIdStr)
			if err != nil {
				sentry.ErrorWithContext(err, errorContext)
				return
			}
		}

		// Get whether the guild is premium
		// TODO: Check whether we actually need this
		isPremium := make(chan bool)
		go utils.IsPremiumGuild(utils.CommandContext{
			Session: s,
			GuildId: ticket.Guild,
			Guild:   guild,
		}, isPremium)

		// Get the member object
		userIdStr := strconv.Itoa(int(payload.User))
		member, err := s.State.Member(guildIdStr, userIdStr)
		if err != nil {
			member, err = s.GuildMember(guildIdStr, userIdStr)
			if err != nil {
				sentry.LogWithContext(err, errorContext)
				return
			}
		}

		// Add reason to args
		reason := strings.Split(payload.Reason, " ")

		ctx := utils.CommandContext{
			Session:     s,
			User:        *member.User,
			UserID:      payload.User,
			Guild:       guild,
			GuildId:     ticket.Guild,
			Channel:     strconv.Itoa(int(*ticket.Channel)),
			ChannelId:   *ticket.Channel,
			MessageId:   0,
			Root:        "close",
			Args:        reason,
			IsPremium:   <-isPremium,
			ShouldReact: false,
			Member:      member,
		}

		go command.CloseCommand{}.Execute(ctx)
	}
}
