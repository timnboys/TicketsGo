package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/robfig/go-cache"
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
	if premium, ok := premiumCache.Get(ctx.Guild); ok {
		ch<-premium.(bool)
		return
	}

	// First lookup by premium key, then votes, then patreon
	keyLookup := make(chan bool)
	go database.IsPremium(ctx.GuildId, keyLookup)

	if <-keyLookup {
		if err := premiumCache.Add(ctx.Guild, true, 10 * time.Minute); err != nil {
			sentry.Error(err)
		}

		ch<-true
	} else {
		// Get guild object
		guild, err := ctx.Session.State.Guild(ctx.Guild); if err != nil {
			guild, err = ctx.Session.Guild(ctx.Guild); if err != nil {
				sentry.Error(err)
				ch<-false
				return
			}
		}

		// Lookup votes
		ownerId, err := strconv.ParseInt(guild.OwnerID, 10, 64); if err != nil {
			sentry.Error(err)
			ch <- false
			return
		}

		hasVoted := make(chan bool)
		database.HasVoted(ownerId, hasVoted)
		if <-hasVoted {
			ch <- true

			if err := premiumCache.Add(ctx.Guild, true, 10 * time.Minute); err != nil {
				sentry.Error(err)
			}

			return
		}

		// Lookup Patreon
		client := &http.Client{
			Timeout: time.Second * 3,
		}

		url := fmt.Sprintf("%s/ispremium?key=%s&id=%s", config.Conf.Bot.PremiumLookupProxyUrl, config.Conf.Bot.PremiumLookupProxyKey, guild.OwnerID)
		req, err := http.NewRequest("GET", url, nil)

		res, err := client.Do(req); if err != nil {
			sentry.Error(err)
			ch<-false
			return
		}
		defer res.Body.Close()

		content, err := ioutil.ReadAll(res.Body); if err != nil {
			sentry.Error(err)
			ch<-false
			return
		}

		var proxyResponse ProxyResponse
		if err = json.Unmarshal(content, &proxyResponse); err != nil {
			sentry.Error(err)
			ch<-false
			return
		}

		if err := premiumCache.Add(ctx.Guild, proxyResponse.Premium, 10 * time.Minute); err != nil {
			sentry.Error(err)
		}
		ch <-proxyResponse.Premium
	}
}
