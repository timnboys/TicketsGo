package sentry

import (
	"encoding/json"
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
	"strconv"
)

type ErrorContext struct {
	Guild       string
	User        string
	Channel     string
	Shard       int
	Command     string
	Premium     bool
	Permissions []*discordgo.PermissionOverwrite
}

func Error(e error) {
	wrapped := errors.New(e)
	raven.Capture(ConstructErrorPacket(wrapped), nil)
}

func LogWithContext(e error, ctx ErrorContext) {
	perms, _ := json.Marshal(ctx.Permissions)

	wrapped := errors.New(e)
	raven.Capture(ConstructPacket(wrapped, raven.INFO), map[string]string{
		"guild":       ctx.Guild,
		"user":        ctx.User,
		"channel":     ctx.Channel,
		"shard":       strconv.Itoa(ctx.Shard),
		"command":     ctx.Command,
		"premium":     strconv.FormatBool(ctx.Premium),
		"permissions": string(perms),
	})
}

func ErrorWithContext(e error, ctx ErrorContext) {
	perms, _ := json.Marshal(ctx.Permissions)

	wrapped := errors.New(e)
	raven.Capture(ConstructErrorPacket(wrapped), map[string]string{
		"guild":       ctx.Guild,
		"user":        ctx.User,
		"channel":     ctx.Channel,
		"shard":       strconv.Itoa(ctx.Shard),
		"command":     ctx.Command,
		"premium":     strconv.FormatBool(ctx.Premium),
		"permissions": string(perms),
	})
}
