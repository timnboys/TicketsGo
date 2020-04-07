package servercounter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/go-errors/errors"
	"github.com/rxdn/gdl/gateway"
	"io/ioutil"
	"net/http"
	"time"
)

type ServerCount struct {
	Key string `json:"key"`
	Shard int `json:"shard"`
	ServerCount int `json:"serverCount"`
}


func UpdateServerCount(shardManager *gateway.ShardManager) {
	for _, shard := range shardManager.Shards {
		client := http.Client{
			Timeout: 5 * time.Second,
		}

		data := ServerCount{
			Key: config.Conf.ServerCounter.Key,
			Shard: shard.ShardId,
			ServerCount: len(shard.GetShardGuildIds()),
		}

		encoded, err := json.Marshal(data); if err != nil {
			sentry.Error(err)
			return
		}

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/update", config.Conf.ServerCounter.BaseUrl), bytes.NewBuffer(encoded)); if err != nil {
			sentry.Error(err)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		res, err := client.Do(req); if err != nil {
			sentry.Error(err)
			return
		}

		if res.StatusCode != 200 {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				sentry.Error(err)
			} else {
				sentry.Error(errors.New(body))
			}
		}

		res.Body.Close()
	}
}
