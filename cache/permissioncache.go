package cache

import (
	"fmt"
	"strconv"
	"time"
)

func (c *RedisClient) GetPermissionLevel(guildId, userId uint64) (int, error) {
	key := fmt.Sprintf("permissions:%d:%d", guildId, userId)

	res, err := c.Get(key).Result()
	if err != nil {
		return 0, err
	}

	levelId, err := strconv.Atoi(res)
	if err != nil {
		return 0, err
	}

	return levelId, nil
}

func (c *RedisClient) SetPermissionLevel(guildId, userId uint64, level int) {
	key := fmt.Sprintf("permissions:%d:%d", guildId, userId)
	c.Set(key, level, time.Minute * 10)
}

func (c *RedisClient) DeletePermissionLevel(guildId, userId uint64) {
	key := fmt.Sprintf("permissions:%d:%d", guildId, userId)
	c.Del(key)
}

