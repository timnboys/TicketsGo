package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/go-redis/redis"
	"github.com/jonas747/dshardmanager"
	gocache "github.com/patrickmn/go-cache"
	"strconv"
	"sync"
	"time"
)

type UserGuild struct {
	Shard int
	Id    int64
	Name  string
}

const timeout = 5 * time.Second

var guildCache = gocache.New(time.Minute, time.Minute)

func GetMutualGuilds(userId int64, res chan []UserGuild) {
	// Check cache
	key := strconv.Itoa(int(userId))
	cached, ok := guildCache.Get(key); if ok {
		res <- cached.([]UserGuild)
		return
	}

	shards := config.Conf.Bot.Sharding.Total
	publishGuildRequest(userId)

	var wg sync.WaitGroup
	for i := 0; i < shards; i++ {
		wg.Add(1)
	}

	guilds := make([]UserGuild, 0)
	var guildsLock sync.Mutex

	go func() {
		pubsub := cache.Client.Subscribe(fmt.Sprintf("tickets:userguilds:%d", userId))
		for i := 0; i < shards; i++ {
			msg, err := pubsub.ReceiveTimeout(timeout); if err != nil {
				sentry.Error(err)
				continue
			}

			switch msg := msg.(type) {
			case *redis.Message: {
				var shardGuilds []UserGuild
				if err := json.Unmarshal([]byte(msg.Payload), &shardGuilds); err != nil {
					sentry.Error(err)
					return
				}

				guildsLock.Lock()
				guilds = append(guilds, shardGuilds...)
				guildsLock.Unlock()

				wg.Done()
			}
			default:
				continue
			}
		}
	}()

	CountdownWithTimeout(&wg, timeout)

	// Sort by guild ID
	sorted := make([]UserGuild, len(guilds))
	for i := 0; i < len(guilds); i++ {
		max := int64(0)

		for _, guild := range guilds {
			if guild.Id > max {
				// Check that we haven't already added this guild
				isAlreadySorted := false
				for _, sortedGuild := range sorted {
					if sortedGuild.Id == guild.Id {
						isAlreadySorted = true
						break
					}
				}

				if !isAlreadySorted {
					sorted[i] = guild
				}
			}
		}
	}

	guildCache.Set(key, sorted, gocache.DefaultExpiration)
	res <- sorted
}

func publishGuildRequest(userId int64) {
	cache.Client.Publish("tickets:getuserguilds", strconv.Itoa(int(userId)))
}

func publishUserGuilds(userId int64, guilds []UserGuild) {
	encoded, err := json.Marshal(&guilds); if err != nil {
		sentry.Error(err)
		return
	}

	cache.Client.Publish(fmt.Sprintf("tickets:userguilds:%d", userId), string(encoded))
}

func ListenGuildRequests(manager *dshardmanager.Manager) {
	pubsub := cache.Client.Subscribe("tickets:getuserguilds")

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			sentry.Error(err)
			continue
		}

		guilds := make([]UserGuild, 0)
		for _, shard := range manager.Sessions {
			for _, guild := range shard.State.Guilds {
				memberloop:
				for _, member := range guild.Members {
					if member.User.ID == msg.Payload {
						guildId, err := strconv.ParseInt(guild.ID, 10, 64); if err != nil {
							sentry.Error(err)
							break memberloop
						}

						guilds = append(guilds, UserGuild{
							Shard: shard.ShardID,
							Id:    guildId,
							Name:  guild.Name,
						})

						break memberloop
					}
				}
			}
		}
	}
}
