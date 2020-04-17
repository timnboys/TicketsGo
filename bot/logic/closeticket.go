package logic

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/archive"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel/embed"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/rest"
	"strconv"
	"strings"
)

func CloseTicket(s *gateway.Shard, guildId, channelId, messageId uint64, member member.Member, args []string, fromReaction, isPremium bool) {
	reference := message.MessageReference{
		MessageId: messageId,
		ChannelId: channelId,
		GuildId:   guildId,
	}

	// Verify this is a ticket or modmail channel
	isTicketChan := make(chan bool)
	go database.IsTicketChannel(channelId, isTicketChan)
	isTicket := <-isTicketChan

	// Cannot happen if fromReaction
	if !isTicket {
		utils.ReactWithCross(s, reference)
		utils.SendEmbed(s, channelId, utils.Red, "Error", "This is not a ticket channel", nil, 30, isPremium)

		return
	}

	// Create reason
	var reason string
	silentClose := false
	for _, arg := range args {
		if arg == "--silent" {
			silentClose = true
		} else {
			reason += fmt.Sprintf("%s ", arg)
		}
	}
	reason = strings.TrimSuffix(reason, " ")

	// Check the user is permitted to close the ticket
	permissionLevelChan := make(chan utils.PermissionLevel)
	go utils.GetPermissionLevel(s, member, guildId, permissionLevelChan)
	permissionLevel := <-permissionLevelChan

	// Get ticket struct
	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(channelId, ticketChan)
	ticket := <-ticketChan

	usersCanCloseChan := make(chan bool)
	go database.IsUserCanClose(guildId, usersCanCloseChan)
	usersCanClose := <-usersCanCloseChan

	if (permissionLevel == utils.Everyone && ticket.Owner != member.User.Id) || (permissionLevel == utils.Everyone && !usersCanClose) {
		if !fromReaction {
			utils.ReactWithCross(s, reference)
			utils.SendEmbed(s, channelId, utils.Red, "Error", "You are not permitted to close this ticket", nil, 30, isPremium)
		}
		return
	}

	// TODO: Re-add permission check
	/*if !permission.HasPermissions(s, guildId, s.SelfId(), permission.ManageChannels) {
		ctx.ReactWithCross()
		ctx.SendEmbed(utils.Red, "Error", "I do not have permission to delete this channel")
		return
	}*/

	if !fromReaction {
		utils.ReactWithCheck(s, reference)
	}

	// Archive
	msgs := make([]message.Message, 0)

	lastId := uint64(0)
	count := -1
	for count != 0 {
		array, err := s.GetChannelMessages(channelId, rest.GetChannelMessagesData{
			Before: lastId,
			Limit:  100,
		})

		count = len(array)
		if err != nil {
			count = 0
			sentry.Error(err)
		}

		if count > 0 {
			lastId = array[len(array)-1].Id

			msgs = append(msgs, array...)
		}
	}

	// Reverse messages
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}

	if err := archive.ArchiverClient.Store(msgs, guildId, ticket.Id, isPremium); err != nil {
		sentry.Error(err)
	}

	// Set ticket state as closed and delete channel
	go database.Close(guildId, ticket.Id)
	if _, err := s.DeleteChannel(channelId); err != nil {
		sentry.Error(err)
	}

	// Send logs to archive channel
	archiveChannelChan := make(chan uint64)
	go database.GetArchiveChannel(guildId, archiveChannelChan)
	archiveChannelId := <-archiveChannelChan

	channelExists := true
	if _, err := s.GetChannel(archiveChannelId); err != nil {
		channelExists = false
	}

	// Save space - delete the webhook
	go database.DeleteWebhookByUuid(ticket.Uuid)

	if channelExists {
		embed := embed.NewEmbed().
			SetTitle("Ticket Closed").
			SetColor(int(utils.Green)).
			AddField("Ticket ID", strconv.Itoa(ticket.Id), true).
			AddField("Closed By", member.User.Mention(), true).
			AddField("Archive", fmt.Sprintf("[Click here](https://panel.ticketsbot.net/manage/%d/logs/view/%d)", guildId, ticket.Id), true).
			AddField("Reason", reason, false)

		if _, err := s.CreateMessageEmbed(archiveChannelId, embed); err != nil {
			sentry.Error(err)
		}

		msg := fmt.Sprintf("Archive of `#ticket-%d` (closed by %s#%s)", ticket.Id, member.User.Username, utils.PadDiscriminator(member.User.Discriminator))
		if reason != "" {
			msg += fmt.Sprintf(" with reason `%s`", reason)
		}

		// Notify user and send logs in DMs
		if !silentClose {
			dmChannel, err := s.CreateDM(ticket.Owner)

			// Only send the msg if we could create the channel
			if err == nil {
				if _, err := s.CreateMessageEmbed(dmChannel.Id, embed); err != nil {
					sentry.Error(err)
				}
			}
		}
	}
}
