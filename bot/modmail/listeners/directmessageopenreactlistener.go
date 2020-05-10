package listeners

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/modmail"
	modmailutils "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/gateway/payloads/events"
)

func OnDirectOpenMessageReact(s *gateway.Shard, e *events.MessageReactionAdd) {
	if e.GuildId != 0 { // DMs only
		return
	}

	if e.UserId == s.SelfId() { // ignore our own reactions
		return
	}

	session, err := database.Client.ModmailSession.GetByUser(e.UserId)
	if err != nil {
		sentry.Error(err)
		return
	}

	if session.UserId != 0 {
		return
	}

	// Determine which emoji was used
	reaction := -1
	for i, emoji := range modmailutils.Emojis {
		if emoji == e.Emoji.Name {
			reaction = i
			break
		}
	}

	// Check a number emoji was used
	if reaction == -1 {
		return
	}

	// Remove reaction
	_ = s.DeleteUserReaction(e.ChannelId, e.MessageId, e.UserId, e.Emoji.Name)

	// Create DM channel
	dmChannel, err := s.CreateDM(e.UserId)
	if err != nil {
		// TODO: Error logging
		return
	}

	// Determine which guild we should open the channel in
	guilds := modmailutils.GetMutualGuilds(s, e.UserId)

	if reaction-1 >= len(guilds) {
		return
	}

	targetGuild := guilds[reaction-1]

	// Check blacklist
	isBlacklisted, err := database.Client.Blacklist.IsBlacklisted(targetGuild.Id, e.UserId)
	if err != nil {
		sentry.Error(err)
	}

	if isBlacklisted {
		utils.SendEmbed(s, dmChannel.Id, utils.Red, "Error", "You are blacklisted in this server!", nil, 30, true)
		return
	}

	// Get user object
	user, err := s.GetUser(e.UserId)
	if err != nil {
		sentry.Error(err)
		return
	}

	utils.SendEmbed(s, dmChannel.Id, utils.Green, "Modmail", fmt.Sprintf("Your modmail ticket in %s has been opened! Use `t!close` to close the session.", targetGuild.Name), nil, 0, true)

	// Send guild's welcome message
	welcomeMessage, err := database.Client.WelcomeMessages.Get(targetGuild.Id); if err != nil {
		sentry.Error(err)
		welcomeMessage = "Thank you for contacting support.\nPlease describe your issue (and provide an invite to your server if applicable) and wait for a response."
	}

	welcomeMessageId, err := utils.SendEmbedWithResponse(s, dmChannel.Id, utils.Green, "Modmail", welcomeMessage, nil, 0, true)
	if err != nil {
		utils.SendEmbed(s, dmChannel.Id, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), nil, 30, true)
		return
	}

	staffChannel, err := modmail.OpenModMailTicket(s, targetGuild, user, welcomeMessageId.Id)
	if err != nil {
		utils.SendEmbed(s, dmChannel.Id, utils.Red, "Error", fmt.Sprintf("An error has occurred: %s", err.Error()), nil, 30, true)
		return
	}

	utils.SendEmbed(s, staffChannel, utils.Green, "Modmail", welcomeMessage, nil, 0, true)
}
