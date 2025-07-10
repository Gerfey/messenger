package api

import "reflect"

type Stamp interface{}

type Envelope interface {
	Message() any
	WithStamp(Stamp) Envelope
	Stamps() []Stamp
	StampsOfType(t reflect.Type) []Stamp
	LastStampOfType(t reflect.Type) any
}
