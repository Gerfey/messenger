package event

import "github.com/gerfey/messenger/api"

type SendFailedMessageEvent struct {
	TransportName string
	Envelope      api.Envelope
	Error         error
}
