package logic

import (
	"context"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	dbclient "github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/TicketsBot/database"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	"golang.org/x/sync/errgroup"
	"strings"
	"time"
)

// if panel != nil, msg should be artifically filled, excluding the message ID
func OpenTicket(s *gateway.Shard, user user.User, msg message.MessageReference, isPremium bool, args []string, panel *database.Panel) {
	// If we're using a panel, then we need to create the ticket in the specified category

	var category uint64
	if panel != nil && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		var err error
		category, err = dbclient.Client.ChannelCategory.Get(msg.GuildId); if err != nil {
			sentry.Error(err)
		}
	}

	// TODO: Re-add permission check
	/*requiredPerms := []permission.Permission{
		permission.ManageChannels,
		permission.ManageRoles,
		permission.ViewChannel,
		permission.SendMessages,
		permission.ReadMessageHistory,
	}

	if !permission.HasPermissions(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), requiredPerms...) {
		ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
		if ctx.ShouldReact {
			ctx.ReactWithCross()
		}
		return
	}*/

	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		_, err := s.GetChannel(category)
		if err != nil {
			useCategory = false
			//go database.DeleteCategory(ctx.GuildId) TODO: Could this be due to a Discord outage? Check specifically for a 404
		} else {
			// TODO: Re-add permission check
			/*if !permission.HasPermissionsChannel(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), category, requiredPerms...) {
				ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
				if ctx.ShouldReact {
					ctx.ReactWithCross()
				}
				return
			}*/
		}
	}

	// create DM channel
	dmChannel, err := s.CreateDM(user.Id)

	// target channel for messaging the user
	// either DMs or the channel where the command was run
	var targetChannel uint64
	if panel == nil {
		targetChannel = msg.ChannelId
	} else {
		targetChannel = dmChannel.Id
	}

	// Make sure ticket count is within ticket limit
	violatesTicketLimit, limit := getTicketLimit(msg.GuildId, user.Id)
	if violatesTicketLimit {
		// Notify the user
		if targetChannel != 0 {
			ticketsPluralised := "ticket"
			if limit > 1 {
				ticketsPluralised += "s"
			}
			content := fmt.Sprintf("You are only able to open %d %s at once", limit, ticketsPluralised)
			utils.SendEmbed(s, targetChannel, utils.Red, "Error", content, nil, 30, isPremium)
		}

		return
	}

	// Generate subject
	subject := "No subject given"
	if panel != nil && panel.Title != "" { // If we're using a panel, use the panel title as the subject
		subject = panel.Title
	} else { // Else, take command args as the subject
		if len(args) > 0 {
			subject = strings.Join(args, " ")
		}
		if len(subject) > 256 {
			subject = subject[0:255]
		}
	}

	// Make sure there's not > 50 channels in a category
	if useCategory {
		channels, _ := s.GetGuildChannels(msg.GuildId)

		channelCount := 0
		for _, channel := range channels {
			if channel.ParentId == category {
				channelCount++
			}
		}

		if channelCount >= 50 {
			utils.SendEmbed(s, msg.ChannelId, utils.Red, "Error", "There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category", nil, 30, isPremium)
			return
		}
	}

	if panel == nil {
		utils.ReactWithCheck(s, msg)
	}

	// Create channel
	id, err := dbclient.Client.Tickets.Create(msg.GuildId, user.Id); if err != nil {
		sentry.Error(err)
		return
	}

	overwrites := CreateOverwrites(msg.GuildId, user.Id, s.SelfId(),panel.MessageId)

	// Create ticket name
	var name string

	namingScheme, err := dbclient.Client.NamingScheme.Get(msg.GuildId)
	if err != nil {
		namingScheme = database.Id
		sentry.Error(err)
	}

	if namingScheme == database.Username {
		name = fmt.Sprintf("ticket-%s", user.Username)
	} else {
		name = fmt.Sprintf("ticket-%d", id)
	}

	data := rest.CreateChannelData{
		Name:                 name,
		Type:                 channel.ChannelTypeGuildText,
		Topic:                subject,
		PermissionOverwrites: overwrites,
		ParentId:             category,
	}
	if useCategory {
		data.ParentId = category
	}

	channel, err := s.CreateGuildChannel(msg.GuildId, data)
	if err != nil { // Bot likely doesn't have permission
		sentry.Error(err)

		// To prevent tickets getting in a glitched state, we should mark it as closed (or delete it completely?)
		if err := dbclient.Client.Tickets.Close(id, msg.GuildId); err != nil {
			sentry.Error(err)
		}

		return
	}

	welcomeMessageId := sendWelcomeMessage(s, msg.GuildId, channel.Id, user.Id, isPremium, subject)

	// UpdateUser channel in DB
	go func() {
		if err := dbclient.Client.Tickets.SetTicketProperties(msg.GuildId, id, channel.Id, welcomeMessageId); err != nil {
			sentry.Error(err)
		}
	}()

	// Ping @everyone
	pingEveryone, err := dbclient.Client.PingEveryone.Get(msg.GuildId); if err != nil {
		sentry.Error(err)
	}

	if pingEveryone {
		pingMessage, err := s.CreateMessageComplex(channel.Id, rest.CreateMessageData{
			Content:         "@everyone",
			AllowedMentions: message.MentionEveryone,
		})

		if err != nil {
			sentry.Error(err)
		} else {
			// error is likely to be a permission error
			_ = s.DeleteMessage(channel.Id, pingMessage.Id)
		}
	}
	
	// Ping @rolementions
        pingRoles, err := dbclient.Client.PanelRoleMentions.GetRoles(panel.MessageId)
        if err != nil {
                sentry.Error(err)
        }

        //if pingRoles {
                for _, mention := range pingRoles {
                str := strconv.FormatUint(mention, 10)
                _, err := s.CreateMessageComplex(channel.Id, rest.CreateMessageData{
                        Content: "<@&"+str+">",
                        AllowedMentions: message.MentionEveryone,
                })

                if err != nil {
                        sentry.Error(err)
                } else {
                        // error is likely to be a permission error
                        //_ = s.DeleteMessage(channel.Id, pingRoleMsg.Id)
                        //fmt.Println("Error doing ping, check perms!")
                }
            }
     //}

	// Let the user know the ticket has been opened
	if panel == nil {
		utils.SendEmbed(s, msg.ChannelId, utils.Green, "Ticket", fmt.Sprintf("Opened a new ticket: %s", channel.Mention()), nil, 30, isPremium)
	} else {
		dmOnOpen, err := dbclient.Client.DmOnOpen.Get(msg.GuildId); if err != nil {
			sentry.Error(err)
		}

		if dmOnOpen && dmChannel.Id != 0 {
			utils.SendEmbed(s, dmChannel.Id, utils.Green, "Ticket", fmt.Sprintf("Opened a new ticket: %s", channel.Mention()), nil, 0, isPremium)
		}
	}

	go statsd.IncrementKey(statsd.TICKETS)

	if isPremium {
		go createWebhook(s, id, msg.GuildId, channel.Id)
	}
}

// has hit ticket limit, ticket limit
func getTicketLimit(guildId, userId uint64) (bool, int) {
	var openedTickets []database.Ticket
	var ticketLimit uint8

	group, _ := errgroup.WithContext(context.Background())

	// get ticket limit
	group.Go(func() (err error) {
		ticketLimit, err = dbclient.Client.TicketLimit.Get(guildId)
		return
	})

	group.Go(func() (err error) {
		openedTickets, err = dbclient.Client.Tickets.GetOpenByUser(guildId, userId)
		return
	})

	if err := group.Wait(); err != nil {
		sentry.Error(err)
		return true, 1
	}

	return len(openedTickets) >= int(ticketLimit), int(ticketLimit)
}

func createWebhook(s *gateway.Shard, ticketId int, guildId, channelId uint64) {
	// TODO: Re-add permission check
	//if permission.HasPermissionsChannel(ctx.Shard, ctx.GuildId, ctx.Shard.SelfId(), channelId, permission.ManageWebhooks) { // Do we actually need this?
	webhook, err := s.CreateWebhook(channelId, rest.WebhookData{
		Username: s.SelfUsername(),
		Avatar:   s.SelfAvatar(256),
	})
	if err != nil {
		sentry.Error(err)
		return
	}

	dbWebhook := database.Webhook{
		Id:    webhook.Id,
		Token: webhook.Token,
	}

	if err := dbclient.Client.Webhooks.Create(guildId, ticketId, dbWebhook); err != nil {
		sentry.Error(err)
	}
	//}
}

// returns msg id
func sendWelcomeMessage(s *gateway.Shard, guildId, channelId, userId uint64, isPremium bool, subject string) uint64 {
	// Send welcome message
	welcomeMessage, err := dbclient.Client.WelcomeMessages.Get(guildId); if err != nil {
		sentry.Error(err)
		welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
	}

	// %average_response%
	if isPremium && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime, err := dbclient.Client.FirstResponseTime.GetAverage(guildId, time.Hour * 24 * 7)
		if err != nil {
			sentry.Error(err)
		} else {
			strings.Replace(welcomeMessage, "%average_response%", utils.FormatTime(*weeklyResponseTime), -1)
		}
	}

	// variables
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", fmt.Sprintf("<@%d>", userId), -1)
	// welcomeMessage = strings.Replace(welcomeMessage, "%server%", ctx.Guild.Name, -1)

	// Send welcome message
	if msg, err := utils.SendEmbedWithResponse(s, channelId, utils.Green, subject, welcomeMessage, nil, 0, isPremium); err == nil {
		// Add close reaction to the welcome message
		err := s.CreateReaction(channelId, msg.Id, "🔒")
		if err != nil {
			sentry.Error(err)
		}

		return msg.Id
	}

	return 0
}

func CreateOverwrites(guildId, userId, rolesmentioned, selfId uint64) (overwrites []channel.PermissionOverwrite) {
	// Apply permission overwrites
	overwrites = append(overwrites, channel.PermissionOverwrite{ // @everyone
		Id:    guildId,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	// Create list of members & roles who should be added to the ticket
	allowedUsers := make([]uint64, 0)
	allowedRoles := make([]uint64, 0)

	// Get support reps & admins
	supportUsers, err := dbclient.Client.Permissions.GetSupport(guildId); if err != nil {
		sentry.Error(err)
	}

	for _, user := range supportUsers {
		allowedUsers = append(allowedUsers, user)
	}
	
	neededroles, err := dbclient.Client.PanelRoleMentions.GetRoles(rolesmentioned)
        if err != nil {
                sentry.Error(err)
        }

	// Get support roles & admin roles
	supportRoles, err := dbclient.Client.RolePermissions.GetSupportRoles(guildId); if err != nil {
		sentry.Error(err)
	}

	for _, role := range supportRoles {
		allowedRoles = append(allowedRoles, role)
	}
	
	for _, role := range neededroles {
		allowedRoles = append(allowedRoles, role)
	}

	// Add ourselves and the sender
	allowedUsers = append(allowedUsers, selfId, userId)

	for _, member := range allowedUsers {
		allow := []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == selfId {
			allow = append(allow, permission.ManageWebhooks)
		}

		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    member,
			Type:  channel.PermissionTypeMember,
			Allow: permission.BuildPermissions(allow...),
			Deny:  0,
		})
	}

	for _, role := range allowedRoles {
		overwrites = append(overwrites, channel.PermissionOverwrite{
			Id:    role,
			Type:  channel.PermissionTypeRole,
			Allow: permission.BuildPermissions(permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks),
			Deny:  0,
		})
	}

	return overwrites
}
