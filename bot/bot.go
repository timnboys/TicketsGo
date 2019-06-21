package bot

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/listeners"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/apex/log"
	"github.com/bwmarrin/discordgo"
	"os"
	"time"
)

func Start(ch chan os.Signal) {
	discord, err := discordgo.New(fmt.Sprintf("Bot %s", config.Conf.Bot.Token)); if err != nil {
		panic(err)
	}

	discord.AddHandler(listeners.OnChannelDelete)
	discord.AddHandler(listeners.OnCommand)
	discord.AddHandler(listeners.OnFirstResponse)
	discord.AddHandler(listeners.OnGuildCreate)
	discord.AddHandler(listeners.OnSetupProgress)
	discord.AddHandler(listeners.OnUserJoin)
	discord.AddHandler(listeners.OnUserUpdate)

	if err = discord.Open(); err != nil {
		panic(err)
	}

	if self, err := discord.User("@me"); err == nil {
		if self != nil {
			utils.AvatarUrl = self.AvatarURL("128")
			utils.Id = self.ID
		}
	}

	go func() {
		for {
			time.Sleep(20 * time.Second)
			UpdateServerCount()
		}
	}()

	<-ch
	if err = discord.Close(); err != nil {
		log.Error(err.Error())
	}
}
