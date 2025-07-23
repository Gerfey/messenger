package api

type Stamp interface{}

type Envelope interface {
	Message() any
	WithStamp(Stamp) Envelope
	Stamps() []Stamp
}
