package modmail

import (
	"errors"
	"fmt"
	modmaildatabase "github.com/TicketsBot/TicketsGo/bot/modmail/database"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	uuid "github.com/satori/go.uuid"
	"strconv"
)

func OpenModMailTicket(shard *discordgo.Session, guild modmailutils.UserGuild, user *discordgo.User, userId int64) error {
	id, err := uuid.NewV4(); if err != nil {
		sentry.Error(err)
		return errors.New("Failed to generate UUID")
	}

	guildId := strconv.Itoa(int(guild.Id))

	// If we're using a panel, then we need to create the ticket in the specified category
	categoryChan := make(chan int64)
	go database.GetCategory(guild.Id, categoryChan)
	category := <-categoryChan

	// Make sure the category exists
	if category != 0 {
		categoryStr := strconv.Itoa(int(category))
		if _, err := shard.State.Channel(categoryStr); err != nil {
			if _, err = shard.Channel(categoryStr); err != nil {
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
	go utils.MemberHasPermission(shard, guildId, utils.Id, utils.Administrator, hasAdmin)
	if !<-hasAdmin {
		for _, perm := range requiredPerms {
			hasPermChan := make(chan bool)
			go utils.MemberHasPermission(shard, guildId, utils.Id, perm, hasPermChan)
			if !<-hasPermChan {
				return errors.New("I do not have the correct permissions required to create the channel in the server")
			}
		}
	}

	categoryStr := strconv.Itoa(int(category))
	useCategory := category != 0
	if useCategory {
		// Check if the category still exists
		ch, err := shard.Channel(categoryStr); if err != nil {
			useCategory = false
			go database.DeleteCategory(guild.Id)
			return errors.New("Ticket category has been deleted")
		}

		if ch.Type != discordgo.ChannelTypeGuildCategory {
			useCategory = false
			go database.DeleteCategory(guild.Id)
			return errors.New("Ticket category is not a ticket category")
		}

		hasAdmin := make(chan bool)
		go utils.ChannelMemberHasPermission(shard, guildId, categoryStr, utils.Id, utils.Administrator, hasAdmin)
		if !<-hasAdmin {
			for _, perm := range requiredPerms {
				hasPermChan := make(chan bool)
				go utils.ChannelMemberHasPermission(shard, guildId, categoryStr, utils.Id, perm, hasPermChan)
				if !<-hasPermChan {
					return errors.New("I am missing the required permissions on the ticket category. Please ask the guild owner to assign me permissions to manage channels and manage roles / manage permissions")
				}
			}
		}
	}

	if useCategory {
		channels, err := shard.GuildChannels(guildId); if err != nil {
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
			return errors.New("There are too many tickets in the ticket category. Ask an admin to close some, or to move them to another category")
		}
	}

	// Create channel
	name := fmt.Sprintf("modmail-%s", user.Username)
	overwrites := createOverwrites(guildId)

	data := discordgo.GuildChannelCreateData{
		Name: name,
		Type: discordgo.ChannelTypeGuildText,
		PermissionOverwrites: overwrites,
	}
	if useCategory {
		data.ParentID = strconv.Itoa(int(category))
	}

	channel, err := shard.GuildChannelCreateComplex(guildId, data); if err != nil {
		sentry.Error(err)
		return err
	}

	channelId, err := strconv.ParseInt(channel.ID, 10, 64); if err != nil {
		sentry.Error(err)
		return err
	}

	go modmaildatabase.CreateModMailSession(id.String(), guild.Id, userId, channelId)
	return nil
}

func createOverwrites(guildId string) []*discordgo.PermissionOverwrite {
	// Apply permission overwrites
	overwrites := make([]*discordgo.PermissionOverwrite, 0)
	overwrites = append(overwrites, &discordgo.PermissionOverwrite{ // @everyone
		ID: guildId,
		Type: "role",
		Allow: 0,
		Deny: utils.SumPermissions(utils.ViewChannel),
	})

	// Create list of members & roles who should be added to the ticket
	allowedUsers := make([]string, 0)
	allowedRoles := make([]string, 0)

	// Get support reps
	supportUsers := make(chan []int64)
	go database.GetSupport(guildId, supportUsers)
	for _, user := range <-supportUsers {
		allowedUsers = append(allowedUsers, strconv.Itoa(int(user)))
	}

	// Get support roles
	supportRoles := make(chan []int64)
	go database.GetSupportRoles(guildId, supportRoles)
	for _, role := range <-supportRoles {
		allowedRoles = append(allowedRoles, strconv.Itoa(int(role)))
	}

	// Add ourselves
	allowedUsers = append(allowedUsers, utils.Id)

	for _, member := range allowedUsers {
		allow := []utils.Permission{utils.ViewChannel, utils.SendMessages, utils.AddReactions, utils.AttachFiles, utils.ReadMessageHistory, utils.EmbedLinks}

		// Give ourselves permissions to create webbooks
		if member == utils.Id {
			allow = append(allow, utils.ManageWebhooks)
		}

		overwrites = append(overwrites, &discordgo.PermissionOverwrite{
			ID: member,
			Type: "member",
			Allow: utils.SumPermissions(allow...),
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

