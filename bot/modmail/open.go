package modmail

import (
	"errors"
	"fmt"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	uuid "github.com/satori/go.uuid"
)

func OpenModMailTicket(shard *gateway.Shard, guild modmailutils.UserGuild, user *user.User) (uint64, error) {
	ticketId := uuid.NewV4()

	// If we're using a panel, then we need to create the ticket in the specified category
	categoryChan := make(chan uint64)
	go database.GetCategory(guild.Id, categoryChan)
	category := <-categoryChan

	// Make sure the category exists
	if category != 0 {
		if _, err := shard.GetChannel(category); err != nil {
			category = 0
		}
	}

	requiredPerms := []permission.Permission{
		permission.ManageChannels,
		permission.ManageRoles,
		permission.ViewChannel,
		permission.SendMessages,
		permission.ReadMessageHistory,
	}

	if !permission.HasPermissions(shard, guild.Id, shard.SelfId(), requiredPerms...) {
		return 0, errors.New("I do not have the correct permissions required to create the channel in the server")
	}

	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		ch, err := shard.GetChannel(category)
		if err != nil {
			useCategory = false
			go database.DeleteCategory(guild.Id)
			return 0, errors.New("Ticket category has been deleted")
		}

		if ch.Type != channel.ChannelTypeGuildCategory {
			useCategory = false
			go database.DeleteCategory(guild.Id)
			return 0, errors.New("Ticket category is not a ticket category")
		}

		if !permission.HasPermissionsChannel(shard, guild.Id, shard.SelfId(), category, requiredPerms...) {
			return 0, errors.New("I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions")
		}
	}

	if useCategory {
		channels, err := shard.GetGuildChannels(guild.Id); if err != nil {
			return 0, err
		}

		channelCount := 0
		for _, channel := range channels {
			if channel.ParentId == category {
				channelCount += 1
			}
		}

		if channelCount >= 50 {
			return 0, errors.New("There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category")
		}
	}

	// Create channel
	name := fmt.Sprintf("modmail-%s", user.Username)
	overwrites := createOverwrites(shard, guild.Id)

	data := rest.CreateChannelData{
		Name:                 name,
		Type:                 channel.ChannelTypeGuildText,
		PermissionOverwrites: overwrites,
		ParentId:             category, // If not using category, value will be 0 and omitempty
	}

	channel, err := shard.CreateGuildChannel(guild.Id, data)
	if err != nil {
		sentry.Error(err)
		return 0, err
	}

	// Create webhook
	go createWebhook(shard, guild.Id, channel.Id, ticketId.String())

	go modmaildatabase.CreateModMailSession(ticketId.String(), guild.Id, user.Id, channel.Id)
	return channel.Id, nil
}

func createWebhook(shard *gateway.Shard, guildId, channelId uint64, uuid string) {
	self := shard.Cache.GetSelf()
	if self == nil {
		return
	}

	if permission.HasPermissionsChannel(shard, guildId, channelId, self.Id, permission.ManageWebhooks) { // Do we actually need this?
		webhook, err := shard.CreateWebhook(channelId, rest.WebhookData{
			Username: self.Username,
			Avatar:   self.Avatar,
		})
		if err != nil {
			sentry.ErrorWithContext(err, sentry.ErrorContext{
				Guild:   guildId,
				Shard:   shard.ShardId,
				Command: "open",
			})
			return
		}

		formatted := fmt.Sprintf("%s/%s", webhook.Id, webhook.Token)

		ticketWebhook := database.TicketWebhook{
			Uuid:       uuid,
			WebhookUrl: formatted,
		}
		ticketWebhook.AddWebhook()
	}
}

func createOverwrites(shard *gateway.Shard, guildId uint64) []*channel.PermissionOverwrite {
	// Apply permission overwrites
	overwrites := make([]*channel.PermissionOverwrite, 0)
	overwrites = append(overwrites, &channel.PermissionOverwrite{ // @everyone
		Id:    guildId,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	// Create list of members & roles who should be added to the ticket
	allowedUsers := make([]uint64, 0)
	allowedRoles := make([]uint64, 0)

	// Get support reps
	supportUsers := make(chan []uint64)
	go database.GetSupport(guildId, supportUsers)
	for _, user := range <-supportUsers {
		allowedUsers = append(allowedUsers, user)
	}

	// Get support roles
	supportRoles := make(chan []uint64)
	go database.GetSupportRoles(guildId, supportRoles)
	for _, role := range <-supportRoles {
		allowedRoles = append(allowedRoles, role)
	}

	// Add ourselves
	allowedUsers = append(allowedUsers, shard.SelfId())

	for _, member := range allowedUsers {
		allow := []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == shard.SelfId() {
			allow = append(allow, permission.ManageWebhooks)
		}

		overwrites = append(overwrites, &channel.PermissionOverwrite{
			Id:    member,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(allow...),
			Deny:  0,
		})
	}

	for _, role := range allowedRoles {
		overwrites = append(overwrites, &channel.PermissionOverwrite{
			Id:    role,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
			Deny:  0,
		})
	}

	return overwrites
}
