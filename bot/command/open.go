package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
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

var idLocks = make(map[int64]*sync.Mutex)

func (OpenCommand) Execute(ctx utils.CommandContext) {
	ch := make(chan int64)

	// Get the panel struct, if we're using one
	var panel database.Panel
	if ctx.IsFromPanel {
		panelChan := make(chan database.Panel)
		go database.GetPanelByMessageId(ctx.MessageId, panelChan)
		panel = <-panelChan
	}

	// If we're using a panel, then we need to create the ticket in the specified category
	var category int64
	if ctx.IsFromPanel && panel.TargetCategory != 0 {
		category = panel.TargetCategory
	} else { // else we can just use the default category
		go database.GetCategory(ctx.GuildId, ch)
		category = <- ch
	}

	// Make sure the category exists
	if category != 0 {
		categoryStr := strconv.Itoa(int(category))
		if _, err := ctx.Session.State.Channel(categoryStr); err != nil {
			if _, err = ctx.Session.Channel(categoryStr); err != nil {
				category = 0
			}
		}
	}

	requiredPerms := []utils.Permission{
		utils.ManageChannels,
		utils.ManageRoles,
		utils.ViewChannel,
		utils.SendMessages,
		utils.ReadMessageHistory,
	}

	hasAdmin := make(chan bool)
	go ctx.MemberHasPermission(utils.Id, utils.Administrator, hasAdmin)
	if !<-hasAdmin {
		for _, perm := range requiredPerms {
			hasPermChan := make(chan bool)
			go ctx.MemberHasPermission(utils.Id, perm, hasPermChan)
			if !<-hasPermChan {
				ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
				if ctx.ShouldReact {
					ctx.ReactWithCross()
				}
				return
			}
		}
	}

	categoryStr := strconv.Itoa(int(category))
	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		ch, err := ctx.Session.Channel(categoryStr); if err != nil {
			useCategory = false
			go database.DeleteCategory(ctx.GuildId)
			return
		}

		if ch.Type != discordgo.ChannelTypeGuildCategory {
			useCategory = false
			go database.DeleteCategory(ctx.GuildId)
			return
		}

		hasAdmin := make(chan bool)
		go ctx.ChannelMemberHasPermission(categoryStr, utils.Id, utils.Administrator, hasAdmin)
		if !<-hasAdmin {
			for _, perm := range requiredPerms {
				hasPermChan := make(chan bool)
				go ctx.ChannelMemberHasPermission(categoryStr, utils.Id, perm, hasPermChan)
				if !<-hasPermChan {
					ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
					if ctx.ShouldReact {
						ctx.ReactWithCross()
					}
					return
				}
			}
		}
	}

	// Make sure ticket count is whithin ticket limit
	ticketLimitChan := make(chan int)
	go database.GetTicketLimit(ctx.GuildId, ticketLimitChan)
	ticketLimit := <- ticketLimitChan

	userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	ticketCount := 0
	openedTicketsChan := make(chan []string)
	go database.GetOpenTicketsOpenedBy(ctx.GuildId, userId, openedTicketsChan)
	openedTickets := <-openedTicketsChan
	ticketCount = len(openedTickets)

	if ticketCount >= ticketLimit {
		if ctx.ShouldReact {
			ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("You are only able to open %d tickets at once", ticketLimit))
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
		channels, err := ctx.Session.GuildChannels(ctx.Guild.ID); if err != nil {
			channels = make([]*discordgo.Channel, 0)
		}

		channelCount := 0
		categoryRaw := strconv.Itoa(int(category))
		for _, channel := range channels {
			if channel.ParentID != "" && channel.ParentID == categoryRaw {
				channelCount += 1
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

	// ID lock
	lock := idLocks[ctx.GuildId]
	if lock == nil {
		lock = &sync.Mutex{}
	}
	idLocks[ctx.GuildId] = lock

	// Create channel
	lock.Lock()
	idChan := make(chan int)
	go database.CreateTicket(ctx.GuildId, userId, idChan)
	id := <- idChan
	lock.Unlock()

	overwrites := createOverwrites(ctx)

	// Create ticket name
	var name string

	namingScheme := make(chan database.NamingScheme)
	go database.GetTicketNamingScheme(ctx.GuildId, namingScheme)
	if <-namingScheme == database.Username {
		name = fmt.Sprintf("ticket-%s", ctx.User.Username)
	} else {
		name = fmt.Sprintf("ticket-%d", id)
	}

	data := discordgo.GuildChannelCreateData{
		Name: name,
		Type: discordgo.ChannelTypeGuildText,
		Topic: subject,
		PermissionOverwrites: overwrites,
	}
	if useCategory {
		data.ParentID = strconv.Itoa(int(category))
	}

	c, err := ctx.Session.GuildChannelCreateComplex(ctx.Guild.ID, data)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// UpdateUser channel in DB
	channelId, err := strconv.ParseInt(c.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}
	go database.SetTicketChannel(id, ctx.GuildId, channelId)

	sendWelcomeMessage(ctx, c, subject, id)

	// Ping @everyone
	pingEveryoneChan := make(chan bool)
	go database.IsPingEveryone(ctx.GuildId, pingEveryoneChan)
	pingEveryone := <- pingEveryoneChan

	if pingEveryone {
		msg, err := ctx.Session.ChannelMessageSend(c.ID, "@everyone")
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		} else {
			if err = ctx.Session.ChannelMessageDelete(c.ID, msg.ID); err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			}
		}
	}

	if ctx.ShouldReact {
		// Let the user know the ticket has been opened
		ctx.SendEmbed(utils.Green, "Ticket", fmt.Sprintf("Opened a new ticket: %s", c.Mention()))
	}

	go statsd.IncrementKey(statsd.TICKETS)
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

func (OpenCommand) AdminOnly() bool {
	return false
}

func (OpenCommand) HelperOnly() bool {
	return false
}

func calculateWeeklyResponseTime(ctx utils.CommandContext, res chan int64) {
	responseTimesChan := make(chan map[string]int64)
	go database.GetGuildResponseTimes(ctx.GuildId, responseTimesChan)
	responseTimes := <-responseTimesChan

	openTimesChan := make(chan map[string]*int64)
	go database.GetOpenTimes(ctx.GuildId, openTimesChan)
	openTimes := <-openTimesChan

	current := time.Now().UnixNano() / int64(time.Millisecond)

	var weekly int64
	var weeklyCounter int
	for uuid, t := range responseTimes {
		openTime := openTimes[uuid]
		if openTime == nil {
			continue
		}

		if current - *openTime < 60 * 60 * 24 * 7 * 1000 {
			weekly += t
			weeklyCounter++
		}
	}
	if weeklyCounter > 0 {
		weekly = weekly / int64(weeklyCounter)
	}

	res <- weekly
}

func sendWelcomeMessage(ctx utils.CommandContext, channel *discordgo.Channel, subject string, ticketId int) {
	// Send welcome message
	welcomeMessageChan := make(chan string)
	go database.GetWelcomeMessage(ctx.GuildId, welcomeMessageChan)
	welcomeMessage := <- welcomeMessageChan

	// %average_response%
	if ctx.IsPremium && strings.Contains(welcomeMessage, "%average_response%") {
		weeklyResponseTime := make(chan int64)
		go calculateWeeklyResponseTime(ctx, weeklyResponseTime)
		welcomeMessage = strings.Replace(welcomeMessage, "%average_response%", utils.FormatTime(<-weeklyResponseTime), -1)
	}

	// %user%
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", ctx.User.Mention(), -1)

	if welcomeMessage == "" {
		welcomeMessage = "No message specified"
	}

	// Send welcome message
	if msg := utils.SendEmbedWithResponse(ctx.Session, channel.ID, utils.Green, subject, welcomeMessage, 0, ctx.IsPremium); msg != nil {
		// Add close reaction to the welcome message
		err := ctx.Session.MessageReactionAdd(channel.ID, msg.ID, "🔒")
		if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
		} else {
			// Parse message ID
			messageId, err := strconv.ParseInt(msg.ID, 10, 64)
			if err != nil {
				sentry.ErrorWithContext(err, ctx.ToErrorContext())
			} else {
				go database.SetWelcomeMessageId(ticketId, ctx.GuildId, messageId)
			}
		}
	}
}

func createOverwrites(ctx utils.CommandContext) []*discordgo.PermissionOverwrite {
	// Apply permission overwrites
	overwrites := make([]*discordgo.PermissionOverwrite, 0)
	overwrites = append(overwrites, &discordgo.PermissionOverwrite{ // @everyone
		ID: ctx.Guild.ID,
		Type: "role",
		Allow: 0,
		Deny: utils.SumPermissions(utils.ViewChannel),
	})

	// Create list of members & roles who should be added to the ticket
	allowedUsers := make([]string, 0)
	allowedRoles := make([]string, 0)

	// Get support reps
	supportUsers := make(chan []int64)
	go database.GetSupport(ctx.Guild.ID, supportUsers)
	for _, user := range <-supportUsers {
		allowedUsers = append(allowedUsers, strconv.Itoa(int(user)))
	}

	// Get support roles
	supportRoles := make(chan []int64)
	go database.GetSupportRoles(ctx.Guild.ID, supportRoles)
	for _, role := range <-supportRoles {
		allowedRoles = append(allowedRoles, strconv.Itoa(int(role)))
	}

	// Add ourselves and the sender
	allowedUsers = append(allowedUsers, utils.Id, ctx.User.ID)

	for _, member := range allowedUsers {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID: member,
			Type: "member",
			Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
			Deny: 0,
		})
	}

	for _, role := range allowedRoles {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID: role,
			Type: "role",
			Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
			Deny: 0,
		})
	}

	return overwrites
}
