package listeners

import (
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func OnCommand(s *discordgo.Session, e *discordgo.MessageCreate) {
	if e.Author.Bot {
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
	var args []string
	if len(split) > 1 {
		args = split[1:]
	}

	var c command.Command
	for _, cmd := range command.Commands {
		if strings.ToLower(cmd.Name()) == strings.ToLower(root) {
			parent := cmd
			index := 0

			for {
				if len(args) > index {
					childName := args[index]
					found := false

					for _, child := range parent.Children() {
						if strings.ToLower(child.Name()) == strings.ToLower(childName) {
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
				childArgs = args[:index]
			}

			args = childArgs
			c = parent
		}
	}

	premiumChan := make(chan bool)
	utils.IsPremiumGuild(e.GuildID, premiumChan)
	premiumGuild := <- premiumChan

	ctx := command.CommandContext{
		Session: s,
		User: *e.Author,
		Guild: e.GuildID,
		Channel: e.ChannelID,
		Message: *e.Message,
		Root: root,
		Args: args,
	}

	if c != nil {
		permLevel := make(chan utils.PermissionLevel)
		ctx.GetPermissionLevel(permLevel)
		if int(c.PermissionLevel()) > int(<- permLevel) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", NO_PERMISSION)
			return
		}

		if c.AdminOnly() && !isBotAdmin(e.Author.ID) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", NO_PERMISSION)
			return
		}

		if c.HelperOnly() && !isBotHelper(e.Author.ID) {
			ctx.ReactWithCross()
			ctx.SendEmbed(utils.Red, "Error", NO_PERMISSION)
			return
		}

		if c.PremiumOnly() {
			if !premiumGuild {
				ctx.ReactWithCross()
				ctx.SendEmbed(utils.Red, "Premium Only Command", PREMIUM_MESSAGE)
				return
			}
		}

		go c.Execute(ctx)

		utils.DeleteAfter(utils.SentMessage{Session: s, Message: e.Message}, 30)
	}
}
