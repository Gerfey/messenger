package transport

import (
	"github.com/gerfey/messenger/envelope"
)

type Serializer interface {
	Marshal(env *envelope.Envelope) ([]byte, map[string]string, error)
	Unmarshal(body []byte, headers map[string]string) (*envelope.Envelope, error)
}
