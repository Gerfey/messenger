package handler

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	"github.com/gerfey/messenger/api"
)

type Locator struct {
	handlers map[reflect.Type][]api.HandlerFunc
}

func NewHandlerLocator() api.HandlerLocator {
	return &Locator{
		handlers: make(map[reflect.Type][]api.HandlerFunc),
	}
}

func (r *Locator) Register(handler any) error {
	v := reflect.ValueOf(handler)
	t := reflect.TypeOf(handler)

	method, ok := t.MethodByName("Handle")
	if !ok {
		return fmt.Errorf("no Handle method found")
	}

	if method.Type.NumIn() != 3 {
		return fmt.Errorf("handle must accept (context.Context, Message)")
	}

	ctxType := method.Type.In(1)
	msgType := method.Type.In(2)

	if !ctxType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return fmt.Errorf("first argument must be context.Context")
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if method.Type.NumOut() != 1 && method.Type.NumOut() != 2 {
		return fmt.Errorf("handle must return error or (result, error)")
	}
	if !method.Type.Out(method.Type.NumOut() - 1).Implements(errorInterface) {
		return fmt.Errorf("last return value must be error")
	}

	busName := ""

	messageHandlerType := reflect.TypeOf((*api.MessageHandlerType)(nil)).Elem()
	if t.Implements(messageHandlerType) {
		if messageHandler, ok := handler.(api.MessageHandlerType); ok {
			busName = messageHandler.GetBusName()
		}
	}

	r.handlers[msgType] = append(r.handlers[msgType], api.HandlerFunc{
		Fn:         v.MethodByName("Handle"),
		InputType:  msgType,
		HandlerStr: runtimeFuncName(handler),
		BusName:    busName,
	})

	return nil
}

func (r *Locator) GetAll() []api.HandlerFunc {
	var all []api.HandlerFunc
	for _, handlers := range r.handlers {
		all = append(all, handlers...)
	}

	return all
}

func (r *Locator) Get(msg any) []api.HandlerFunc {
	t := reflect.TypeOf(msg)
	return r.handlers[t]
}

func (r *Locator) ResolveMessageType(typeStr string) (reflect.Type, error) {
	for t := range r.handlers {
		if t.String() == typeStr {
			return t, nil
		}

		if t.Kind() == reflect.Ptr && t.Elem().String() == typeStr {
			return t, nil
		}

		ptr := reflect.PointerTo(t)
		if ptr.String() == typeStr {
			return ptr, nil
		}
	}

	return nil, fmt.Errorf("message type %q not found in registry", typeStr)
}

func runtimeFuncName(i any) string {
	return strconv.Itoa(int(reflect.ValueOf(i).Pointer()))
}
