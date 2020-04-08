package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/patrickmn/go-cache"
	"github.com/rxdn/gdl/gateway"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type ProxyResponse struct {
	Premium bool
	Tier    int
}

const cacheTimeout = 10 * time.Minute

var premiumCache = cache.New(cacheTimeout, cacheTime)

func IsPremiumGuild(s *gateway.Shard, guildId uint64, ch chan bool) {
	guildStr := strconv.FormatUint(guildId, 10)
	premium, ok := premiumCache.Get(guildStr)

	if ok {
		ch <- premium.(bool)
		return
	}

	// First lookup by premium key, then votes, then patreon
	keyLookup := make(chan bool)
	go database.IsPremium(guildId, keyLookup)

	if <-keyLookup {
		premiumCache.Set(guildStr, true, cacheTimeout)
		ch <- true
	} else {
		// get guild owner
		guild, err := s.GetGuild(guildId); if err != nil {
			sentry.Error(err)
			ch <- false
			return
		}

		// Lookup votes
		hasVoted := make(chan bool)
		go database.HasVoted(guild.OwnerId, hasVoted)
		if <-hasVoted {
			premiumCache.Set(guildStr, true, cacheTimeout)
			ch <- true
			return
		}

		// Lookup Patreon
		client := &http.Client{
			Timeout: time.Second * 3,
		}

		url := fmt.Sprintf("%s/ispremium?key=%s&id=%d", config.Conf.Bot.PremiumLookupProxyUrl, config.Conf.Bot.PremiumLookupProxyKey, guild.OwnerId)
		req, err := http.NewRequest("GET", url, nil)

		res, err := client.Do(req)
		if err != nil {
			sentry.Error(err)
			ch <- false
			return
		}
		defer res.Body.Close()

		content, err := ioutil.ReadAll(res.Body)
		if err != nil {
			sentry.Error(err)
			ch <- false
			return
		}

		var proxyResponse ProxyResponse
		if err = json.Unmarshal(content, &proxyResponse); err != nil {
			sentry.Error(err)
			ch <- false
			return
		}

		premiumCache.Set(guildStr, proxyResponse.Premium, cacheTimeout)
		ch <- proxyResponse.Premium
	}
}

func CacheGuildAsPremium(guildId uint64) {
	premiumCache.Set(strconv.FormatUint(guildId, 10), true, cacheTimeout)
}
