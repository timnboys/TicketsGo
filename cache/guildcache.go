package cache

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
)

func (c *RedisClient) CacheGuildProperties(guild *discordgo.Guild) {
	key := fmt.Sprintf("tickets:guilds:%s", guild.ID)
	c.HSet(key, "owner", guild.OwnerID)
	c.HSet(key, "name", guild.Name)
}

func (c *RedisClient) GetGuildOwner(id string, res chan string) {
	key := fmt.Sprintf("tickets:guilds:%s", id)
	response := c.HGet(key, "owner")

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

	if response.Err() != nil {
		sentry.Error(response.Err())
		res <- ""
		return
	}

	res <- response.Val()
}

