package messages

type ExampleHelloMessage struct {
	Text string
}

func (m *ExampleHelloMessage) RoutingKey() string {
	return "test_routing_key"
}
