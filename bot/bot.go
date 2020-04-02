package bot

import (
	"github.com/TicketsBot/TicketsGo/bot/listeners"
	"github.com/TicketsBot/TicketsGo/bot/listeners/messagequeue"
	modmaillisteners "github.com/TicketsBot/TicketsGo/bot/modmail/listeners"
	"github.com/TicketsBot/TicketsGo/bot/modmail/utils"
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/user"
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

	shardManager.Presence = user.BuildStatus(user.ActivityTypePlaying, "DM for help | t!help")

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

	shardManager.Connect()

	go messagequeue.ListenPanelCreations(&shardManager)
	go messagequeue.ListenTicketClose(&shardManager)
	go utils.ListenMutualGuildRequests(&shardManager)
	go utils.ListenMutualGuildResponses()

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
