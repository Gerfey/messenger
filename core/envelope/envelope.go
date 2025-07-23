package envelope

import (
	"github.com/gerfey/messenger/api"
)

type Envelope struct {
	message any
	stamps  []api.Stamp
}

func NewEnvelope(message any) api.Envelope {
	return &Envelope{
		message: message,
		stamps:  []api.Stamp{},
	}
}

func (e *Envelope) Message() any {
	return e.message
}

func (e *Envelope) WithStamp(s api.Stamp) api.Envelope {
	newStamps := append([]api.Stamp{}, e.stamps...)
	newStamps = append(newStamps, s)

	return &Envelope{
		message: e.message,
		stamps:  newStamps,
	}
}

func (e *Envelope) Stamps() []api.Stamp {
	return e.stamps
}
