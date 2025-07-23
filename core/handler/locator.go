package handler

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/gerfey/messenger/api"
)

const (
	expectedHandlerParams = 3
	messageTypeParamIndex = 2
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
		return fmt.Errorf("handler %T does not have a Handle method", handler)
	}

	if method.Type.NumIn() != expectedHandlerParams {
		return fmt.Errorf(
			"handler %T: Handle method must accept exactly 2 parameters (context.Context, Message), got %d",
			handler,
			method.Type.NumIn()-1,
		)
	}

	ctxType := method.Type.In(1)
	msgType := method.Type.In(messageTypeParamIndex)

	if !ctxType.Implements(reflect.TypeOf((*context.Context)(nil)).Elem()) {
		return fmt.Errorf("handler %T: first parameter must be context.Context, got %v", handler, ctxType)
	}

	errorInterface := reflect.TypeOf((*error)(nil)).Elem()
	if method.Type.NumOut() != 1 && method.Type.NumOut() != 2 {
		return fmt.Errorf(
			"handler %T: Handle method must return error or (result, error), got %d return values",
			handler,
			method.Type.NumOut(),
		)
	}
	if !method.Type.Out(method.Type.NumOut() - 1).Implements(errorInterface) {
		return fmt.Errorf(
			"handler %T: last return value must be error, got %v",
			handler,
			method.Type.Out(method.Type.NumOut()-1),
		)
	}

	busName := ""

	messageHandlerType := reflect.TypeOf((*api.MessageHandlerType)(nil)).Elem()
	if t.Implements(messageHandlerType) {
		if messageHandler, handlerOk := handler.(api.MessageHandlerType); handlerOk {
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
	all := make([]api.HandlerFunc, 0)
	for _, handlers := range r.handlers {
		all = append(all, handlers...)
	}

	return all
}

func (r *Locator) Get(msg any) []api.HandlerFunc {
	t := reflect.TypeOf(msg)

	handlers, ok := r.handlers[t]
	if !ok {
		return []api.HandlerFunc{}
	}

	return handlers
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
	v := reflect.ValueOf(i)
	var fn reflect.Value

	if v.Kind() == reflect.Func {
		fn = v
	} else {
		method := v.MethodByName("Handle")
		if !method.IsValid() || method.IsZero() || method.Kind() != reflect.Func {
			return fmt.Sprintf("%T.Handle (invalid)", i)
		}
		fn = method
	}

	ptr := fn.Pointer()
	if ptr == 0 {
		return fmt.Sprintf("%T.Handle (no pointer)", i)
	}

	rf := runtime.FuncForPC(ptr)
	if rf != nil {
		return rf.Name()
	}

	return fmt.Sprintf("%T.Handle (no symbol)", i)
}
