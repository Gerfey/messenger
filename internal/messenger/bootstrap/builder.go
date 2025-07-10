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
	"github.com/gerfey/messenger/internal/handler"
	"github.com/gerfey/messenger/middlewares"
	"github.com/gerfey/messenger/routing"
	"github.com/gerfey/messenger/transport"
	"github.com/gerfey/messenger/transport/amqp"
	"github.com/gerfey/messenger/transport/factory"
	"github.com/gerfey/messenger/transport/inmemory"
)

type Builder struct {
	cfg               *config.MessengerConfig
	resolver          *core.StaticTypeResolver
	transportFactory  *factory.TransportFactoryChain
	handlersLocator   *handler.HandlersLocator
	transportLocator  *transport.TransportLocator
	middlewareLocator *middlewares.MiddlewareLocator
	busLocator        *bus.BusLocator
}

func NewBuilder(cfg *config.MessengerConfig) *Builder {
	resolver := core.NewStaticTypeResolver()

	tf := factory.NewChain(
		amqp.NewAMQPTransportFactory(resolver),
		inmemory.NewInMemoryTransportFactory(resolver),
	)

	return &Builder{
		cfg:               cfg,
		resolver:          resolver,
		transportFactory:  tf,
		handlersLocator:   handler.NewHandlerLocator(),
		transportLocator:  transport.NewTransportLocator(),
		middlewareLocator: middlewares.NewMiddlewareLocator(),
		busLocator:        bus.NewBusLocator(),
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

func (b *Builder) RegisterMiddleware(name string, mw middlewares.Middleware) {
	b.middlewareLocator.Register(name, mw)
}

func (b *Builder) RegisterTransportFactory(f factory.TransportFactory) {
	b.transportFactory = factory.NewChain(
		append(b.transportFactory.Factories(), f)...,
	)
}

func (b *Builder) Build() (*messenger.Messenger, error) {
	router := routing.NewRouter()
	for msgTypeStr, transportName := range b.cfg.Routing {
		t, err := b.handlersLocator.ResolveMessageType(msgTypeStr)
		if err != nil {
			return nil, fmt.Errorf("unknown message type in routing: %s", msgTypeStr)
		}
		router.RouteTypeTo(t, transportName)
	}

	if _, errSendMiddleware := b.middlewareLocator.Get("send_message"); errSendMiddleware != nil {
		b.middlewareLocator.Register(
			"send_message",
			middlewares.NewSendMessageMiddleware(router, b.transportLocator),
		)
	}

	if _, errHandleMiddleware := b.middlewareLocator.Get("handle_message"); errHandleMiddleware != nil {
		b.middlewareLocator.Register(
			"handle_message",
			middlewares.NewHandleMessageMiddleware(b.handlersLocator),
		)
	}

	for name, cfg := range b.cfg.Buses {
		var chain []middlewares.Middleware

		for _, mwName := range cfg.Middleware {
			mw, err := b.middlewareLocator.Get(mwName)
			if err != nil {
				return nil, fmt.Errorf("middleware %q not found", mwName)
			}
			chain = append(chain, mw)
		}

		sendMessageMiddleware, err := b.middlewareLocator.Get("send_message")
		if err != nil {
			return nil, fmt.Errorf("no middleware found for send_message")
		}

		handlerMessageMiddleware, err := b.middlewareLocator.Get("handle_message")
		if err != nil {
			return nil, fmt.Errorf("no middleware found for handle_message")
		}

		chain = append(chain, sendMessageMiddleware)
		chain = append(chain, handlerMessageMiddleware)

		createNewBus := bus.NewBus(name, chain...)

		errBusRegister := b.busLocator.Register(name, createNewBus)
		if errBusRegister != nil {
			return nil, fmt.Errorf("failed to register bus: %w", errBusRegister)
		}
	}

	defaultBus, ok := b.busLocator.Get(b.cfg.DefaultBus)
	if !ok {
		return nil, fmt.Errorf("default_bus %q not found", defaultBus)
	}

	busMap := make(map[reflect.Type]string)
	for _, h := range b.handlersLocator.GetAll() {
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

		activeBus, ok := b.busLocator.Get(busName)
		if !ok {
			return fmt.Errorf("bus %q not found", busName)
		}

		_, err := activeBus.DispatchWithEnvelope(ctx, env)
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

	return messenger.NewMessenger(defaultBus, manager, b.busLocator), nil
}
