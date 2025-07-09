package messages

type HelloMessage struct {
	Text string
}

func (m *HelloMessage) RoutingKey() string {
	return "test_routing_key"
}
