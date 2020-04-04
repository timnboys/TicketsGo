package utils

import (
	"encoding/json"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/jwangsadinata/go-multimap/slicemultimap"
	gocache "github.com/patrickmn/go-cache"
	"github.com/rxdn/gdl/gateway"
	"strconv"
	"sync"
	"time"
)

type UserGuild struct {
	Id   uint64
	Name string
}

type mutualGuildResponse struct {
	UserId uint64
	Shard  int
	Guilds []UserGuild
}

const timeout = 4 * time.Second

var guildCache = gocache.New(time.Minute, time.Minute)

func GetMutualGuilds(userId uint64, res chan []UserGuild) {
	// Check cache
	key := strconv.FormatUint(userId, 10)
	cached, ok := guildCache.Get(key)
	if ok {
		res <- cached.([]UserGuild)
		return
	} else {
		totalShards := config.Conf.Bot.Sharding.Total
		ch := make(chan mutualGuildResponse)

		callbackLock.Lock()
		callbackMap.Put(userId, ch)
		callbackLock.Unlock()

		guilds := make([]UserGuild, 0)
		guildsLock := sync.Mutex{}

		wg := sync.WaitGroup{}
		wg.Add(totalShards)

		go func() {
			for res := range ch {
				guildsLock.Lock()
				guilds = append(guilds, res.Guilds...)
				guildsLock.Unlock()

				wg.Done()
			}
		}()

		publishGuildRequest(userId)
		CountdownWithTimeout(&wg, timeout)

		callbackLock.Lock()
		callbackMap.Remove(userId, ch)
		callbackLock.Unlock()

		// Sort by guild ID
		sorted := make([]UserGuild, 0)
		for _, guild := range guilds {
			max := uint64(0)

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
					sorted = append(sorted, guild)
				}
			}
		}

		guildCache.Set(key, sorted, gocache.DefaultExpiration)
		res <- sorted
	}
}

func publishGuildRequest(userId uint64) {
	cache.Client.Publish("tickets:getuserguilds", strconv.Itoa(int(userId)))
}

func publishUserGuilds(userId uint64, shard int, guilds []UserGuild) {
	response := mutualGuildResponse{
		UserId: userId,
		Shard:  shard,
		Guilds: guilds,
	}

	encoded, err := json.Marshal(&response)
	if err != nil {
		sentry.Error(err)
		return
	}

	cache.Client.Publish("tickets:userguilds", string(encoded))
}

func ListenMutualGuildRequests(manager *gateway.ShardManager) {
	pubsub := cache.Client.Subscribe("tickets:getuserguilds")

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			sentry.Error(err)
			continue
		}

		userId, err := strconv.ParseUint(msg.Payload, 10, 64); if err != nil {
			sentry.Error(err)
			continue
		}

		for _, shard := range manager.Shards {
			guilds := make([]UserGuild, 0)

			// Loop over guilds managed by shard
			for _, guild := range shard.Cache.GetGuilds() {
				// Verify that the user is a member of the guild
			memberloop:
				for _, member := range guild.Members {
					if strconv.FormatUint(member.User.Id, 10) == msg.Payload {
						guilds = append(guilds, UserGuild{
							Id:   guild.Id,
							Name: guild.Name,
						})

						break memberloop
					}
				}
			}

			go publishUserGuilds(userId, shard.ShardId, guilds)
		}
	}
}

var(
	callbackLock sync.Mutex
	callbackMap  = slicemultimap.New()
)

func ListenMutualGuildResponses() {
	pubsub := cache.Client.Subscribe("tickets:userguilds")

	for {
		msg, err := pubsub.ReceiveMessage()
		if err != nil {
			sentry.Error(err)
			continue
		}

		var response mutualGuildResponse
		if err := json.Unmarshal([]byte(msg.Payload), &response); err != nil {
			sentry.Error(err)
			continue
		}

		callbackLock.Lock()
		callbacks, found := callbackMap.Get(response.UserId)
		if found {
			for _, callback := range callbacks {
				ch := callback.(chan mutualGuildResponse)
				ch <- response
			}
		}
		callbackLock.Unlock()
	}
}
