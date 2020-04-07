package utils

import (
	"context"
	"errors"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/cache"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/guild"
)

type UserGuild struct {
	Id   uint64
	Name string
}

func GetMutualGuilds(shard *gateway.Shard, userId uint64) []guild.Guild {
	cache, ok := shard.Cache.(*cache.PgCache)
	if !ok {
		sentry.Error(errors.New("failed to cast cache to PgCache"))
		return nil
	}

	var guildIds []uint64
	rows, err := cache.Query(context.Background(), `SELECT "guild_id" FROM members WHERE "user_id" = $1;`, userId)
	defer rows.Close()
	if err != nil {
		sentry.Error(err)
		return nil
	}

	for rows.Next() {
		var guildId uint64
		if err := rows.Scan(&guildId); err != nil {
			sentry.Error(err)
			continue
		}

		guildIds = append(guildIds, guildId)
	}

	var guilds []guild.Guild
	for _, guildId := range guildIds {
		guild, err := shard.GetGuild(guildId); if err != nil {
			sentry.Error(err)
			continue
		}

		guilds = append(guilds, guild)
	}

	return guilds
}
