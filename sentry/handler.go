package sentry

import (
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
)

func Error(e error) {
	wrapped := errors.New(e)
	raven.Capture(ConstructPacket(wrapped), nil)
}
