package cache

import (
	"fmt"
	"strconv"
	"time"
)

func (c *RedisClient) IsPremium(guildId uint64) (bool, error) {
	key := fmt.Sprintf("premium:%d", guildId)

	res, err := c.Get(key).Result()
	if err != nil {
		return false, err
	}

	return res == "true", nil
}

func (c *RedisClient) SetPremium(guildId uint64, premiumStatus bool) {
	key := fmt.Sprintf("premium:%d", guildId)

	c.Set(key, strconv.FormatBool(premiumStatus), time.Minute * 10)
}
