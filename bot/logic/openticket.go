package logic

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"
	"time"
)

var(
	idLocks = make(map[uint64]*sync.Mutex)
	idLocksLock = &sync.Mutex{}
)

// if panel != nil, msg should be artifically filled, excluding the message ID
func OpenTicket(s *gateway.Shard, user user.User, msg message.MessageReference, isPremium bool, args []string, panel *database.Panel) {
	// If we're using a panel, then we need to create the ticket in the specified category
	if msg.GuildId == 508392876359680000 {
		fmt.Println(1)
	}

	var category uint64
	if panel != nil && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		ch := make(chan uint64)
		go database.GetCategory(msg.GuildId, ch)
		category = <-ch
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

	if msg.GuildId == 508392876359680000 {
		fmt.Println(2)
	}

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

	if msg.GuildId == 508392876359680000 {
		fmt.Println(3)
	}

	// target channel for messaging the user
	// either DMs or the channel where the command was run
	var targetChannel uint64
	if panel == nil {
		targetChannel = msg.ChannelId
	} else {
		if dmChannel, err := s.CreateDM(user.Id); err == nil {
			targetChannel = dmChannel.Id
		} else {
			return
		}
	}

	if msg.GuildId == 508392876359680000 {
		fmt.Println(4)
	}

	// Make sure ticket count is within ticket limit
	violatesTicketLimit, limit := getTicketLimit(msg.GuildId, user.Id)
	if violatesTicketLimit {
		// Notify the user
		ticketsPluralised := "ticket"
		if limit > 1 {
			ticketsPluralised += "s"
		}
		content := fmt.Sprintf("You are only able to open %d %s at once", limit, ticketsPluralised)
		utils.SendEmbed(s, targetChannel, utils.Red, "Error", content, 30, isPremium)

		return
	}

	if msg.GuildId == 508392876359680000 {
		fmt.Println(5)
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

	if msg.GuildId == 508392876359680000 {
		fmt.Println(6)
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
			utils.SendEmbed(s, msg.ChannelId, utils.Red, "Error", "There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category", 30, isPremium)
			return
		}
	}

	if msg.GuildId == 508392876359680000 {
		fmt.Println(7)
	}

	if panel == nil {
		utils.ReactWithCheck(s, msg)
	}

	if msg.GuildId == 508392876359680000 {
		fmt.Println(8)
	}

	// ID lock: If we open 2 tickets simultaneously, they will end up having the same ID. Instead we should lock the guild until the ticket has been opened
	idLocksLock.Lock()
	lock := idLocks[msg.GuildId]
	if lock == nil {
		lock = &sync.Mutex{}
		idLocks[msg.GuildId] = lock
	}
	idLocksLock.Unlock()

	if msg.GuildId == 508392876359680000 {
		fmt.Println(9)
	}

	// Create channel
	ticketUuid := uuid.NewV4()

	lock.Lock()
	idChan := make(chan int)
	go database.CreateTicket(ticketUuid.String(), msg.GuildId, user.Id, idChan)
	id := <-idChan
	lock.Unlock()

	if msg.GuildId == 508392876359680000 {
		fmt.Println(10)
	}

	overwrites := createOverwrites(msg.GuildId, user.Id, s.SelfId())

	if msg.GuildId == 508392876359680000 {
		fmt.Println(11)
	}

	// Create ticket name
	var name string

	namingScheme := make(chan database.NamingScheme)
	go database.GetTicketNamingScheme(msg.GuildId, namingScheme)
	if <-namingScheme == database.Username {
		name = fmt.Sprintf("ticket-%s", user.Username)
	} else {
		name = fmt.Sprintf("ticket-%d", id)
	}

	if msg.GuildId == 508392876359680000 {
		fmt.Println(12)
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
	if err != nil {
		sentry.Error(err)
		return
	}

	if msg.GuildId == 508392876359680000 {
		fmt.Println(13)
	}

	// UpdateUser channel in DB
	go database.SetTicketChannel(id, msg.GuildId, channel.Id)

	if msg.GuildId == 508392876359680000 {
		fmt.Println(4)
	}

	sendWelcomeMessage(s, msg.GuildId, channel.Id, user.Id, isPremium, subject, id)

	// Ping @everyone
	pingEveryoneChan := make(chan bool)
	go database.IsPingEveryone(msg.GuildId, pingEveryoneChan)
	pingEveryone := <-pingEveryoneChan

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

	// Let the user know the ticket has been opened
	utils.SendEmbed(s, targetChannel, utils.Green, "Ticket", fmt.Sprintf("Opened a new ticket: %s", channel.Mention()), 30, isPremium)

	go statsd.IncrementKey(statsd.TICKETS)

	if isPremium {
		go createWebhook(s, channel.Id, ticketUuid.String())
	}
}

func getTicketLimit(guildId, userId uint64) (bool, int) {
	ticketLimitChan := make(chan int)
	go database.GetTicketLimit(guildId, ticketLimitChan)
	ticketLimit := <-ticketLimitChan

	ticketCount := 0
	openedTicketsChan := make(chan []string)
	go database.GetOpenTicketsOpenedBy(guildId, userId, openedTicketsChan)
	openedTickets := <-openedTicketsChan
	ticketCount = len(openedTickets)

	return ticketCount >= ticketLimit, ticketLimit
}

func createWebhook(s *gateway.Shard, channelId uint64, uuid string) {
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

	formatted := fmt.Sprintf("%d/%s", webhook.Id, webhook.Token)

	ticketWebhook := database.TicketWebhook{
		Uuid:       uuid,
		WebhookUrl: formatted,
	}
	ticketWebhook.AddWebhook()
	//}
}

func calculateWeeklyResponseTime(guildId uint64, res chan int64) {
	responseTimesChan := make(chan map[string]int64)
	go database.GetGuildResponseTimes(guildId, responseTimesChan)
	responseTimes := <-responseTimesChan

	openTimesChan := make(chan map[string]*int64)
	go database.GetOpenTimes(guildId, openTimesChan)
	openTimes := <-openTimesChan

	current := time.Now().UnixNano() / int64(time.Millisecond)

	var weekly int64
	var weeklyCounter int
	for uuid, t := range responseTimes {
		openTime := openTimes[uuid]
		if openTime == nil {
			continue
		}

		if current-*openTime < 60*60*24*7*1000 {
			weekly += t
			weeklyCounter++
		}
	}
	if weeklyCounter > 0 {
		weekly = weekly / int64(weeklyCounter)
	}

	res <- weekly
}

func sendWelcomeMessage(s *gateway.Shard, guildId, channelId, userId uint64, isPremium bool, subject string, ticketId int) {
	// Send welcome message
	welcomeMessageChan := make(chan string)
	go database.GetWelcomeMessage(guildId, welcomeMessageChan)
	welcomeMessage := <-welcomeMessageChan

	// %average_response%
	if isPremium && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime := make(chan int64)
		go calculateWeeklyResponseTime(guildId, weeklyResponseTime)
		welcomeMessage = strings.Replace(welcomeMessage, "%average_response%", utils.FormatTime(<-weeklyResponseTime), -1)
	}

	// variables
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", fmt.Sprintf("<@%d>", userId), -1)
	// welcomeMessage = strings.Replace(welcomeMessage, "%server%", ctx.Guild.Name, -1)

	if welcomeMessage == "" {
		welcomeMessage = "No message specified"
	}

	// Send welcome message
	if msg, err := utils.SendEmbedWithResponse(s, channelId, utils.Green, subject, welcomeMessage, 0, isPremium); err == nil {
		// Add close reaction to the welcome message
		err := s.CreateReaction(channelId, msg.Id, "🔒")
		if err != nil {
			sentry.Error(err)
		} else {
			// Parse message ID
			go database.SetWelcomeMessageId(ticketId, guildId, msg.Id)
		}
	}
}

func createOverwrites(guildId, userId, selfId uint64) []*channel.PermissionOverwrite {
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

	// Add ourselves and the sender
	allowedUsers = append(allowedUsers, selfId, userId)

	for _, member := range allowedUsers {
		allow := []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == selfId {
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