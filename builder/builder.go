package builder

import (
	"context"
	"fmt"
	"reflect"

	"github.com/gerfey/messenger"
	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core/bus"
	"github.com/gerfey/messenger/core/handler"
	"github.com/gerfey/messenger/core/middleware"
	"github.com/gerfey/messenger/core/middleware/implementation"
	"github.com/gerfey/messenger/core/routing"
	"github.com/gerfey/messenger/transport"
	"github.com/gerfey/messenger/transport/amqp"
	"github.com/gerfey/messenger/transport/inmemory"
)

type Builder struct {
	cfg               *config.MessengerConfig
	resolver          api.TypeResolver
	transportFactory  *transport.FactoryChain
	handlersLocator   api.HandlerLocator
	transportLocator  api.TransportLocator
	middlewareLocator api.MiddlewareLocator
	busLocator        api.BusLocator
}

func NewBuilder(cfg *config.MessengerConfig) *Builder {
	resolver := NewStaticTypeResolver()

	tf := transport.NewFactoryChain(
		amqp.NewTransportFactory(resolver),
		inmemory.NewTransportFactory(resolver),
	)

	return &Builder{
		cfg:               cfg,
		resolver:          resolver,
		transportFactory:  tf,
		handlersLocator:   handler.NewHandlerLocator(),
		transportLocator:  transport.NewLocator(),
		middlewareLocator: middleware.NewMiddlewareLocator(),
		busLocator:        bus.NewLocator(),
	}
}

func (b *Builder) RegisterMessage(msg any) {
	b.resolver.RegisterMessage(msg)
}

func (b *Builder) RegisterHandler(handler any) error {
	if err := b.handlersLocator.Register(handler); err != nil {
		return fmt.Errorf("register handler: %w", err)
	}

	for _, h := range b.handlersLocator.GetAll() {
		b.resolver.Register(h.InputType.String(), h.InputType)
	}

	return nil
}

func (b *Builder) RegisterMiddleware(name string, mw api.Middleware) {
	b.middlewareLocator.Register(name, mw)
}

func (b *Builder) RegisterTransportFactory(f api.TransportFactory) {
	b.transportFactory = transport.NewFactoryChain(
		append(b.transportFactory.Factories(), f)...,
	)
}

func (b *Builder) Build() (api.Messenger, error) {
	if err := b.setupBuses(); err != nil {
		return nil, err
	}

	return b.createMessenger()
}

func (b *Builder) setupBuses() error {
	router := routing.NewRouter()
	for msgTypeStr, transportName := range b.cfg.Routing {
		t, err := b.handlersLocator.ResolveMessageType(msgTypeStr)
		if err != nil {
			return fmt.Errorf("unknown message type in routing: %s", msgTypeStr)
		}
		router.RouteTypeTo(t, transportName)
	}

	for name, cfg := range b.cfg.Buses {
		var chain []api.Middleware

		for _, mwName := range cfg.Middleware {
			mw, err := b.middlewareLocator.Get(mwName)
			if err != nil {
				return fmt.Errorf("middleware %q not found", mwName)
			}
			chain = append(chain, mw)
		}

		chain = append(chain, implementation.NewAddBusNameMiddleware(name))
		chain = append(chain, implementation.NewSendMessageMiddleware(router, b.transportLocator))
		chain = append(chain, implementation.NewHandleMessageMiddleware(b.handlersLocator))

		createNewBus := bus.NewBus(chain...)

		errBusRegister := b.busLocator.Register(name, createNewBus)
		if errBusRegister != nil {
			return fmt.Errorf("failed to register bus: %w", errBusRegister)
		}
	}

	return nil
}

func (b *Builder) createMessenger() (api.Messenger, error) {
	busMap := make(map[reflect.Type]string)
	for _, h := range b.handlersLocator.GetAll() {
		busName := h.BusName
		if busName == "" {
			busName = b.cfg.DefaultBus
		}
		busMap[h.InputType] = busName
	}

	manager := transport.NewManager(func(ctx context.Context, env api.Envelope) error {
		msgType := reflect.TypeOf(env.Message())
		busName, ok := busMap[msgType]
		if !ok {
			busName = b.cfg.DefaultBus
		}

		activeBus, ok := b.busLocator.Get(busName)
		if !ok {
			return fmt.Errorf("bus %q not found", busName)
		}

		_, err := activeBus.Dispatch(ctx, env)
		return err
	})

	for name, tCfg := range b.cfg.Transports {
		tr, err := b.transportFactory.CreateTransport(name, tCfg)
		if err != nil {
			return nil, fmt.Errorf("create transport %q: %w", name, err)
		}

		manager.AddTransport(tr)

		errTransportLocator := b.transportLocator.Register(name, tr)
		if errTransportLocator != nil {
			return nil, fmt.Errorf("register transport %q: %w", name, errTransportLocator)
		}
	}

	defaultBus, ok := b.busLocator.Get(b.cfg.DefaultBus)
	if !ok {
		return nil, fmt.Errorf("default_bus %q not found", defaultBus)
	}

	return messenger.NewMessenger(defaultBus, manager, b.busLocator), nil
}
