package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"strconv"
)

func OnCloseReact(s *discordgo.Session, e *discordgo.MessageReactionAdd) {
	// Check the right emoji has been used
	if e.Emoji.Name != "🔒" {
		return
	}

	// Create error context for later
	errorContext := sentry.ErrorContext{
		Guild:   e.GuildID,
		User:    e.UserID,
		Channel: e.ChannelID,
		Shard:   s.ShardID,
	}

	// Parse message ID
	msgId, err := strconv.ParseInt(e.MessageID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Parse user ID
	userId, err := strconv.ParseInt(e.UserID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// In DMs
	if e.GuildID == "" {
		return
	}

	// Parse guild ID
	guildId, err := strconv.ParseInt(e.GuildID, 10, 64)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Get user object
	user, err := s.User(e.UserID)
	if err != nil {
		sentry.ErrorWithContext(err, errorContext)
		return
	}

	// Ensure that the user is an actual user, not a bot
	if user.Bot {
		return
	}

	// Parse the channel ID
	channelId, err := strconv.ParseInt(e.ChannelID, 10, 64)
	if err != nil {
		sentry.LogWithContext(err, errorContext)
		return
	}

	// Get the ticket properties
	ticketChan := make(chan database.Ticket)
	go database.GetTicketByChannel(channelId, ticketChan)
	ticket := <-ticketChan

	// Check that this channel is a ticket channel
	if ticket.Uuid == "" {
		return
	}

	// Check that the ticket has a welcome message
	if ticket.WelcomeMessageId == nil {
		return
	}

	// Check that the message being reacted to is the welcome message
	if msgId != *ticket.WelcomeMessageId {
		return
	}

	// No need to remove the reaction since we'ere deleting the channel anyway

	// Get guild obj
	guild, err := s.State.Guild(e.GuildID)
	if err != nil {
		guild, err = s.Guild(e.GuildID)
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
		GuildId: guildId,
		Guild:   guild,
	}, isPremium)

	// Get the member object
	member, err := s.State.Member(e.GuildID, e.UserID)
	if err != nil {
		member, err = s.GuildMember(e.GuildID, e.UserID)
		if err != nil {
			sentry.LogWithContext(err, errorContext)
			return
		}
	}

	ctx := utils.CommandContext{
		Session:     s,
		User:        *user,
		UserID:      userId,
		Guild:       guild,
		GuildId:     guildId,
		Channel:     e.ChannelID,
		ChannelId:   channelId,
		MessageId:   msgId,
		Root:        "close",
		Args:        make([]string, 0),
		IsPremium:   <-isPremium,
		ShouldReact: false,
		Member:      member,
	}

	go command.CloseCommand{}.Execute(ctx)
}
