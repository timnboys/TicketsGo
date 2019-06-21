package servercounter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/apex/log"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type ServerCount struct {
	Key string `json:"key"`
	Shard int `json:"shard"`
	ServerCount int `json:"serverCount"`
}

var(
	serverCountCache = make(map[int]int)
	cacheLock sync.Mutex
)

func UpdateServerCount() {
	cacheLock.Lock()
	clone := copyCache()
	cacheLock.Unlock()

	for shard, count := range clone {
		client := http.Client{
			Timeout: 5 * time.Second,
		}

		data := ServerCount{
			Key: config.Conf.ServerCounter.Key,
			Shard: shard,
			ServerCount: count,
		}

		encoded, err := json.Marshal(data); if err != nil {
			log.Error(err.Error())
			return
		}

		req, err := http.NewRequest(fmt.Sprintf("%s/update", config.Conf.ServerCounter.BaseUrl), "application/json", bytes.NewBuffer(encoded)); if err != nil {
			log.Error(err.Error())
			return
		}

		res, err := client.Do(req); if err != nil {
			log.Error(err.Error())
			return
		}

		if res.StatusCode != 200 {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				log.Error(err.Error())
			} else {
				log.Error(string(body))
			}
		}

		res.Body.Close()
	}
}

func UpdateCache(shard, count int) {
	cacheLock.Lock()
	serverCountCache[shard] = count
	cacheLock.Unlock()
}

func copyCache() map[int]int {
	clone := make(map[int]int)

	for k, v := range serverCountCache {
		clone[k] = v
	}

	return clone
}
