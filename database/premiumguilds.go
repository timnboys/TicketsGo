package database

import (
	"strings"
	"time"
)

type PremiumGuilds struct {
	GuildId int64 `gorm:"column:GUILDID;unique;primary_key"`
	Expiry int64 `gorm:"column:EXPIRY"`
	ActivatedBy int64 `gorm:"column:ACTIVATEDBY"`
	Keys string `gorm:"column:KEYSUSED"`
}

func (PremiumGuilds) TableName() string {
	return "premiumguilds"
}

func IsPremium(guild int64, ch chan bool) {
	var node PremiumGuilds
	Db.Where(PremiumGuilds{GuildId: guild}).First(&node)

	if node.Expiry == 0 {
		ch <- false
		return
	}

	current := time.Now().UnixNano() / int64(time.Millisecond)
	ch <- node.Expiry > current
}

func AddPremium(key string, guild int64, length int64, activatedBy int64) {
	var expiry int64

	hasPrem := make(chan bool)
	IsPremium(guild, hasPrem)
	isPremium := <- hasPrem

	if isPremium {
		expiryChan := make(chan int64)
		GetExpiry(guild, expiryChan)
		currentExpiry := <- expiryChan

		expiry = currentExpiry + length
	} else {
		current := time.Now().UnixNano() / int64(time.Millisecond)
		expiry = current + length
	}

	keysChan := make(chan []string)
	GetKeysUsed(guild, keysChan)
	keys := <- keysChan
	keys = append(keys, key)
	keysStr := strings.Join(keys,",")

	var node PremiumGuilds
	Db.Where(PremiumGuilds{GuildId: guild}).Assign(PremiumGuilds{Expiry: expiry, ActivatedBy: activatedBy, Keys: keysStr}).FirstOrCreate(&node)
}

func GetExpiry(guild int64, ch chan int64) {
	var node PremiumGuilds
	Db.Where(PremiumGuilds{GuildId: guild}).First(&node)
	ch <- node.Expiry
}

func GetKeysUsed(guild int64, ch chan []string) {
	var node PremiumGuilds
	Db.Where(PremiumGuilds{GuildId: guild}).First(&node)
	ch <- strings.Split(node.Keys, ",")
}
