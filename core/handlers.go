package core

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type HandlerFunc struct {
	Fn         reflect.Value
	InputType  reflect.Type
	HandlerStr string
	BusName    string
}

type HandlersRegistry struct {
	handlers map[reflect.Type][]HandlerFunc
}

func NewHandlerRegistry() *HandlersRegistry {
	return &HandlersRegistry{
		handlers: make(map[reflect.Type][]HandlerFunc),
	}
}

func (r *HandlersRegistry) Register(handler any) error {
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

	elem := t.Elem()
	if elem.NumField() > 0 {
		field := elem.Field(0)
		tag := field.Tag.Get("messenger")
		opts := r.parserTagOptions(tag)
		busName = opts["bus"]
	}

	r.handlers[msgType] = append(r.handlers[msgType], HandlerFunc{
		Fn:         v.MethodByName("Handle"),
		InputType:  msgType,
		HandlerStr: runtimeFuncName(handler),
		BusName:    busName,
	})

	return nil
}

func (r *HandlersRegistry) GetAllHandlers() []HandlerFunc {
	var all []HandlerFunc
	for _, handlers := range r.handlers {
		all = append(all, handlers...)
	}
	return all
}

func (r *HandlersRegistry) GetHandlers(msg any) []HandlerFunc {
	t := reflect.TypeOf(msg)
	return r.handlers[t]
}

func (r *HandlersRegistry) ResolveMessageType(typeStr string) (reflect.Type, error) {
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

func (r *HandlersRegistry) registerViaInterface(handler any) error {
	hValue := reflect.ValueOf(handler)
	hType := reflect.TypeOf(handler)

	method, ok := hType.MethodByName("Handle")
	if !ok {
		return fmt.Errorf("method Handle not found")
	}

	if method.Type.NumIn() < 3 {
		return fmt.Errorf("handle method must accept (context.Context, Message)")
	}

	inType := method.Type.In(2)

	r.handlers[inType] = append(r.handlers[inType], HandlerFunc{
		Fn:         hValue.MethodByName("Handle"),
		InputType:  inType,
		HandlerStr: runtimeFuncName(handler),
	})

	return nil
}

func (r *HandlersRegistry) parserTagOptions(tag string) map[string]string {
	opts := make(map[string]string)

	parts := strings.Fields(tag)

	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		opts[key] = val
	}

	return opts
}
