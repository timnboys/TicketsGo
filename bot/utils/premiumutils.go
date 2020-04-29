package utils

import (
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"io/ioutil"
	"net/http"
	"time"
)

type ProxyResponse struct {
	Premium bool
	Tier    int
}

func IsPremiumGuild(s *gateway.Shard, guildId uint64, ch chan bool) {
	// we can block for this
	if cached, err := cache.Client.IsPremium(guildId); err == nil {
		ch <- cached
		return
	}

	// Patreon -> Key -> Votes

	// Lookup Patreon
	// get guild owner
	guild, err := s.GetGuild(guildId)
	if err != nil {
		sentry.Error(err)
		ch <- false
		return
	}

	client := &http.Client{
		Timeout: time.Second * 3,
	}

	url := fmt.Sprintf("%s/ispremium?key=%s&id=%d", config.Conf.Bot.PremiumLookupProxyUrl, config.Conf.Bot.PremiumLookupProxyKey, guild.OwnerId)
	res, err := client.Get(url)
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

	if proxyResponse.Premium {
		go cache.Client.SetPremium(guildId, proxyResponse.Premium)
		ch <- proxyResponse.Premium
		return
	}

	// Lookup key
	keyLookup := make(chan bool)
	go database.IsPremium(guildId, keyLookup)
	if <-keyLookup {
		go cache.Client.SetPremium(guildId, true)
		ch <- true
	}

	// Lookup votes
	hasVoted := make(chan bool)
	go database.HasVoted(guild.OwnerId, hasVoted)
	if <-hasVoted {
		go cache.Client.SetPremium(guildId, true)
		ch <- true
		return
	}

	go cache.Client.SetPremium(guildId, false)
	ch <- false
}
