package command

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
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

func (OpenCommand) Execute(ctx CommandContext) {
	ch := make(chan int64)

	guildId, err := strconv.ParseInt(ctx.Guild, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}

	go database.GetCategory(guildId, ch)
	category := <- ch

	requiredPerms := []utils.Permission{
		utils.ManageChannels,
		utils.ManageRoles,
		utils.ViewChannel,
		utils.SendMessages,
		utils.ReadMessageHistory,
	}

	for _, perm := range requiredPerms {
		hasPermChan := make(chan bool)
		go ctx.MemberHasPermission(utils.Id, perm, hasPermChan)
		if !<-hasPermChan {
			ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
			ctx.ReactWithCross()
			return
		}
	}

	useCategory := category != 0
	if useCategory {
		for _, perm := range requiredPerms {
			hasPermChan := make(chan bool)
			go ctx.ChannelMemberHasPermission(strconv.Itoa(int(category)), utils.Id, perm, hasPermChan)
			if !<-hasPermChan {
				ctx.SendEmbed(utils.Red, "Error", "I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions.")
				ctx.ReactWithCross()
				return
			}
		}
	}

	// Make sure ticket count is whithin ticket limit
	ticketLimitChan := make(chan int)
	go database.GetTicketLimit(guildId, ticketLimitChan)
	ticketLimit := <- ticketLimitChan

	userId, err := strconv.ParseInt(ctx.User.ID, 10, 64); if err != nil {
		log.Error(err.Error())
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
		ctx.SendEmbed(utils.Red, "Error", fmt.Sprintf("You are only able to open %d tickets at once", ticketLimit))
		ctx.ReactWithCross()
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

		if channelCount > 50 {
			ctx.SendEmbed(utils.Red, "Error", "There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category")
			return
		}
	}

	ctx.ReactWithCheck()

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
		log.Error(err.Error())
		return
	}

	// Update channel in DB
	channelId, err := strconv.ParseInt(c.ID, 10, 64); if err != nil {
		log.Error(err.Error())
		return
	}
	go database.SetTicketChannel(id, guildId, channelId)

	// Send welcome message
	// TODO: %average_response%
	welcomeMessageChan := make(chan string)
	go database.GetWelcomeMessage(guildId, welcomeMessageChan)
	welcomeMessage := <- welcomeMessageChan

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
			log.Error(err.Error())
		} else {
			if err = ctx.Session.ChannelMessageDelete(c.ID, msg.ID); err != nil {
				log.Error(err.Error())
			}
		}
	}

	// Let the user know the ticket has been opened
	ctx.SendEmbed(utils.Green, "Ticket", fmt.Sprintf("Opened a new ticket: %s", c.Mention()))
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
