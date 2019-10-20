package bot

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/listeners"
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/jonas747/dshardmanager"
	"os"
	"time"
)

func Start(ch chan os.Signal) {
	discord := dshardmanager.New(fmt.Sprintf("Bot %s", config.Conf.Bot.Token))
	discord.SetNumShards(config.Conf.Bot.Sharding.Total)

	discord.AddHandler(listeners.OnChannelDelete)
	discord.AddHandler(listeners.OnCommand)
	discord.AddHandler(listeners.OnFirstResponse)
	discord.AddHandler(listeners.OnGuildCreate)
	//discord.AddHandler(listeners.OnMessage)
	discord.AddHandler(listeners.OnGuildCreateUserData)
	discord.AddHandler(listeners.OnPanelReact)
	discord.AddHandler(listeners.OnSetupProgress)
	discord.AddHandler(listeners.OnUserJoin)
	discord.AddHandler(listeners.OnUserUpdate)

	if err := discord.Start(); err != nil {
		panic(err)
	}

	if self, err := discord.Session(0).User("@me"); err == nil {
		if self != nil {
			utils.AvatarUrl = self.AvatarURL("128")
			utils.Id = self.ID
		}
	}

	if config.Conf.ServerCounter.Enabled {
		go func() {
			for {
				time.Sleep(20 * time.Second)
				servercounter.UpdateServerCount()
			}
		}()
	}

	<-ch
	if err := discord.StopAll(); err != nil {
		sentry.Error(err)
	}
}
