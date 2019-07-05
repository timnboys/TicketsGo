package sentry

import (
	"github.com/getsentry/raven-go"
	"github.com/go-errors/errors"
)

func Error(e error) {
	wrapped := e.(*errors.Error)
	raven.Capture(ConstructPacket(wrapped), nil)
}
