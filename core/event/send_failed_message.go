package event

import "github.com/gerfey/messenger/api"

type SendFailedMessageEvent struct {
	Envelope      api.Envelope
	Error         error
	TransportName string
}
