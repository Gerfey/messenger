package messages

type UserCreatedMessage struct {
	ID   int
	Name string
}

func (u *UserCreatedMessage) RoutingKey() string {
	return "user_created"
}
