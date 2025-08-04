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
	"github.com/gerfey/messenger/transport/kafka"
	"github.com/gerfey/messenger/transport/redis"
	"github.com/gerfey/messenger/transport/sync"
)

type Builder struct {
	cfg               *config.MessengerConfig
	resolver          api.TypeResolver
	transportFactory  *transport.FactoryChain
	handlersLocator   api.HandlerLocator
	senderLocator     api.SenderLocator
	middlewareLocator api.MiddlewareLocator
	busLocator        api.BusLocator
	eventDispatcher   api.EventDispatcher
	logger            *slog.Logger
}

func NewBuilder(cfg *config.MessengerConfig, logger *slog.Logger) api.Builder {
	resolver := NewResolver()

	busLocator := bus.NewLocator()

	tf := transport.NewFactoryChain(
		amqp.NewTransportFactory(logger, resolver),
		inmemory.NewTransportFactory(logger, resolver),
		sync.NewTransportFactory(logger, busLocator),
		kafka.NewTransportFactory(logger, resolver),
		redis.NewTransportFactory(logger, resolver),
	)

	return &Builder{
		cfg:               cfg,
		resolver:          resolver,
		transportFactory:  tf,
		handlersLocator:   handler.NewHandlerLocator(),
		senderLocator:     transport.NewSenderLocator(),
		middlewareLocator: middleware.NewMiddlewareLocator(),
		busLocator:        busLocator,
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
	b.registerStamps()

	if err := b.setupBuses(); err != nil {
		return nil, err
	}

	return b.createMessenger()
}

func (b *Builder) setupBuses() error {
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
		chain = append(chain, implementation.NewAddMessageIDMiddleware())

		chain = append(
			chain,
			implementation.NewSendMessageMiddleware(b.logger, b.senderLocator, b.eventDispatcher),
		)
		chain = append(chain, implementation.NewHandleMessageMiddleware(b.logger, b.handlersLocator))

		createNewBus := bus.NewBus(chain...)

		errBusRegister := b.busLocator.Register(name, createNewBus)
		if errBusRegister != nil {
			return fmt.Errorf("failed to register bus: %w", errBusRegister)
		}
	}

	return nil
}

func (b *Builder) createMessenger() (api.Messenger, error) {
	router, err := b.setupRouting()
	if err != nil {
		return nil, err
	}

	busMap := b.createBusMap()
	handlerManager := b.createHandlerManager(busMap)
	manager := transport.NewManager(b.logger, handlerManager, b.eventDispatcher)

	createdTransports, transportNames, err := b.createTransports(manager)
	if err != nil {
		return nil, err
	}

	b.setupFallbackTransports(transportNames)
	b.setupRetryListeners(createdTransports)

	defaultBus, ok := b.busLocator.Get(b.cfg.DefaultBus)
	if !ok {
		return nil, fmt.Errorf("default_bus %q not found", defaultBus)
	}

	return messenger.NewMessenger(b.cfg.DefaultBus, manager, b.busLocator, router), nil
}

func (b *Builder) setupRouting() (api.Router, error) {
	router := routing.NewRouter()

	for msgTypeStr, transportName := range b.cfg.Routing {
		t, err := b.handlersLocator.ResolveMessageType(msgTypeStr)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve message type '%s' in routing configuration: %w", msgTypeStr, err)
		}
		router.RouteTypeTo(t, transportName)
		b.senderLocator.RegisterMessageType(t, []string{transportName})
	}

	return router, nil
}

func (b *Builder) createBusMap() map[reflect.Type]string {
	busMap := make(map[reflect.Type]string)
	for _, h := range b.handlersLocator.GetAll() {
		busName := h.BusName
		if busName == "" {
			busName = b.cfg.DefaultBus
		}
		busMap[h.InputType] = busName
	}

	return busMap
}

func (b *Builder) createHandlerManager(busMap map[reflect.Type]string) func(context.Context, api.Envelope) error {
	return func(ctx context.Context, env api.Envelope) error {
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
}

func (b *Builder) createTransports(manager *transport.Manager) (map[string]api.Transport, []string, error) {
	var transportNames []string
	createdTransports := make(map[string]api.Transport)

	b.createdSyncTransport(createdTransports)

	for name, tCfg := range b.cfg.Transports {
		tr, err := b.transportFactory.CreateTransport(name, tCfg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create transport '%s': %w", name, err)
		}

		createdTransports[name] = tr
	}

	for nameTransport, tr := range createdTransports {
		manager.AddTransport(tr)
		transportNames = append(transportNames, nameTransport)

		errTransportLocator := b.senderLocator.Register(nameTransport, tr)
		if errTransportLocator != nil {
			return nil, nil, fmt.Errorf("failed to register transport '%s': %w", nameTransport, errTransportLocator)
		}
	}

	return createdTransports, transportNames, nil
}

func (b *Builder) setupFallbackTransports(transportNames []string) {
	if len(b.cfg.Routing) == 0 && len(transportNames) > 0 {
		b.senderLocator.SetFallback(transportNames)
	}
}

func (b *Builder) setupRetryListeners(createdTransports map[string]api.Transport) {
	for name, tCfg := range b.cfg.Transports {
		t := createdTransports[name]

		if retryable, ok := t.(api.RetryableTransport); ok && tCfg.RetryStrategy != nil {
			strategy := retry.NewMultiplierRetryStrategy(
				tCfg.RetryStrategy.MaxRetries,
				tCfg.RetryStrategy.Delay,
				tCfg.RetryStrategy.Multiplier,
				tCfg.RetryStrategy.MaxDelay,
			)

			var failureTransport api.Transport
			if b.cfg.FailureTransport != "" {
				if ft, exists := createdTransports[b.cfg.FailureTransport]; exists {
					failureTransport = ft
				}
			}

			lst := listener.NewSendFailedMessageForRetryListener(b.logger, retryable, failureTransport, strategy)
			b.eventDispatcher.AddListener(event.SendFailedMessageEvent{}, lst)
		}
	}
}

func (b *Builder) registerStamps() {
	b.resolver.RegisterStamp(stamps.BusNameStamp{})
	b.resolver.RegisterStamp(stamps.RedeliveryStamp{})
	b.resolver.RegisterStamp(stamps.MessageIDStamp{})
}

func (b *Builder) createdSyncTransport(createdTransports map[string]api.Transport) {
	cfg := config.TransportConfig{
		DSN: "sync://",
	}

	if syncTransport, err := b.transportFactory.CreateTransport("sync", cfg); err == nil {
		createdTransports["sync"] = syncTransport
	}
}
