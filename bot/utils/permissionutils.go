package utils

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/bwmarrin/discordgo"
	"github.com/robfig/go-cache"
	"strconv"
	"time"
)

// Snowflake -> PermissionLevel
var cacheTime = time.Minute * 2
var permissionCache = cache.New(time.Minute * 2, cacheTime)

func GetPermissionLevel(session *discordgo.Session, member *discordgo.Member, ch chan PermissionLevel) {
	// Check if the user is a bot adminUser
	for _, admin := range config.Conf.Bot.Admins {
		if admin == member.User.ID {
			ch <- Admin
			return
		}
	}

	// Check if user is guild owner
	g, err := session.State.Guild(member.GuildID); if err != nil {
		// Not cached
		g, err = session.Guild(member.GuildID)
		if err != nil {
			sentry.ErrorWithContext(err, sentry.ErrorContext{
				Guild: member.GuildID,
				User:  member.User.ID,
				Shard: session.ShardID,
			})
		}
	}

	if g != nil {
		if member.User.ID == g.OwnerID {
			ch <- Admin
			return
		}
	}

	// Check user ID in cache
	if cached, ok := permissionCache.Get(getMemberId(member)); ok {
		ch <- cached.(PermissionLevel)
		return
	}

	// Check user perms for admin
	adminUser := make(chan bool)
	go database.IsAdmin(member.GuildID, member.User.ID, adminUser)
	if <-adminUser {
		ch <- Admin
		permissionCache.Set(getMemberId(member), Admin, cacheTime)
		return
	}

	// Check role perms for admin
	// Check cached roles
	for _, userRole := range member.Roles {
		if permLevel, ok := permissionCache.Get(userRole); ok {
			if permLevel == Admin {
				ch <- Admin
				permissionCache.Set(getMemberId(member), Admin, cacheTime)
				return
			}
		}
	}

	// Check roles from DB
	adminRolesChan := make(chan []int64)
	go database.GetAdminRoles(member.GuildID, adminRolesChan)
	adminRoles := <-adminRolesChan
	for _, adminRoleId := range adminRoles {
		adminRoleIdStr := strconv.Itoa(int(adminRoleId))
		permissionCache.Set(adminRoleIdStr, Admin, cacheTime)

		for _, userRoleId := range member.Roles {
			if adminRoleIdStr == userRoleId {
				ch <- Admin
				permissionCache.Set(getMemberId(member), Admin, cacheTime)
				return
			}
		}
	}

	// Check if user has Administrator permission
	adminPermissionChan := make(chan bool)
	go MemberHasPermission(session, member.GuildID, member.User.ID, Administrator, adminPermissionChan)
	hasAdminPermission := <-adminPermissionChan

	if hasAdminPermission {
		ch <- Admin
		permissionCache.Set(getMemberId(member), Admin, cacheTime)
		return
	}

	// Check user perms for support
	supportUser := make(chan bool)
	go database.IsSupport(member.GuildID, member.User.ID, supportUser)
	if <-supportUser {
		ch <- Support
		permissionCache.Set(getMemberId(member), Support, cacheTime)
		return
	}

	// Check role perms for support
	// Check cached roles
	for _, userRole := range member.Roles {
		if permLevel, ok := permissionCache.Get(userRole); ok {
			if permLevel == Support {
				ch <- Support
				permissionCache.Set(getMemberId(member), Support, cacheTime)
				return
			}
		}
	}

	// Check DB
	supportRolesChan := make(chan []int64)
	go database.GetSupportRoles(member.GuildID, supportRolesChan)
	supportRoles := <-supportRolesChan
	for _, supportRoleId := range supportRoles {
		supportRoleIdStr := strconv.Itoa(int(supportRoleId))
		permissionCache.Set(supportRoleIdStr, Support, cacheTime)

		for _, userRoleId := range member.Roles {
			if supportRoleIdStr == userRoleId {
				ch <- Support
				permissionCache.Set(getMemberId(member), Support, cacheTime)
				return
			}
		}
	}

	ch <- Everyone
	permissionCache.Set(getMemberId(member), Everyone, cacheTime)
}

func getMemberId(member *discordgo.Member) string {
	return fmt.Sprintf("%s-%s", member.GuildID, member.User.ID)
}
