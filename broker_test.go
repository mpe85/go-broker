package broker

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNew(t *testing.T) {
	assertions := assert.New(t)

	broker := New[int]()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	t.Cleanup(func() {
		broker.Close()
	})
}

func TestNewBuilder(t *testing.T) {
	assertions := assert.New(t)

	broker := NewBuilder[int]().Build()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	t.Cleanup(func() {
		broker.Close()
	})
}

func TestNewBuilderTimeout(t *testing.T) {
	assertions := assert.New(t)

	timeout := time.Millisecond

	broker := NewBuilder[int]().Timeout(timeout).Build()
	assertions.NotNil(broker)
	assertions.Equal(timeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	t.Cleanup(func() {
		broker.Close()
	})
}

func TestNewBuilderBufferSize(t *testing.T) {
	assertions := assert.New(t)

	bufferSize := 100

	broker := NewBuilder[int]().BufferSize(bufferSize).Build()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(bufferSize, cap(broker.messages))

	t.Cleanup(func() {
		broker.Close()
	})
}

func BenchmarkPublish(b *testing.B) {
	broker := New[int]()
	for i := 0; i < b.N; i++ {
		broker.Publish(i)
	}

	b.Cleanup(func() {
		broker.Close()
	})
}

func TestSubscribe(t *testing.T) {
	assertions := assert.New(t)
	msg := 4711

	broker := New[int]()
	assertions.NotNil(broker)

	client := broker.Subscribe()
	assertions.NotNil(client)

	go broker.Publish(msg)

	m, ok := <-client
	assertions.Equal(msg, m)
	assertions.True(ok)

	t.Cleanup(func() {
		broker.Close()
	})
}

func TestUnsubscribe(t *testing.T) {
	assertions := assert.New(t)
	msg := 4711

	broker := New[int]()
	assertions.NotNil(broker)

	client := broker.Subscribe()
	assertions.NotNil(client)

	broker.Unsubscribe(client)

	go broker.Publish(msg)

	m, ok := <-client
	assertions.Equal(0, m)
	assertions.False(ok)

	t.Cleanup(func() {
		broker.Close()
	})
}

func ExampleNewBuilder() {
	NewBuilder[string]().Timeout(10 * time.Millisecond).BufferSize(100).Build()
}
