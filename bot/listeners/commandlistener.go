package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

func OnCommand(s *discordgo.Session, e *discordgo.MessageCreate) {
	if e.Author.Bot {
		return
	}

	// Ignore commands in DMs
	if e.GuildID == "" {
		return
	}

	guildId, err := strconv.ParseInt(e.GuildID, 10, 64); if err != nil {
		return
	}

	userId, err := strconv.ParseInt(e.Author.ID, 10, 64); if err != nil {
		return
	}

	ch := make(chan string)
	go database.GetPrefix(e.GuildID, ch)

	customPrefix := <- ch
	defaultPrefix := config.Conf.Bot.Prefix
	var usedPrefix string

	if strings.HasPrefix(e.Content, defaultPrefix) {
		usedPrefix = defaultPrefix
	} else if strings.HasPrefix(e.Content, customPrefix) {
		usedPrefix = customPrefix
	} else { // Not a command
		return
	}

	split := strings.Split(e.Content, " ")
	root := strings.TrimPrefix(split[0], usedPrefix)

	args := make([]string, 0)
	if len(split) > 1 {
		for _, arg := range split[1:] {
			if arg != "" {
				args = append(args, arg)
			}
		}
	}

	var c command.Command
	for _, cmd := range command.Commands {
		if strings.ToLower(cmd.Name()) == strings.ToLower(root) || contains(cmd.Aliases(), strings.ToLower(root)) {
			parent := cmd
			index := 0

			for {
				if len(args) > index {
					childName := args[index]
					found := false

					for _, child := range parent.Children() {
						if strings.ToLower(child.Name()) == strings.ToLower(childName) || contains(child.Aliases(), strings.ToLower(childName)) {
							parent = child
							found = true
							index++
						}
					}

					if !found {
						break
					}
				} else {
					break
				}
			}

			var childArgs []string
			if len(args) > 0 {
				childArgs = args[index:]
			}

			args = childArgs
			c = parent
		}
	}

	premiumChan := make(chan bool)
	go utils.IsPremiumGuild(e.GuildID, premiumChan)
	premiumGuild := <- premiumChan

	ctx := command.CommandContext{
		Session: s,
		User: *e.Author,
		UserID: userId,
		Guild: e.GuildID,
		GuildId: guildId,
		Channel: e.ChannelID,
		Message: *e.Message,
		Root: root,
		Args: args,
		IsPremium: premiumGuild,
		ShouldReact: true,
	}

	// Ensure user isn't blacklisted
	blacklisted := make(chan bool)
	go database.IsBlacklisted(ctx.GuildId, ctx.UserID, blacklisted)
	if <-blacklisted {
		ctx.ReactWithCross()
		return
	}

	if c != nil {
		permLevel := make(chan utils.PermissionLevel)
		go ctx.GetPermissionLevel(permLevel)
		if int(c.PermissionLevel()) > int(<- permLevel) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.AdminOnly() && !isBotAdmin(e.Author.ID) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.HelperOnly() && !isBotHelper(e.Author.ID) && !isBotAdmin(e.Author.ID) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", utils.NO_PERMISSION)
			return
		}

		if c.PremiumOnly() {
			if !premiumGuild {
				ctx.ReactWithCross()
				ctx.SendEmbed(utils.Red, "Premium Only Command", utils.PREMIUM_MESSAGE)
				return
			}
		}

		go c.Execute(ctx)

		utils.DeleteAfter(utils.SentMessage{Session: s, Message: e.Message}, 30)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

