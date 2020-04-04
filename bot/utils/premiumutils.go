package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type ProxyResponse struct {
	Premium bool
	Tier int
}

var premiumCache = cache.New(10 * time.Minute, 10 * time.Minute)

func IsPremiumGuild(ctx CommandContext, ch chan bool) {
	guildStr := strconv.FormatUint(ctx.Guild.Id, 10)
	premium, ok := premiumCache.Get(guildStr)

	if ok {
		ch<-premium.(bool)
		return
	}

	// First lookup by premium key, then votes, then patreon
	keyLookup := make(chan bool)
	go database.IsPremium(ctx.Guild.Id, keyLookup)

	if <-keyLookup {
		premiumCache.Set(guildStr, true, 10 * time.Minute)
		ch<-true
	} else {
		// Lookup votes
		hasVoted := make(chan bool)
		go database.HasVoted(ctx.Guild.OwnerId, hasVoted)
		if <-hasVoted {
			premiumCache.Set(guildStr, true, 10 * time.Minute)
			ch <- true
			return
		}

		// Lookup Patreon
		client := &http.Client{
			Timeout: time.Second * 3,
		}

		url := fmt.Sprintf("%s/ispremium?key=%s&id=%d", config.Conf.Bot.PremiumLookupProxyUrl, config.Conf.Bot.PremiumLookupProxyKey, ctx.Guild.OwnerId)
		req, err := http.NewRequest("GET", url, nil)

		res, err := client.Do(req); if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ch<-false
			return
		}
		defer res.Body.Close()

		content, err := ioutil.ReadAll(res.Body); if err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ch<-false
			return
		}

		var proxyResponse ProxyResponse
		if err = json.Unmarshal(content, &proxyResponse); err != nil {
			sentry.ErrorWithContext(err, ctx.ToErrorContext())
			ch<-false
			return
		}

		premiumCache.Set(guildStr, proxyResponse.Premium, 10 * time.Minute)
		ch <-proxyResponse.Premium
	}
}

func CacheGuildAsPremium(guildId uint64) {
	premiumCache.Set(strconv.FormatUint(guildId, 10), true, 10 * time.Minute)
}
