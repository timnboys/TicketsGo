package bot

import (
	"github.com/TicketsBot/TicketsGo/bot/listeners"
	"github.com/TicketsBot/TicketsGo/bot/listeners/messagequeue"
	modmaillisteners "github.com/TicketsBot/TicketsGo/bot/modmail/listeners"
	utils2 "github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/bot/utils"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects"
	"os"
	"time"
)

func Start(ch chan os.Signal) {
	cacheFactory := cache.MemoryCacheFactory(cache.CacheOptions{
		Guilds:      true,
		Users:       true,
		Channels:    true,
		Roles:       true,
		Emojis:      false,
		VoiceStates: false,
	})

	shardManager := gateway.NewShardManager(config.Conf.Bot.Token, gateway.ShardOptions{
		Total:   config.Conf.Bot.Sharding.Total,
		Lowest:  config.Conf.Bot.Sharding.Lowest,
		Highest: config.Conf.Bot.Sharding.Max,
	}, cacheFactory)

	shardManager.Presence = objects.BuildStatus(objects.ActivityTypePlaying, "DM for help | t!help")

	shardManager.RegisterListeners(
		listeners.OnChannelCreate,
		listeners.OnChannelDelete,
		listeners.OnCloseReact,
		listeners.OnCommand,
		listeners.OnFirstResponse,
		listeners.OnMessage,
		listeners.OnGuildCreate,
		listeners.OnGuildCreateUserData,
		listeners.OnGuildLeave,
		listeners.OnPanelReact,
		listeners.OnReady,
		listeners.OnSetupProgress,
		listeners.OnUserJoin,
		listeners.OnUserUpdate,

		modmaillisteners.OnDirectMessage,
		modmaillisteners.OnDirectMessageReact,
		modmaillisteners.OnModMailChannelMessage,
	)

	if err := shardManager.Connect(); err != nil {
		panic(err)
	}

	go messagequeue.ListenPanelCreations(shardManager)
	go messagequeue.ListenTicketClose(shardManager)
	go utils2.ListenMutualGuildRequests(shardManager)
	go utils2.ListenMutualGuildResponses()

	discordgo.Session{}.User("")
	if self, err := shardManager.Session(0).User("@me"); err == nil {
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
}
