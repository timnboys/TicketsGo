package sentry

import (
	"github.com/bwmarrin/discordgo"
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
	"strconv"
)

type ErrorContext struct {
	Guild       uint64
	User        uint64
	Channel     uint64
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
	wrapped := errors.New(e)
	raven.Capture(ConstructPacket(wrapped, raven.INFO), map[string]string{
		"guild":       ctx.Guild,
		"user":        ctx.User,
		"channel":     ctx.Channel,
		"shard":       strconv.Itoa(ctx.Shard),
		"command":     ctx.Command,
		"premium":     strconv.FormatBool(ctx.Premium),
	})
}

func ErrorWithContext(e error, ctx ErrorContext) {
	wrapped := errors.New(e)
	raven.Capture(ConstructErrorPacket(wrapped), map[string]string{
		"guild":       ctx.Guild,
		"user":        ctx.User,
		"channel":     ctx.Channel,
		"shard":       strconv.Itoa(ctx.Shard),
		"command":     ctx.Command,
		"premium":     strconv.FormatBool(ctx.Premium),
	})
}
