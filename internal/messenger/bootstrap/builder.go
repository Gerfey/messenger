package bootstrap

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger"
	"github.com/gerfey/messenger/bus"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core"
	"github.com/gerfey/messenger/envelope"
	"github.com/gerfey/messenger/middlewares"
	"github.com/gerfey/messenger/routing"
	"github.com/gerfey/messenger/transport"
	"github.com/gerfey/messenger/transport/amqp"
	"github.com/gerfey/messenger/transport/factory"
	"github.com/gerfey/messenger/transport/inmemory"
)

type Builder struct {
	cfg                *config.MessengerConfig
	resolver           *core.StaticTypeResolver
	handlerRegistry    *core.HandlersRegistry
	middlewareRegistry map[string]core.Middleware
	transportFactory   *factory.TransportFactoryChain
	messageBusMap      map[reflect.Type]string
}

func NewBuilder(cfg *config.MessengerConfig) *Builder {
	resolver := core.NewStaticTypeResolver()

	tf := factory.NewChain(
		amqp.NewAMQPTransportFactory(resolver),
		inmemory.NewInMemoryTransportFactory(resolver),
	)

	return &Builder{
		cfg:                cfg,
		resolver:           resolver,
		handlerRegistry:    core.NewHandlerRegistry(),
		middlewareRegistry: make(map[string]core.Middleware),
		transportFactory:   tf,
		messageBusMap:      make(map[reflect.Type]string),
	}
}

func (b *Builder) RegisterMessage(msg any) {
	b.resolver.RegisterMessage(msg)
}

func (b *Builder) RegisterHandler(handler any) error {
	if err := b.handlerRegistry.Register(handler); err != nil {
		return err
	}

	for _, h := range b.handlerRegistry.GetAllHandlers() {
		b.resolver.Register(h.InputType.String(), h.InputType)
	}

	return nil
}

func (b *Builder) RegisterMiddleware(name string, mw core.Middleware) {
	b.middlewareRegistry[name] = mw
}

func (b *Builder) RegisterTransportFactory(f factory.TransportFactory) {
	b.transportFactory = factory.NewChain(
		append(b.transportFactory.Factories(), f)...,
	)
}

func (b *Builder) Build() (*messenger.Messenger, error) {
	transports := make(map[string]transport.Transport)
	for name, tCfg := range b.cfg.Transports {
		tr, err := b.transportFactory.CreateTransport(name, tCfg)
		if err != nil {
			return nil, fmt.Errorf("create transport %q: %w", name, err)
		}
		transports[name] = tr
	}

	router := routing.NewRouter()
	for msgTypeStr, transportName := range b.cfg.Routing {
		t, err := b.handlerRegistry.ResolveMessageType(msgTypeStr)
		if err != nil {
			return nil, fmt.Errorf("unknown message type in routing: %s", msgTypeStr)
		}
		router.RouteTypeTo(t, transportName)
	}

	if _, ok := b.middlewareRegistry["send_message"]; !ok {
		b.middlewareRegistry["send_message"] = middlewares.NewSendMessageMiddleware(router, transports)
	}
	if _, ok := b.middlewareRegistry["handle_message"]; !ok {
		b.middlewareRegistry["handle_message"] = middlewares.NewHandleMessageMiddleware(b.handlerRegistry)
	}

	buses := make(map[string]*bus.Bus)
	for name, cfg := range b.cfg.Buses {
		var chain []core.Middleware

		chain = append(chain, b.middlewareRegistry["send_message"])
		chain = append(chain, b.middlewareRegistry["handle_message"])

		for _, mwName := range cfg.Middleware {
			mw, ok := b.middlewareRegistry[mwName]
			if !ok {
				return nil, fmt.Errorf("middleware %q not found", mwName)
			}
			chain = append(chain, mw)
		}

		buses[name] = bus.NewBus(name, chain...)
	}

	defaultBus, ok := buses[b.cfg.DefaultBus]
	if !ok {
		return nil, fmt.Errorf("default_bus %q not found", b.cfg.DefaultBus)
	}

	busMap := make(map[reflect.Type]string)
	for _, h := range b.handlerRegistry.GetAllHandlers() {
		busName := h.BusName
		if busName == "" {
			busName = b.cfg.DefaultBus
		}
		busMap[h.InputType] = busName
	}

	manager := transport.NewManager(func(ctx context.Context, env *envelope.Envelope) error {
		msgType := reflect.TypeOf(env.Message())
		busName, ok := busMap[msgType]
		if !ok {
			busName = b.cfg.DefaultBus
		}

		selectBus, ok := buses[busName]
		if !ok {
			return fmt.Errorf("bus %q not found", busName)
		}

		_, err := selectBus.DispatchWithEnvelope(ctx, env)
		return err
	})

	for _, t := range transports {
		manager.AddTransport(t)
	}

	return messenger.NewMessenger(defaultBus, manager, transports, buses), nil
}
