package bot

import (
	"context"
	"github.com/TicketsBot/TicketsGo/bot/listeners"
	"github.com/TicketsBot/TicketsGo/bot/listeners/messagequeue"
	modmaillisteners "github.com/TicketsBot/TicketsGo/bot/modmail/listeners"
	"github.com/TicketsBot/TicketsGo/bot/servercounter"
	redis "github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/metrics/statsd"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/user"
	"github.com/rxdn/gdl/rest/ratelimit"
	"os"
	"time"
)

func Start(ch chan os.Signal) {

	shardOptions := gateway.ShardOptions{
		ShardCount: gateway.ShardCount{
			Total:   config.Conf.Bot.Sharding.Total,
			Lowest:  config.Conf.Bot.Sharding.Lowest,
			Highest: config.Conf.Bot.Sharding.Max,
		},
		CacheFactory:       buildCache(),
		RateLimitStore:     ratelimit.NewRedisStore(redis.Client.Client, "ratelimit"),
		GuildSubscriptions: false,
		Presence:           user.BuildStatus(user.ActivityTypePlaying, "DM for help | t!help"),
		Hooks: gateway.Hooks{
			ReconnectHook: func(shard *gateway.Shard) {
				go statsd.IncrementKey(statsd.RECONNECT)
			},
			IdentifyHook: func(shard *gateway.Shard) {
				go statsd.IncrementKey(statsd.IDENTIFY)
			},
			RestHook: func(url string) {
				go statsd.IncrementKey(statsd.REST)
				//@lgo sentry.LogRestRequest(url)
			},
		},
		Debug: true,
	}

	shardManager := gateway.NewShardManager(config.Conf.Bot.Token, shardOptions)

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

	if config.Conf.ServerCounter.Enabled {
		go func() {
			for {
				time.Sleep(20 * time.Second)
				servercounter.UpdateServerCount(&shardManager)
			}
		}()
	}

	<-ch
}

func buildCache() cache.CacheFactory {
	pool, err := pgxpool.Connect(context.Background(), config.Conf.Cache.Uri)
	if err != nil {
		panic(err)
	}

	options := cache.CacheOptions{
		Guilds:      true,
		Users:       true,
		Members:     true,
		Channels:    true,
		Roles:       true,
		Emojis:      false,
		VoiceStates: false,
	}

	return cache.PgCacheFactory(pool, options)
}
