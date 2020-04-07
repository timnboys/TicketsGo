package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/channel"
	"github.com/rxdn/gdl/objects/channel/message"
	"github.com/rxdn/gdl/permission"
	"github.com/rxdn/gdl/rest"
	uuid "github.com/satori/go.uuid"
	"strings"
	"sync"
	"time"
)

type OpenCommand struct {
}

func (OpenCommand) Name() string {
	return "open"
}

func (OpenCommand) Description() string {
	return "Opens a new ticket"
}

func (OpenCommand) Aliases() []string {
	return []string{"new"}
}

func (OpenCommand) PermissionLevel() utils.PermissionLevel {
	return utils.Everyone
}

var idLocks = make(map[uint64]*sync.Mutex)

func (OpenCommand) Execute(ctx utils.CommandContext) {
	// Get the panel struct, if we're using one
	var panel database.Panel
	if ctx.IsFromPanel {
		panelChan := make(chan database.Panel)
		go database.GetPanelByMessageId(ctx.Message.Id, panelChan)
		panel = <-panelChan
	}

	// If we're using a panel, then we need to create the ticket in the specified category
	var category uint64
	if ctx.IsFromPanel && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		ch := make(chan uint64)
		go database.GetCategory(ctx.Guild.Id, ch)
		category = <-ch
	}

	// Make sure the category exists
	if category != 0 {
		if _, err := ctx.Shard.GetChannel(category); err != nil {
			category = 0
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

	if !permission.HasPermissions(ctx.Shard, ctx.Guild.Id, ctx.Shard.SelfId(), requiredPerms...) {
		ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
		if ctx.ShouldReact {
			ctx.ReactWithCross()
		}
		return
	}*/

	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		ch, err := ctx.Shard.GetChannel(category)
		if err != nil {
			useCategory = false
			go database.DeleteCategory(ctx.Guild.Id)
			return
		}

		if ch.Type != channel.ChannelTypeGuildCategory {
			useCategory = false
			go database.DeleteCategory(ctx.Guild.Id)
			return
		}

		// TODO: Re-add permission check
		/*if !permission.HasPermissionsChannel(ctx.Shard, ctx.Guild.Id, ctx.Shard.SelfId(), category, requiredPerms...) {
			ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
			if ctx.ShouldReact {
				ctx.ReactWithCross()
			}
			return
		}*/
	}

	// Make sure ticket count is whithin ticket limit
	ticketLimitChan := make(chan int)
	go database.GetTicketLimit(ctx.Guild.Id, ticketLimitChan)
	ticketLimit := <-ticketLimitChan

	ticketCount := 0
	openedTicketsChan := make(chan []string)
	go database.GetOpenTicketsOpenedBy(ctx.Guild.Id, ctx.User.Id, openedTicketsChan)
	openedTickets := <-openedTicketsChan
	ticketCount = len(openedTickets)

	if ticketCount >= ticketLimit {
		if ctx.ShouldReact {
			ticketsPluralised := "ticket"
			if ticketLimit > 1 {
				ticketsPluralised += "s"
			}

			ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("You are only able to open %d %s at once", ticketLimit, ticketsPluralised))
			ctx.ReactWithCross()
		}
		return
	}

	// Generate subject
	subject := "No subject given"
	if ctx.IsFromPanel && panel.Title != "" { // If we're using a panel, use the panel title as the subject
		subject = panel.Title
	} else {
		if len(ctx.Args) > 0 {
			subject = strings.Join(ctx.Args, " ")
		}
		if len(subject) > 256 {
			subject = subject[0:255]
		}
	}

	// Make sure there's not > 50 channels in a category
	if useCategory {
		channels, _ := ctx.Shard.GetGuildChannels(ctx.Guild.Id)

		channelCount := 0
		for _, channel := range channels {
			if channel.ParentId == category {
				channelCount++
			}
		}

		if channelCount >= 50 {
			ctx.SendEmbed(utils.Red, "Error", "There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category")
			return
		}
	}

	if ctx.ShouldReact {
		ctx.ReactWithCheck()
	}

	// ID lock: If we open 2 tickets simultaneously, they will end up having the same ID. Instead we should lock the guild until the ticket has been opened
	lock := idLocks[ctx.Guild.Id]
	if lock == nil {
		lock = &sync.Mutex{}
	}
	idLocks[ctx.Guild.Id] = lock

	// Create channel
	ticketUuid := uuid.NewV4()

	lock.Lock()
	idChan := make(chan int)
	go database.CreateTicket(ticketUuid.String(), ctx.Guild.Id, ctx.User.Id, idChan)
	id := <-idChan
	lock.Unlock()

	overwrites := createOverwrites(ctx)

	// Create ticket name
	var name string

	namingScheme := make(chan database.NamingScheme)
	go database.GetTicketNamingScheme(ctx.Guild.Id, namingScheme)
	if <-namingScheme == database.Username {
		name = fmt.Sprintf("ticket-%s", ctx.User.Username)
	} else {
		name = fmt.Sprintf("ticket-%d", id)
	}

	data := rest.CreateChannelData{
		Name:                 name,
		Type:                 channel.ChannelTypeGuildText,
		Topic:                subject,
		PermissionOverwrites: overwrites,
		ParentId:             category, //
	}
	if useCategory {
		data.ParentId = category
	}

	channel, err := ctx.Shard.CreateGuildChannel(ctx.Guild.Id, data)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// UpdateUser channel in DB
	go database.SetTicketChannel(id, ctx.Guild.Id, channel.Id)

	sendWelcomeMessage(ctx, &channel, subject, id)

	// Ping @everyone
	pingEveryoneChan := make(chan bool)
	go database.IsPingEveryone(ctx.Guild.Id, pingEveryoneChan)
	pingEveryone := <-pingEveryoneChan

	if pingEveryone {
		pingMessage, err := ctx.Shard.CreateMessageComplex(channel.Id, rest.CreateMessageData{
			Content:         "@everyone",
			AllowedMentions: message.MentionEveryone,
		})

		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		} else {
			if err = ctx.Shard.DeleteMessage(channel.Id, pingMessage.Id); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}
	}

	if ctx.ShouldReact {
		// Let the user know the ticket has been opened
		ctx.SendEmbed(utils.Green, "Ticket", fmt.Sprintf("Opened a new ticket: %s", channel.Mention()))
	}

	go statsd.IncrementKey(statsd.TICKETS)

	if ctx.IsPremium {
		go createWebhook(ctx, channel.Id, ticketUuid.String())
	}
}

func (OpenCommand) Parent() interface{} {
	return nil
}

func (OpenCommand) Children() []Command {
	return make([]Command, 0)
}

func (OpenCommand) PremiumOnly() bool {
	return false
}

func (OpenCommand) Category() Category {
	return Tickets
}

func (OpenCommand) AdminOnly() bool {
	return false
}

func (OpenCommand) HelperOnly() bool {
	return false
}

func createWebhook(ctx utils.CommandContext, channelId uint64, uuid string) {
	// TODO: Re-add permission check
	//if permission.HasPermissionsChannel(ctx.Shard, ctx.Guild.Id, ctx.Shard.SelfId(), channelId, permission.ManageWebhooks) { // Do we actually need this?
		webhook, err := ctx.Shard.CreateWebhook(channelId, rest.WebhookData{
			Username: ctx.Shard.SelfUsername(),
			Avatar:   ctx.Shard.SelfAvatar(256),
		})
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
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

func calculateWeeklyResponseTime(ctx utils.CommandContext, res chan int64) {
	responseTimesChan := make(chan map[string]int64)
	go database.GetGuildResponseTimes(ctx.Guild.Id, responseTimesChan)
	responseTimes := <-responseTimesChan

	openTimesChan := make(chan map[string]*int64)
	go database.GetOpenTimes(ctx.Guild.Id, openTimesChan)
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

func sendWelcomeMessage(ctx utils.CommandContext, channel *channel.Channel, subject string, ticketId int) {
	// Send welcome message
	welcomeMessageChan := make(chan string)
	go database.GetWelcomeMessage(ctx.Guild.Id, welcomeMessageChan)
	welcomeMessage := <-welcomeMessageChan

	// %average_response%
	if ctx.IsPremium && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime := make(chan int64)
		go calculateWeeklyResponseTime(ctx, weeklyResponseTime)
		welcomeMessage = strings.Replace(welcomeMessage, "%average_response%", utils.FormatTime(<-weeklyResponseTime), -1)
	}

	// variables
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", ctx.User.Mention(), -1)
	welcomeMessage = strings.Replace(welcomeMessage, "%server%", ctx.Guild.Name, -1)

	if welcomeMessage == "" {
		welcomeMessage = "No message specified"
	}

	// Send welcome message
	if msg, err := utils.SendEmbedWithResponse(ctx.Shard, channel.Id, utils.Green, subject, welcomeMessage, 0, ctx.IsPremium); err == nil {
		// Add close reaction to the welcome message
		err := ctx.Shard.CreateReaction(channel.Id, msg.Id, "🔒")
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		} else {
			// Parse message ID
			go database.SetWelcomeMessageId(ticketId, ctx.Guild.Id, msg.Id)
		}
	}
}

func createOverwrites(ctx utils.CommandContext) []*channel.PermissionOverwrite {
	// Apply permission overwrites
	overwrites := make([]*channel.PermissionOverwrite, 0)
	overwrites = append(overwrites, &channel.PermissionOverwrite{ // @everyone
		Id:    ctx.Guild.Id,
		Type:  channel.PermissionTypeRole,
		Allow: 0,
		Deny:  permission.BuildPermissions(permission.ViewChannel),
	})

	// Create list of members & roles who should be added to the ticket
	allowedUsers := make([]uint64, 0)
	allowedRoles := make([]uint64, 0)

	// Get support reps
	supportUsers := make(chan []uint64)
	go database.GetSupport(ctx.Guild.Id, supportUsers)
	for _, user := range <-supportUsers {
		allowedUsers = append(allowedUsers, user)
	}

	// Get support roles
	supportRoles := make(chan []uint64)
	go database.GetSupportRoles(ctx.Guild.Id, supportRoles)
	for _, role := range <-supportRoles {
		allowedRoles = append(allowedRoles, role)
	}

	// Add ourselves and the sender
	allowedUsers = append(allowedUsers, ctx.Shard.SelfId(), ctx.User.Id)

	for _, member := range allowedUsers {
		allow := []permission.Permission{permission.ViewChannel, permission.SendMessages, permission.AddReactions, permission.AttachFiles, permission.ReadMessageHistory, permission.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == ctx.Shard.SelfId() {
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
