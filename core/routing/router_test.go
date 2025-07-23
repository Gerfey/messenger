package routing_test

import (
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/routing"
	"github.com/gerfey/messenger/tests/helpers"
)

type TestMessage struct {
	ID      string
	Content string
}

type AnotherMessage struct {
	Data string
}

func TestNewRouter(t *testing.T) {
	t.Run("creates new router", func(t *testing.T) {
		router := routing.NewRouter()

		require.NotNil(t, router)
		assert.Empty(t, router.GetUsedTransports())
	})
}

func TestRouter_RouteMessageTo(t *testing.T) {
	t.Run("routes message to single transport", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg, "amqp")

		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp"}, transports)
	})

	t.Run("routes message to multiple transports", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg, "amqp", "inmemory", "redis")

		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp", "inmemory", "redis"}, transports)
	})

	t.Run("overwrites existing route", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg, "amqp")
		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp"}, transports)

		router.RouteMessageTo(msg, "inmemory", "redis")
		transports = router.GetTransportFor(msg)
		assert.Equal(t, []string{"inmemory", "redis"}, transports)
	})

	t.Run("routes different message types separately", func(t *testing.T) {
		router := routing.NewRouter()
		msg1 := &TestMessage{ID: "1", Content: "test"}
		msg2 := &AnotherMessage{Data: "data"}

		router.RouteMessageTo(msg1, "amqp")
		router.RouteMessageTo(msg2, "inmemory")

		transports1 := router.GetTransportFor(msg1)
		transports2 := router.GetTransportFor(msg2)

		assert.Equal(t, []string{"amqp"}, transports1)
		assert.Equal(t, []string{"inmemory"}, transports2)
	})

	t.Run("handles nil message", func(t *testing.T) {
		router := routing.NewRouter()

		router.RouteMessageTo(nil, "amqp")

		transports := router.GetTransportFor(nil)
		assert.Equal(t, []string{"amqp"}, transports)
	})

	t.Run("handles empty transport list", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg)

		transports := router.GetTransportFor(msg)
		assert.Empty(t, transports)
	})
}

func TestRouter_RouteTypeTo(t *testing.T) {
	t.Run("routes type to single transport", func(t *testing.T) {
		router := routing.NewRouter()
		msgType := reflect.TypeOf(&TestMessage{})

		router.RouteTypeTo(msgType, "amqp")

		msg := &TestMessage{ID: "1", Content: "test"}
		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp"}, transports)
	})

	t.Run("routes type to multiple transports", func(t *testing.T) {
		router := routing.NewRouter()
		msgType := reflect.TypeOf(&TestMessage{})

		router.RouteTypeTo(msgType, "amqp", "inmemory")

		msg := &TestMessage{ID: "1", Content: "test"}
		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp", "inmemory"}, transports)
	})

	t.Run("overwrites existing type route", func(t *testing.T) {
		router := routing.NewRouter()
		msgType := reflect.TypeOf(&TestMessage{})

		router.RouteTypeTo(msgType, "amqp")
		msg := &TestMessage{ID: "1", Content: "test"}
		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp"}, transports)

		router.RouteTypeTo(msgType, "inmemory")
		transports = router.GetTransportFor(msg)
		assert.Equal(t, []string{"inmemory"}, transports)
	})

	t.Run("handles nil type", func(t *testing.T) {
		router := routing.NewRouter()

		router.RouteTypeTo(nil, "amqp")

		transports := router.GetTransportFor(nil)
		assert.Equal(t, []string{"amqp"}, transports)
	})
}

func TestRouter_GetTransportFor(t *testing.T) {
	t.Run("returns empty slice for unrouted message", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		transports := router.GetTransportFor(msg)
		assert.Empty(t, transports)
	})

	t.Run("returns correct transports for routed message", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg, "amqp", "inmemory")

		transports := router.GetTransportFor(msg)
		assert.Equal(t, []string{"amqp", "inmemory"}, transports)
	})

	t.Run("returns transports for different instances of same type", func(t *testing.T) {
		router := routing.NewRouter()
		msg1 := &TestMessage{ID: "1", Content: "test1"}
		msg2 := &TestMessage{ID: "2", Content: "test2"}

		router.RouteMessageTo(msg1, "amqp")

		transports := router.GetTransportFor(msg2)
		assert.Equal(t, []string{"amqp"}, transports)
	})

	t.Run("handles nil message", func(t *testing.T) {
		router := routing.NewRouter()

		transports := router.GetTransportFor(nil)
		assert.Empty(t, transports)
	})
}

func TestRouter_GetUsedTransports(t *testing.T) {
	t.Run("returns empty slice when no routes", func(t *testing.T) {
		router := routing.NewRouter()

		transports := router.GetUsedTransports()
		assert.Empty(t, transports)
	})

	t.Run("returns unique transports from single route", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg, "amqp", "inmemory")

		transports := router.GetUsedTransports()
		sort.Strings(transports)
		assert.Equal(t, []string{"amqp", "inmemory"}, transports)
	})

	t.Run("returns unique transports from multiple routes", func(t *testing.T) {
		router := routing.NewRouter()
		msg1 := &TestMessage{ID: "1", Content: "test"}
		msg2 := &AnotherMessage{Data: "data"}

		router.RouteMessageTo(msg1, "amqp", "inmemory")
		router.RouteMessageTo(msg2, "inmemory", "redis")

		transports := router.GetUsedTransports()
		sort.Strings(transports)
		assert.Equal(t, []string{"amqp", "inmemory", "redis"}, transports)
	})

	t.Run("deduplicates transport names", func(t *testing.T) {
		router := routing.NewRouter()
		msg1 := &TestMessage{ID: "1", Content: "test"}
		msg2 := &AnotherMessage{Data: "data"}

		router.RouteMessageTo(msg1, "amqp", "inmemory")
		router.RouteMessageTo(msg2, "amqp", "inmemory")

		transports := router.GetUsedTransports()
		sort.Strings(transports)
		assert.Equal(t, []string{"amqp", "inmemory"}, transports)
	})

	t.Run("handles empty transport lists", func(t *testing.T) {
		router := routing.NewRouter()
		msg := &TestMessage{ID: "1", Content: "test"}

		router.RouteMessageTo(msg)

		transports := router.GetUsedTransports()
		assert.Empty(t, transports)
	})
}

func TestRouter_Integration(t *testing.T) {
	t.Run("complex routing scenario", func(t *testing.T) {
		router := routing.NewRouter()

		testMsg := &helpers.TestMessage{ID: "1", Content: "test"}
		anotherMsg := &AnotherMessage{Data: "data"}

		router.RouteMessageTo(testMsg, "amqp", "redis")
		router.RouteMessageTo(anotherMsg, "inmemory")

		msgType := reflect.TypeOf(&TestMessage{})
		router.RouteTypeTo(msgType, "kafka")

		transports1 := router.GetTransportFor(testMsg)
		transports2 := router.GetTransportFor(anotherMsg)
		transports3 := router.GetTransportFor(&TestMessage{ID: "2", Content: "other"})

		assert.Equal(t, []string{"amqp", "redis"}, transports1)
		assert.Equal(t, []string{"inmemory"}, transports2)
		assert.Equal(t, []string{"kafka"}, transports3)

		allTransports := router.GetUsedTransports()
		sort.Strings(allTransports)
		assert.Equal(t, []string{"amqp", "inmemory", "kafka", "redis"}, allTransports)
	})
}
