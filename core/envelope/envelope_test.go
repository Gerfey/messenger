package envelope_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gerfey/messenger/core/envelope"
	"github.com/gerfey/messenger/tests/helpers"
)

func TestNewEnvelope(t *testing.T) {
	t.Run("with struct message", func(t *testing.T) {
		msg := helpers.NewTestMessage("1", "test content")
		env := envelope.NewEnvelope(msg)

		assert.Equal(t, msg, env.Message())
		assert.Empty(t, env.Stamps())
		assert.NotNil(t, env)
	})

	t.Run("with string message", func(t *testing.T) {
		msg := "simple string message"
		env := envelope.NewEnvelope(msg)

		assert.Equal(t, msg, env.Message())
		assert.Empty(t, env.Stamps())
	})

	t.Run("with nil message", func(t *testing.T) {
		msg := envelope.NewEnvelope(nil)

		assert.Nil(t, msg.Message())
		assert.Empty(t, msg.Stamps())
	})

	t.Run("with interface message", func(t *testing.T) {
		var msg any = helpers.NewComplexMessage("123", "test")
		env := envelope.NewEnvelope(msg)

		assert.Equal(t, msg, env.Message())
		assert.Empty(t, env.Stamps())
	})

	t.Run("with numeric message", func(t *testing.T) {
		msg := 42
		env := envelope.NewEnvelope(msg)

		assert.Equal(t, msg, env.Message())
		assert.Empty(t, env.Stamps())
	})
}

func TestEnvelope_Message(t *testing.T) {
	t.Run("return original message", func(t *testing.T) {
		msg := helpers.NewTestMessage("1", "content")
		env := envelope.NewEnvelope(msg)

		result := env.Message()

		assert.Equal(t, msg, result)
		assert.Same(t, msg, result)
	})

	t.Run("message unchanged after adding stamps", func(t *testing.T) {
		msg := helpers.NewTestMessage("1", "content")
		env := envelope.NewEnvelope(msg)

		envWithStamp := env.WithStamp(helpers.TestStamp{Value: "test"})

		assert.Equal(t, msg, env.Message())
		assert.Equal(t, msg, envWithStamp.Message())
		assert.Same(t, env.Message(), envWithStamp.Message())
	})
}

func TestEnvelope_WithStamp(t *testing.T) {
	t.Run("adds single stamp", func(t *testing.T) {
		msg := helpers.NewTestMessage("1", "content")
		env := envelope.NewEnvelope(msg)
		stamp := helpers.TestStamp{Value: "test stamp"}

		envWithStamp := env.WithStamp(stamp)

		assert.Len(t, envWithStamp.Stamps(), 1)
		assert.Equal(t, stamp, envWithStamp.Stamps()[0])
	})

	t.Run("immutability - original envelope unchanged", func(t *testing.T) {
		msg := helpers.NewTestMessage("1", "content")
		original := envelope.NewEnvelope(msg)
		stamp := helpers.TestStamp{Value: "test stamp"}

		withStamp := original.WithStamp(stamp)

		assert.Len(t, original.Stamps(), 0, "Original envelope should not be modified")
		assert.Len(t, withStamp.Stamps(), 1, "New envelope should have the stamp")

		assert.Equal(t, original.Message(), withStamp.Message())
		assert.NotSame(t, original, withStamp)
	})

	t.Run("adds multiple stamps in sequence", func(t *testing.T) {
		msg := helpers.NewTestMessage("1", "content")
		env := envelope.NewEnvelope(msg)

		env1 := env.WithStamp(helpers.TestStamp{Value: "first"})
		env2 := env1.WithStamp(helpers.AnotherStamp{Number: 42})
		env3 := env2.WithStamp(helpers.TestStamp{Value: "second"})

		assert.Len(t, env.Stamps(), 0)
		assert.Len(t, env1.Stamps(), 1)
		assert.Len(t, env2.Stamps(), 2)
		assert.Len(t, env3.Stamps(), 3)

		stamps := env3.Stamps()
		assert.Equal(t, helpers.TestStamp{Value: "first"}, stamps[0])
		assert.Equal(t, helpers.AnotherStamp{Number: 42}, stamps[1])
		assert.Equal(t, helpers.TestStamp{Value: "second"}, stamps[2])
	})

	t.Run("preserves stamp order", func(t *testing.T) {
		msg := "test"
		env := envelope.NewEnvelope(msg)

		stamp1 := helpers.TestStamp{Value: "first"}
		stamp2 := helpers.AnotherStamp{Number: 42}
		stamp3 := helpers.TestStamp{Value: "second"}

		result := env.WithStamp(stamp1).WithStamp(stamp2).WithStamp(stamp3)

		stamps := result.Stamps()

		assert.Len(t, stamps, 3)
		assert.Equal(t, stamp1, stamps[0])
		assert.Equal(t, stamp2, stamps[1])
		assert.Equal(t, stamp3, stamps[2])
	})
}

func TestEnvelope_Stamps(t *testing.T) {
	t.Run("empty envelope", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		stamps := env.Stamps()

		assert.Empty(t, stamps)
		assert.NotNil(t, stamps)
	})

	t.Run("envelope with stamps", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		stamp1 := helpers.TestStamp{Value: "first"}
		stamp2 := helpers.AnotherStamp{Number: 42}

		envWithStamps := env.WithStamp(stamp1).WithStamp(stamp2)
		stamps := envWithStamps.Stamps()

		assert.Len(t, stamps, 2)
		assert.Equal(t, stamp1, stamps[0])
		assert.Equal(t, stamp2, stamps[1])
	})
}

func TestStampsOf(t *testing.T) {
	t.Run("filters existing stamp type", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "first"})
		env = env.WithStamp(helpers.AnotherStamp{Number: 42})
		env = env.WithStamp(helpers.TestStamp{Value: "second"})

		testStamps := envelope.StampsOf[helpers.TestStamp](env)

		assert.Len(t, testStamps, 2)
		assert.Equal(t, helpers.TestStamp{Value: "first"}, testStamps[0])
		assert.Equal(t, helpers.TestStamp{Value: "second"}, testStamps[1])
	})

	t.Run("filters non-existing stamp type", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "test"})

		anotherStamps := envelope.StampsOf[helpers.AnotherStamp](env)

		assert.Empty(t, anotherStamps)
		assert.NotNil(t, anotherStamps)
	})

	t.Run("empty envelope", func(t *testing.T) {
		env := envelope.NewEnvelope("test")

		stamps := envelope.StampsOf[helpers.TestStamp](env)

		assert.Empty(t, stamps)
		assert.NotNil(t, stamps)
	})

	t.Run("multiple stamps of same type preserve order", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		stamp1 := helpers.TestStamp{Value: "first"}
		stamp2 := helpers.TestStamp{Value: "second"}
		stamp3 := helpers.TestStamp{Value: "third"}

		env = env.WithStamp(stamp1).WithStamp(helpers.AnotherStamp{Number: 42}).WithStamp(stamp2).WithStamp(stamp3)

		testStamps := envelope.StampsOf[helpers.TestStamp](env)

		assert.Len(t, testStamps, 3)
		assert.Equal(t, "first", testStamps[0].Value)
		assert.Equal(t, "second", testStamps[1].Value)
		assert.Equal(t, "third", testStamps[2].Value)
	})
}

func TestFirstStampOf(t *testing.T) {
	t.Run("stamp found - returns first occurrence", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		firstStamp := helpers.TestStamp{Value: "first"}
		secondStamp := helpers.TestStamp{Value: "second"}

		env = env.WithStamp(firstStamp).WithStamp(helpers.AnotherStamp{Number: 42}).WithStamp(secondStamp)

		stamp, found := envelope.FirstStampOf[helpers.TestStamp](env)

		assert.True(t, found)
		assert.Equal(t, firstStamp, stamp)
		assert.Equal(t, "first", stamp.Value)
	})

	t.Run("stamp not found", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "test"})

		stamp, found := envelope.FirstStampOf[helpers.AnotherStamp](env)

		assert.False(t, found)
		assert.Equal(t, helpers.AnotherStamp{}, stamp)
		assert.Equal(t, 0, stamp.Number)
	})

	t.Run("empty envelope", func(t *testing.T) {
		env := envelope.NewEnvelope("test")

		stamp, found := envelope.FirstStampOf[helpers.TestStamp](env)

		assert.False(t, found)
		assert.Equal(t, helpers.TestStamp{}, stamp)
	})

	t.Run("single stamp of requested type", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		testStamp := helpers.TestStamp{Value: "only one"}
		env = env.WithStamp(testStamp)

		stamp, found := envelope.FirstStampOf[helpers.TestStamp](env)

		assert.True(t, found)
		assert.Equal(t, testStamp, stamp)
	})
}

func TestLastStampOf(t *testing.T) {
	t.Run("stamp found - returns last occurrence", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		firstStamp := helpers.TestStamp{Value: "first"}
		lastStamp := helpers.TestStamp{Value: "last"}

		env = env.WithStamp(firstStamp).WithStamp(helpers.AnotherStamp{Number: 42}).WithStamp(lastStamp)

		stamp, found := envelope.LastStampOf[helpers.TestStamp](env)

		assert.True(t, found)
		assert.Equal(t, lastStamp, stamp)
		assert.Equal(t, "last", stamp.Value)
	})

	t.Run("stamp not found", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "test"})

		stamp, found := envelope.LastStampOf[helpers.AnotherStamp](env)

		assert.False(t, found)
		assert.Equal(t, helpers.AnotherStamp{}, stamp)
	})

	t.Run("empty envelope", func(t *testing.T) {
		env := envelope.NewEnvelope("test")

		stamp, found := envelope.LastStampOf[helpers.TestStamp](env)

		assert.False(t, found)
		assert.Equal(t, helpers.TestStamp{}, stamp)
	})

	t.Run("single stamp of requested type", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		testStamp := helpers.TestStamp{Value: "only one"}
		env = env.WithStamp(testStamp)

		stamp, found := envelope.LastStampOf[helpers.TestStamp](env)

		assert.True(t, found)
		assert.Equal(t, testStamp, stamp)
	})
}

func TestHasStampOf(t *testing.T) {
	t.Run("stamp exists", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "test"})
		env = env.WithStamp(helpers.AnotherStamp{Number: 42})

		assert.True(t, envelope.HasStampOf[helpers.TestStamp](env))
		assert.True(t, envelope.HasStampOf[helpers.AnotherStamp](env))
	})

	t.Run("stamp does not exist", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "test"})

		assert.False(t, envelope.HasStampOf[helpers.AnotherStamp](env))
	})

	t.Run("empty envelope", func(t *testing.T) {
		env := envelope.NewEnvelope("test")

		assert.False(t, envelope.HasStampOf[helpers.TestStamp](env))
		assert.False(t, envelope.HasStampOf[helpers.AnotherStamp](env))
	})

	t.Run("multiple stamps of same type", func(t *testing.T) {
		env := envelope.NewEnvelope("test")
		env = env.WithStamp(helpers.TestStamp{Value: "first"})
		env = env.WithStamp(helpers.TestStamp{Value: "second"})

		assert.True(t, envelope.HasStampOf[helpers.TestStamp](env))
	})
}
