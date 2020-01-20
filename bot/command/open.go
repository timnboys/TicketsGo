package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
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

func (OpenCommand) Execute(ctx utils.CommandContext) {
	ch := make(chan int64)

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	go database.GetCategory(guildId, ch)
	category := <- ch

	// Make sure the category exists
	if category != 0 {
		if _, err = ctx.Session.Channel(strconv.Itoa(int(category))); err != nil {
			category = 0
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
	go database.GetTicketLimit(guildId, ticketLimitChan)
	ticketLimit := <- ticketLimitChan

	userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	ticketCount := 0
	ticketsChan := make(chan map[int64]int)
	go database.GetTicketsOpenedBy(guildId, userId, ticketsChan)
	for channel, id := range <-ticketsChan {
		_, err:= ctx.Session.State.Channel(strconv.Itoa(int(channel)))
		if err != nil { // An admin has deleted the channel manually
			go database.Close(guildId, id)
		} else {
			ticketCount += 1
		}
	}

	if ticketCount >= ticketLimit {
		if ctx.ShouldReact {
			ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("You are only able to open %d tickets at once", ticketLimit))
			ctx.ReactWithCross()
		}
		return
	}

	// Generate subject
	subject := "No subject given"
	if len(ctx.Args) > 0 {
		subject = strings.Join(ctx.Args, " ")
	}
	if len(subject) > 256 {
		subject = subject[0:255]
	}

	// Make sure there's not > 50 channels in a category
	if useCategory {
		channels, err := ctx.Session.GuildChannels(ctx.Guild);
		if err != nil {
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

	// Create channel
	idChan := make(chan int)
	go database.CreateTicket(guildId, userId, idChan)
	id := <- idChan

	// Apply permission overwrites
	overwrites := make([]*discordgo.PermissionOverwrite, 0)
	overwrites = append(overwrites, &discordgo.PermissionOverwrite{ // @everyone
		ID: ctx.Guild,
		Type: "role",
		Allow: 0,
		Deny: utils.SumPermissions(utils.ViewChannel),
	})

	// Create list of people who should be added to the ticket
	allowed := make([]string, 0)

	// Get support reps
	supportChan := make(chan []int64)
	go database.GetSupport(ctx.Guild, supportChan)
	support := <- supportChan
	for _, user := range support {
		allowed = append(allowed, strconv.Itoa(int(user)))
	}

	// Get admins
	adminChan := make(chan []int64)
	go database.GetAdmins(ctx.Guild, adminChan)
	admin := <- adminChan
	for _, user := range admin {
		allowed = append(allowed, strconv.Itoa(int(user)))
	}

	// Add ourselves and the sender
	allowed = append(allowed, utils.Id, ctx.User.ID)

	for _, member := range allowed {
		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID: member,
			Type: "member",
			Allow: utils.SumPermissions(utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks),
			Deny: 0,
		})
	}

	data := discordgo.GuildChannelCreateData{
		Name: fmt.Sprintf("ticket-%d", id),
		Type: discordgo.ChannelTypeGuildText,
		Topic: subject,
		PermissionOverwrites: overwrites,
	}
	if useCategory {
		data.ParentID = strconv.Itoa(int(category))
	}

	c, err := ctx.Session.GuildChannelCreateComplex(ctx.Guild, data)
	if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}

	// UpdateUser channel in DB
	channelId, err := strconv.ParseInt(c.ID, 10, 64); if err != nil {
		sentry.ErrorWithContext(err, ctx.ToErrorContext())
		return
	}
	go database.SetTicketChannel(id, guildId, channelId)

	// Send welcome message
	welcomeMessageChan := make(chan string)
	go database.GetWelcomeMessage(guildId, welcomeMessageChan)
	welcomeMessage := <- welcomeMessageChan
	
	// %average_response%
	if ctx.IsPremium && strings.Contains(welcomeMessage, "%average_response%") {
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

			if current - *openTime < 60 * 60 * 24 * 7 * 1000 {
				weekly += t
				weeklyCounter++
			}
		}
		if weeklyCounter > 0 {
			weekly = weekly / int64(weeklyCounter)
		}

		welcomeMessage = strings.Replace(welcomeMessage, "%average_response%", utils.FormatTime(weekly), -1)
	}

	// %user%
	welcomeMessage = strings.Replace(welcomeMessage, "%user%", ctx.User.Mention(), -1)

	if welcomeMessage == "" {
		welcomeMessage = "No message specified"
	}

	utils.SendEmbed(ctx.Session, c.ID, utils.Green, subject, welcomeMessage, 0, ctx.IsPremium)

	// Ping @everyone
	pingEveryoneChan := make(chan bool)
	go database.IsPingEveryone(guildId, pingEveryoneChan)
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
