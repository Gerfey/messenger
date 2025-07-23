package builder

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/gerfey/messenger"
	"github.com/gerfey/messenger/api"
	"github.com/gerfey/messenger/config"
	"github.com/gerfey/messenger/core/bus"
	"github.com/gerfey/messenger/core/event"
	"github.com/gerfey/messenger/core/handler"
	"github.com/gerfey/messenger/core/listener"
	"github.com/gerfey/messenger/core/middleware"
	"github.com/gerfey/messenger/core/middleware/implementation"
	"github.com/gerfey/messenger/core/retry"
	"github.com/gerfey/messenger/core/routing"
	"github.com/gerfey/messenger/core/stamps"
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
	eventDispatcher   api.EventDispatcher
	logger            *slog.Logger
}

func NewBuilder(cfg *config.MessengerConfig, logger *slog.Logger) api.Builder {
	resolver := NewResolver()

	tf := transport.NewFactoryChain(
		amqp.NewTransportFactory(resolver, logger),
		inmemory.NewTransportFactory(resolver, logger),
	)

	return &Builder{
		cfg:               cfg,
		resolver:          resolver,
		transportFactory:  tf,
		handlersLocator:   handler.NewHandlerLocator(),
		transportLocator:  transport.NewLocator(),
		middlewareLocator: middleware.NewMiddlewareLocator(),
		busLocator:        bus.NewLocator(),
		eventDispatcher:   event.NewEventDispatcher(logger),
		logger:            logger,
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

func (b *Builder) RegisterStamp(stamp any) {
	b.resolver.RegisterStamp(stamp)
}

func (b *Builder) RegisterListener(event any, listener any) {
	b.eventDispatcher.AddListener(event, listener)
}

func (b *Builder) Build() (api.Messenger, error) {
	router := routing.NewRouter()
	for msgTypeStr, transportName := range b.cfg.Routing {
		t, err := b.handlersLocator.ResolveMessageType(msgTypeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve message type '%s' in routing configuration: %w", msgTypeStr, err)
		}
		router.RouteTypeTo(t, transportName)
	}

	b.registerStamps()

	if err := b.setupBuses(router); err != nil {
		return nil, err
	}

	return b.createMessenger(router)
}

func (b *Builder) setupBuses(router api.Router) error {
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
		chain = append(
			chain,
			implementation.NewSendMessageMiddleware(router, b.transportLocator, b.eventDispatcher, b.logger),
		)
		chain = append(chain, implementation.NewHandleMessageMiddleware(b.handlersLocator, b.logger))

		createNewBus := bus.NewBus(chain...)

		errBusRegister := b.busLocator.Register(name, createNewBus)
		if errBusRegister != nil {
			return fmt.Errorf("failed to register bus: %w", errBusRegister)
		}
	}

	return nil
}

func (b *Builder) createMessenger(router api.Router) (api.Messenger, error) {
	busMap := make(map[reflect.Type]string)
	for _, h := range b.handlersLocator.GetAll() {
		busName := h.BusName
		if busName == "" {
			busName = b.cfg.DefaultBus
		}
		busMap[h.InputType] = busName
	}

	handlerManager := func(ctx context.Context, env api.Envelope) error {
		msgType := reflect.TypeOf(env.Message())
		busName, ok := busMap[msgType]
		if !ok {
			busName = b.cfg.DefaultBus
		}

		activeBus, ok := b.busLocator.Get(busName)
		if !ok {
			return fmt.Errorf("bus '%s' not found for message type %T", busName, env.Message())
		}

		_, err := activeBus.Dispatch(ctx, env)

		return err
	}

	manager := transport.NewManager(handlerManager, b.eventDispatcher, b.logger)

	for name, tCfg := range b.cfg.Transports {
		tr, err := b.transportFactory.CreateTransport(name, tCfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create transport '%s': %w", name, err)
		}

		manager.AddTransport(tr)

		errTransportLocator := b.transportLocator.Register(name, tr)
		if errTransportLocator != nil {
			return nil, fmt.Errorf("failed to register transport '%s': %w", name, errTransportLocator)
		}
	}

	for name, tCfg := range b.cfg.Transports {
		tr := b.transportLocator.GetTransport(name)

		if retryable, ok := tr.(api.RetryableTransport); ok && tCfg.RetryStrategy != nil {
			strategy := retry.NewMultiplierRetryStrategy(
				tCfg.RetryStrategy.MaxRetries,
				tCfg.RetryStrategy.Delay,
				tCfg.RetryStrategy.Multiplier,
				tCfg.RetryStrategy.MaxDelay,
			)

			var failureTransport api.Transport
			if b.cfg.FailureTransport != "" {
				failureTransport = b.transportLocator.GetTransport(b.cfg.FailureTransport)
			}

			lst := listener.NewSendFailedMessageForRetryListener(retryable, failureTransport, strategy, b.logger)
			b.eventDispatcher.AddListener(event.SendFailedMessageEvent{}, lst)
		}
	}

	defaultBus, ok := b.busLocator.Get(b.cfg.DefaultBus)
	if !ok {
		return nil, fmt.Errorf("default_bus %q not found", defaultBus)
	}

	return messenger.NewMessenger(b.cfg.DefaultBus, manager, b.busLocator, router), nil
}

func (b *Builder) registerStamps() {
	b.resolver.RegisterStamp(stamps.BusNameStamp{})
	b.resolver.RegisterStamp(stamps.RedeliveryStamp{})
}
