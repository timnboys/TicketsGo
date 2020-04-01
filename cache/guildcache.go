package cache

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/guild"
)

func (c *RedisClient) CacheGuildProperties(guild *guild.Guild) {
	key := fmt.Sprintf("tickets:guilds:%s", guild.Id)
	c.HSet(key, "owner", guild.OwnerId)
	c.HSet(key, "name", guild.Name)
}

func (c *RedisClient) GetGuildOwner(id string, res chan string) {
	key := fmt.Sprintf("tickets:guilds:%s", id)
	response := c.HGet(key, "owner")

	// TODO: Publish message to shard to repopulate cache if not present
	if response.Err() != nil {
		sentry.Error(response.Err())
		res <- ""
		return
	}

	res <- response.Val()
}

func (c *RedisClient) GetGuildName(id string, res chan string) {
	key := fmt.Sprintf("tickets:guilds:%s", id)
	response := c.HGet(key, "name")

	// TODO: Publish message to shard to repopulate cache if not present
	if response.Err() != nil {
		sentry.Error(response.Err())
		res <- ""
		return
	}

	res <- response.Val()
}

