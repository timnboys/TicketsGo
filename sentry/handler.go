package sentry

import (
	"github.com/apex/log"
	"github.com/getsentry/raven-go"
)

type Handler struct {
}

var Default = NewHandler()

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HandleLog(e *log.Entry) error {
	//packet := ConstructPacket(e)
	raven.CaptureMessage(e.Message, nil)
	return nil
}
