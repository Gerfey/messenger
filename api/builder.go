package api

type Builder interface {
	RegisterMessage(any)
	RegisterHandler(any) error
	RegisterStamp(any)
	RegisterListener(any, any)
	RegisterMiddleware(string, Middleware)
	RegisterSerializer(string, Serializer)
	RegisterTransportFactory(TransportFactory)
	Build() (Messenger, error)
}
