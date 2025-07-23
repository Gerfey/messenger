package bus_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gerfey/messenger/core/bus"

	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewLocator(t *testing.T) {
	t.Run("create new locator", func(t *testing.T) {
		locator := bus.NewLocator()

		require.NotNil(t, locator)
		assert.IsType(t, &bus.Locator{}, locator)

		all := locator.GetAll()
		assert.Empty(t, all)
	})
}

func TestLocator_Register(t *testing.T) {
	t.Run("register single bus", func(t *testing.T) {
		locator := bus.NewLocator()
		b := bus.NewBus()

		err := locator.Register("test-bus", b)

		require.NoError(t, err)

		retrievedBus, found := locator.Get("test-bus")
		assert.True(t, found)
		assert.Equal(t, b, retrievedBus)
	})

	t.Run("register multiple buses", func(t *testing.T) {
		locator := bus.NewLocator()
		bus1 := bus.NewBus()
		bus2 := bus.NewBus(&helpers.TestMiddleware{})

		err1 := locator.Register("bus1", bus1)
		err2 := locator.Register("bus2", bus2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		retrievedBus1, found1 := locator.Get("bus1")
		assert.True(t, found1)
		assert.Equal(t, bus1, retrievedBus1)

		retrievedBus2, found2 := locator.Get("bus2")
		assert.True(t, found2)
		assert.Equal(t, bus2, retrievedBus2)
	})

	t.Run("register bus with same name overwrites previous", func(t *testing.T) {
		locator := bus.NewLocator()
		bus1 := bus.NewBus()
		bus2 := bus.NewBus(&helpers.TestMiddleware{})

		err1 := locator.Register("test-bus", bus1)
		err2 := locator.Register("test-bus", bus2)

		require.NoError(t, err1)
		require.NoError(t, err2)

		retrievedBus, found := locator.Get("test-bus")
		assert.True(t, found)
		assert.Equal(t, bus2, retrievedBus)
		assert.NotEqual(t, bus1, retrievedBus)
	})

	t.Run("register bus with empty name", func(t *testing.T) {
		locator := bus.NewLocator()
		b := bus.NewBus()

		err := locator.Register("", b)

		require.NoError(t, err)

		retrievedBus, found := locator.Get("")
		assert.True(t, found)
		assert.Equal(t, b, retrievedBus)
	})

	t.Run("register nil bus", func(t *testing.T) {
		locator := bus.NewLocator()

		err := locator.Register("test-bus", nil)

		require.NoError(t, err)

		retrievedBus, found := locator.Get("test-bus")
		assert.True(t, found)
		assert.Nil(t, retrievedBus)
	})
}

func TestLocator_Get(t *testing.T) {
	t.Run("get existing bus", func(t *testing.T) {
		locator := bus.NewLocator()
		b := bus.NewBus()

		err := locator.Register("test-bus", b)
		require.NoError(t, err)

		retrievedBus, found := locator.Get("test-bus")

		assert.True(t, found)
		assert.Equal(t, b, retrievedBus)
	})

	t.Run("get non-existing bus", func(t *testing.T) {
		locator := bus.NewLocator()

		retrievedBus, found := locator.Get("non-existing")

		assert.False(t, found)
		assert.Nil(t, retrievedBus)
	})

	t.Run("get bus with empty name", func(t *testing.T) {
		locator := bus.NewLocator()

		retrievedBus, found := locator.Get("")

		assert.False(t, found)
		assert.Nil(t, retrievedBus)
	})

	t.Run("get after multiple registrations", func(t *testing.T) {
		locator := bus.NewLocator()
		bus1 := bus.NewBus()
		bus2 := bus.NewBus(&helpers.TestMiddleware{})
		bus3 := bus.NewBus(&helpers.TestMiddleware{}, &helpers.ErrorMiddleware{})

		require.NoError(t, locator.Register("bus1", bus1))
		require.NoError(t, locator.Register("bus2", bus2))
		require.NoError(t, locator.Register("bus3", bus3))

		retrievedBus1, found1 := locator.Get("bus1")
		assert.True(t, found1)
		assert.Equal(t, bus1, retrievedBus1)

		retrievedBus2, found2 := locator.Get("bus2")
		assert.True(t, found2)
		assert.Equal(t, bus2, retrievedBus2)

		retrievedBus3, found3 := locator.Get("bus3")
		assert.True(t, found3)
		assert.Equal(t, bus3, retrievedBus3)

		_, found4 := locator.Get("bus4")
		assert.False(t, found4)
	})
}

func TestLocator_GetAll(t *testing.T) {
	t.Run("get all from empty locator", func(t *testing.T) {
		locator := bus.NewLocator()

		all := locator.GetAll()

		assert.Empty(t, all)
		assert.NotNil(t, all)
	})

	t.Run("get all with single bus", func(t *testing.T) {
		locator := bus.NewLocator()
		b := bus.NewBus()

		require.NoError(t, locator.Register("test-bus", b))

		all := locator.GetAll()

		assert.Len(t, all, 1)
		assert.Contains(t, all, b)
	})

	t.Run("get all with multiple buses", func(t *testing.T) {
		locator := bus.NewLocator()
		bus1 := bus.NewBus()
		bus2 := bus.NewBus(&helpers.TestMiddleware{})
		bus3 := bus.NewBus(&helpers.TestMiddleware{}, &helpers.ErrorMiddleware{})

		require.NoError(t, locator.Register("bus1", bus1))
		require.NoError(t, locator.Register("bus2", bus2))
		require.NoError(t, locator.Register("bus3", bus3))

		all := locator.GetAll()

		assert.Len(t, all, 3)
		assert.Contains(t, all, bus1)
		assert.Contains(t, all, bus2)
		assert.Contains(t, all, bus3)
	})

	t.Run("get all with nil bus", func(t *testing.T) {
		locator := bus.NewLocator()
		b := bus.NewBus()

		require.NoError(t, locator.Register("real-bus", b))
		require.NoError(t, locator.Register("nil-bus", nil))

		all := locator.GetAll()

		assert.Len(t, all, 2)
		assert.Contains(t, all, b)
		assert.Contains(t, all, nil)
	})

	t.Run("get all returns copy of buses", func(t *testing.T) {
		locator := bus.NewLocator()
		bus1 := bus.NewBus()
		bus2 := bus.NewBus()

		require.NoError(t, locator.Register("bus1", bus1))
		require.NoError(t, locator.Register("bus2", bus2))

		all1 := locator.GetAll()
		all2 := locator.GetAll()

		assert.Equal(t, all1, all2)
		assert.NotSame(t, &all1, &all2)

		all1[0] = nil
		all3 := locator.GetAll()
		assert.NotEqual(t, all1, all3)
		assert.Equal(t, all2, all3)
	})
}

func TestLocator_Integration(t *testing.T) {
	t.Run("full workflow with multiple operations", func(t *testing.T) {
		locator := bus.NewLocator()

		all := locator.GetAll()
		assert.Empty(t, all)

		defaultBus := bus.NewBus()
		asyncBus := bus.NewBus(&helpers.TestMiddleware{})

		require.NoError(t, locator.Register("default", defaultBus))
		require.NoError(t, locator.Register("async", asyncBus))

		all = locator.GetAll()
		assert.Len(t, all, 2)

		retrievedDefault, foundDefault := locator.Get("default")
		assert.True(t, foundDefault)
		assert.Equal(t, defaultBus, retrievedDefault)

		retrievedAsync, foundAsync := locator.Get("async")
		assert.True(t, foundAsync)
		assert.Equal(t, asyncBus, retrievedAsync)

		newDefaultBus := bus.NewBus(&helpers.ErrorMiddleware{})
		require.NoError(t, locator.Register("default", newDefaultBus))

		all = locator.GetAll()
		assert.Len(t, all, 2)

		retrievedNewDefault, found := locator.Get("default")
		assert.True(t, found)
		assert.Equal(t, newDefaultBus, retrievedNewDefault)
		assert.NotEqual(t, defaultBus, retrievedNewDefault)
	})
}
