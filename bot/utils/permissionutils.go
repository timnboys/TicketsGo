package utils

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/patrickmn/go-cache"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/permission"
	"strconv"
	"time"
)

// Snowflake -> PermissionLevel
var cacheTime = time.Minute * 2
var permissionCache = cache.New(time.Minute * 2, cacheTime)

func GetPermissionLevel(shard *gateway.Shard, member *member.Member, guildId uint64, ch chan PermissionLevel) {
	// Check if the user is a bot adminUser
	for _, admin := range config.Conf.Bot.Admins {
		if admin == member.User.Id {
			ch <- Admin
			return
		}
	}

	// Check if user is guild owner
	guild, err := shard.GetGuild(guildId)
	if err != nil {
		sentry.ErrorWithContext(err, sentry.ErrorContext{
			Guild: guildId,
			User:  member.User.Id,
			Shard: shard.ShardId,
		})
	}

	if err == nil {
		if member.User.Id == guild.OwnerId {
			ch <- Admin
			return
		}
	}

	// Check user ID in cache
	if cached, ok := permissionCache.Get(getMemberId(guildId, member)); ok {
		ch <- cached.(PermissionLevel)
		return
	}

	// Check user perms for admin
	adminUser := make(chan bool)
	go database.IsAdmin(guildId, member.User.Id, adminUser)
	if <-adminUser {
		permissionCache.Set(getMemberId(guildId, member), Admin, cacheTime)
		ch <- Admin
		return
	}

	// Check role perms for admin
	// Check cached roles
	for _, userRole := range member.Roles {
		if permLevel, ok := permissionCache.Get(strconv.FormatUint(userRole, 10)); ok { // TODO: Cache with int based key
			if permLevel == Admin {
				permissionCache.Set(getMemberId(guildId, member), Admin, cacheTime)
				ch <- Admin
				return
			}
		}
	}

	// Check roles from DB
	adminRolesChan := make(chan []uint64)
	go database.GetAdminRoles(guildId, adminRolesChan)
	adminRoles := <-adminRolesChan
	for _, adminRoleId := range adminRoles {
		permissionCache.Set(strconv.FormatUint(adminRoleId, 10), Admin, cacheTime) // TODO: Cache with int based key

		if member.HasRole(adminRoleId) {
			permissionCache.Set(getMemberId(guildId, member), Admin, cacheTime)
			ch <- Admin
			return
		}
	}

	// Check if user has Administrator permission
	hasAdminPermission := permission.HasPermissions(shard, guildId, member.User.Id, permission.Administrator)
	if hasAdminPermission {
		permissionCache.Set(getMemberId(guildId, member), Admin, cacheTime)
		ch <- Admin
		return
	}

	// Check user perms for support
	supportUser := make(chan bool)
	go database.IsSupport(guildId, member.User.Id, supportUser)
	if <-supportUser {
		permissionCache.Set(getMemberId(guildId, member), Support, cacheTime)
		ch <- Support
		return
	}

	// Check role perms for support
	// Check cached roles
	for _, userRole := range member.Roles {
		if permLevel, ok := permissionCache.Get(strconv.FormatUint(userRole, 10)); ok { // TODO: Cache with int based key
			if permLevel == Support {
				permissionCache.Set(getMemberId(guildId, member), Support, cacheTime)
				ch <- Support
				return
			}
		}
	}

	// Check DB
	supportRolesChan := make(chan []uint64)
	go database.GetSupportRoles(guildId, supportRolesChan)
	supportRoles := <-supportRolesChan
	for _, supportRoleId := range supportRoles {
		permissionCache.Set(strconv.FormatUint(supportRoleId, 10), Support, cacheTime)

		if member.HasRole(supportRoleId) {
			permissionCache.Set(getMemberId(guildId, member), Support, cacheTime)
			ch <- Support
			return
		}
	}

	permissionCache.Set(getMemberId(guildId, member), Everyone, cacheTime)
	ch <- Everyone
}

func getMemberId(guildId uint64, member *member.Member) string {
	return fmt.Sprintf("%d-%d", guildId, member.User.Id)
}
