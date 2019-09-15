package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/bot/command"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/robfig/go-cache"
	"io/ioutil"
	"net/http"
	"time"
)

type ProxyResponse struct {
	Premium bool
	Tier int
}

var premiumCache = cache.New(10 * time.Minute, 10 * time.Minute)

func IsPremiumGuild(ctx command.CommandContext, ch chan bool) {
	if premium, ok := premiumCache.Get(ctx.Guild); ok {
		ch<-premium.(bool)
		return
	}

	// First lookup by premium key, then patreon
	keyLookup := make(chan bool)
	go database.IsPremium(ctx.GuildId, keyLookup)

	if <-keyLookup {
		ch<-true
	} else {
		guild, err := ctx.Session.State.Guild(ctx.Guild); if err != nil {
			guild, err = ctx.Session.Guild(ctx.Guild); if err != nil {
				sentry.Error(err)
				ch<-false
				return
			}
		}

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

		ch<-proxyResponse.Premium
	}
}
