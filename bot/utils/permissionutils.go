package utils

import (
	"github.com/TicketsBot/TicketsGo/cache"
	"github.com/TicketsBot/TicketsGo/config"
	"github.com/TicketsBot/TicketsGo/database"
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/rxdn/gdl/gateway"
	"github.com/rxdn/gdl/objects/member"
	"github.com/rxdn/gdl/permission"
)

func GetPermissionLevel(shard *gateway.Shard, member member.Member, guildId uint64, ch chan PermissionLevel) {
	// Check user ID in cache
	if cached, err := cache.Client.GetPermissionLevel(guildId, member.User.Id); err == nil {
		ch <- cached
		return
	}

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
			go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin)
			ch <- Admin
			return
		}
	}

	// Check user perms for admin
	adminUser := make(chan bool)
	go database.IsAdmin(guildId, member.User.Id, adminUser)
	if <-adminUser {
		go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin)
		ch <- Admin
		return
	}

	// Check roles from DB
	adminRolesChan := make(chan []uint64)
	go database.GetAdminRoles(guildId, adminRolesChan)
	adminRoles := <-adminRolesChan
	for _, adminRoleId := range adminRoles {
		if member.HasRole(adminRoleId) {
			go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin)
			ch <- Admin
			return
		}
	}

	// Check if user has Administrator permission
	hasAdminPermission := permission.HasPermissions(shard, guildId, member.User.Id, permission.Administrator)
	if hasAdminPermission {
		go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin)
		ch <- Admin
		return
	}

	// Check user perms for support
	supportUser := make(chan bool)
	go database.IsSupport(guildId, member.User.Id, supportUser)
	if <-supportUser {
		go cache.Client.SetPermissionLevel(guildId, member.User.Id, Support)
		ch <- Support
		return
	}

	// Check DB for support roles
	supportRolesChan := make(chan []uint64)
	go database.GetSupportRoles(guildId, supportRolesChan)
	supportRoles := <-supportRolesChan
	for _, supportRoleId := range supportRoles {
		if member.HasRole(supportRoleId) {
			go cache.Client.SetPermissionLevel(guildId, member.User.Id, Support)
			ch <- Support
			return
		}
	}

	go cache.Client.SetPermissionLevel(guildId, member.User.Id, Everyone)
	ch <- Everyone
}
