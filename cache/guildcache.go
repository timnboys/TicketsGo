package cache

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/objects/guild"
	"strconv"
)

func (c *RedisClient) CacheGuildProperties(guild *guild.Guild) {
	key := fmt.Sprintf("tickets:guilds:%d", guild.Id)
	c.HSet(key, "owner", guild.OwnerId)
	c.HSet(key, "name", guild.Name)
}

func (c *RedisClient) GetGuildOwner(guildId uint64, res chan uint64) {
	key := fmt.Sprintf("tickets:guilds:%d", guildId)
	response := c.HGet(key, "owner")

	// TODO: Publish message to shard to repopulate cache if not present
	if response.Err() != nil {
		sentry.Error(response.Err())
		res <- 0
		return
	}

	parsed, err := strconv.ParseUint(response.Val(), 10, 64); if err != nil {
		sentry.Error(err)
		res <- 0
		return
	}

	res <- parsed
}

func (c *RedisClient) GetGuildName(guildId uint64, res chan string) {
	key := fmt.Sprintf("tickets:guilds:%s", guildId)
	response := c.HGet(key, "name")

	// TODO: Publish message to shard to repopulate cache if not present
	if response.Err() != nil {
		sentry.Error(response.Err())
		res <- ""
		return
	}

	res <- response.Val()
}

