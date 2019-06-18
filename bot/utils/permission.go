package utils

import (
	"github.com/bwmarrin/discordgo"
)

type Permission int

const (
	CreateInstantInvite Permission = 0x00000001
	KickMembers         Permission = 0x00000002
	BanMembers          Permission = 0x00000004
	Administrator       Permission = 0x00000008
	ManageChannels      Permission = 0x00000010
	ManageGuild         Permission = 0x00000020
	AddReactions        Permission = 0x00000040
	ViewAuditLog        Permission = 0x00000080
	ViewChannel         Permission = 0x00000400 // Read messages
	SendMessages        Permission = 0x00000800
	SendTTSMessages     Permission = 0x00001000
	ManageMessages      Permission = 0x00002000
	EmbedLinks          Permission = 0x00004000
	AttachFiles         Permission = 0x00008000
	ReadMessageHistory  Permission = 0x00010000
	MentionEveryone     Permission = 0x00020000
	Connect             Permission = 0x00100000
	Speak               Permission = 0x00200000
	MuteMembers         Permission = 0x00400000
	DeafenMembers       Permission = 0x00800000
	MoveMembers         Permission = 0x01000000
	UseVAD              Permission = 0x02000000 // Use voice activity
	PrioritySpeaker     Permission = 0x00000100
	ChangeNickname      Permission = 0x04000000
	ManageNicknames     Permission = 0x08000000
	ManageRoles         Permission = 0x10000000 // Manage permissions
	ManageWebhooks      Permission = 0x20000000
	ManageEmojis        Permission = 0x40000000
)

func hasPermission(perms int, permission Permission) bool {
	return perms&int(permission) != 0
}

func MemberHasPermission(session *discordgo.Session, guild string, user string, perm Permission, ch chan bool) {
	m, err := session.State.Member(guild, user); if err != nil {
		// Not cached
		m, err = session.GuildMember(guild, user); if err != nil {
			ch <- false
			return
		}
	}

	for _, role := range m.Roles {
		cb := make(chan bool)
		go RoleHasPermission(session, guild, role, perm, cb)
		if <-cb {
			ch <- true
			return
		}
	}
	ch <- false
}

func ChannelMemberHasPermission(session *discordgo.Session, guild, channel, user string, perm Permission, ch chan bool) {
	member, err := session.State.Member(guild, user); if err != nil {
		// Not cached
		member, err = session.GuildMember(guild, user); if err != nil {
			ch <- false
			return
		}
	}

	c, err := session.State.Channel(channel); if err != nil {
		// Not cached
		c, err = session.Channel(channel); if err != nil {
			ch <- false
			return
		}
	}

	for _, overwrite := range c.PermissionOverwrites {
		if overwrite.ID == user {
			// Check if the role is denied this permission
			if hasPermission(overwrite.Deny, perm) {
				ch <- false
				return
			}

			// Check if the role is allowed this permission in the overwrite
			if hasPermission(overwrite.Allow, perm) {
				ch <- true
				return
			}
		}
	}

	for _, role := range member.Roles {
		hasPerm := make(chan bool)
		go ChannelRoleHasPermission(session, guild, channel, role, perm, hasPerm)
		if <-hasPerm {
			ch <- true
			return
		}
	}

	ch <- false
}

func ChannelRoleHasPermission(session *discordgo.Session, guild, channel, role string, perm Permission, ch chan bool) {
	// Retrieve the overwrite
	c, err := session.State.Channel(channel); if err != nil {
		// Not cached
		c, err = session.Channel(channel); if err != nil {
			ch <- false
			return
		}
	}

	var o *discordgo.PermissionOverwrite = nil
	for _, overwrite := range c.PermissionOverwrites {
		if overwrite.ID == role { // We don't need to check the type because a role and user won't have the same snowflake
			o = overwrite
		}
	}

	// If there's an overwrite, check the overwritten perms
	if o != nil {
		// Check if the role is denied this permission
		if hasPermission(o.Deny, perm) {
			ch <- false
			return
		}

		// Check if the role is allowed this permission in the overwrite
		if hasPermission(o.Allow, perm) {
			ch <- true
			return
		}
	}

	// Else check if they're allowed it at a guild level
	guildPermsChan := make(chan int)
	go GetRolePermissions(session, guild, role, guildPermsChan)

	ch <- hasPermission(<- guildPermsChan, perm)
	return
}

func RoleHasPermission(session *discordgo.Session, guild string, role string, perm Permission, ch chan bool) {
	r, err := session.State.Role(guild, role); if err != nil {
		// Not cached
		roles, err := session.GuildRoles(guild); if err != nil {
			ch <- false
			return
		}

		found := false
		for _, t := range roles {
			if t.ID == role {
				r = t
				found = true
				break
			}
		}
		if !found {
			ch <- false
			return
		}
	}

	ch <- hasPermission(r.Permissions, perm)
}

func GetRolePermissions(session *discordgo.Session, guild string, role string, ch chan int) {
	r, err := session.State.Role(guild, role); if err != nil {
		ch <- 0
		return
	}

	ch <- r.Permissions
}

func SumPermissions(perms ...Permission) int {
	sum := 0
	for _, perm := range perms {
		sum |= int(perm)
	}
	return sum
}
