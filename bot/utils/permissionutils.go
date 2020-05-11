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
		ch <- PermissionLevel(cached)
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
			go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin.Int())
			ch <- Admin
			return
		}
	}

	// Check user perms for admin
	adminUser, err := database.Client.Permissions.IsAdmin(guildId, member.User.Id); if err != nil {
		sentry.Error(err)
	}

	if adminUser {
		go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin.Int())
		ch <- Admin
		return
	}

	// Check roles from DB
	adminRoles, err := database.Client.RolePermissions.GetAdminRoles(guildId); if err != nil {
		sentry.Error(err)
	}

	for _, adminRoleId := range adminRoles {
		if member.HasRole(adminRoleId) {
			go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin.Int())
			ch <- Admin
			return
		}
	}

	// Check if user has Administrator permission
	hasAdminPermission := permission.HasPermissions(shard, guildId, member.User.Id, permission.Administrator)
	if hasAdminPermission {
		go cache.Client.SetPermissionLevel(guildId, member.User.Id, Admin.Int())
		ch <- Admin
		return
	}

	// Check user perms for support
	isSupport, err := database.Client.Permissions.IsSupport(guildId, member.User.Id); if err != nil {
		sentry.Error(err)
	}

	if isSupport {
		go cache.Client.SetPermissionLevel(guildId, member.User.Id, Support.Int())
		ch <- Support
		return
	}

	// Check DB for support roles
	supportRoles, err :=  database.Client.RolePermissions.GetSupportRoles(guildId); if err != nil {
		sentry.Error(err)
	}

	for _, supportRoleId := range supportRoles {
		if member.HasRole(supportRoleId) {
			go cache.Client.SetPermissionLevel(guildId, member.User.Id, Support.Int())
			ch <- Support
			return
		}
	}

	go cache.Client.SetPermissionLevel(guildId, member.User.Id, Everyone.Int())
	ch <- Everyone
}
