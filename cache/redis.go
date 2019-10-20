package cache

import (
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/go-redis/redis"
)

type RedisClient struct {
	*redis.Client
}

var Client RedisClient

func NewRedisClient() RedisClient {
	uri := ParseURI(config.Conf.Redis.Uri)

	client := redis.NewClient(&redis.Options{
		Addr:     uri.Addr,
		Password: uri.Password,
		PoolSize: config.Conf.Redis.Threads,
	})

	return RedisClient{
		client,
	}
}

func (c *RedisClient) IsConnected(ch chan bool) {
	ch <- c.Conn().Ping().Err() == nil
}
