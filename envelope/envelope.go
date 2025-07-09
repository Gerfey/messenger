package envelope

import (
	"reflect"

	"github.com/gerfey/messenger/stamps"
)

type Stamp interface{}

type Envelope struct {
	message any
	stamps  []Stamp
}

func NewEnvelope(message any) *Envelope {
	return &Envelope{
		message: message,
		stamps:  []Stamp{},
	}
}

func (e *Envelope) Message() any {
	return e.message
}

func (e *Envelope) WithStamp(s Stamp) *Envelope {
	newStamps := append([]Stamp{}, e.stamps...)
	newStamps = append(newStamps, s)

	return &Envelope{
		message: e.message,
		stamps:  newStamps,
	}
}

func (e *Envelope) Stamps() []Stamp {
	return e.stamps
}

func (e *Envelope) StampsOfType(t reflect.Type) []Stamp {
	var filtered []Stamp
	for _, s := range e.stamps {
		if reflect.TypeOf(s).AssignableTo(t) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (e *Envelope) LastStampOfType(t reflect.Type) any {
	for i := len(e.stamps) - 1; i >= 0; i-- {
		if reflect.TypeOf(e.stamps[i]).AssignableTo(t) {
			return e.stamps[i]
		}
	}
	return nil
}

func GetLastResult[T any](env *Envelope) (T, bool) {
	var zero T

	if env == nil {
		return zero, false
	}

	stampAny := env.LastStampOfType(reflect.TypeOf(stamps.HandledStamp{}))
	if stampAny == nil {
		return zero, false
	}

	stamp := stampAny.(stamps.HandledStamp)
	result, ok := stamp.Result.(T)
	return result, ok
}
