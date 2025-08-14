package api

type Serializer interface {
	Marshal(Envelope) ([]byte, map[string]string, error)
	Unmarshal(body []byte, headers map[string]string) (Envelope, error)
}

type SerializerLocator interface {
	Register(string, Serializer)
	GetAll() []Serializer
	Get(name string) (Serializer, error)
}
