package message

import "reflect"

type RoutedMessage interface {
	RoutingKey() string
}

func GetRoutingKey(msg any) string {
	if rm, ok := msg.(RoutedMessage); ok {
		return rm.RoutingKey()
	}
	return reflect.TypeOf(msg).String()
}
