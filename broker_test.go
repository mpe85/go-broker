package broker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
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

func TestNewBuilderDefault(t *testing.T) {
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

func TestSubscribe(t *testing.T) {
	assertions := assert.New(t)
	answer := 42

	broker := New[int]()
	assertions.NotNil(broker)

	client := broker.Subscribe()
	assertions.NotNil(client)

	go broker.Publish(answer)

	msg, ok := <-client.Channel()
	assertions.Equal(answer, msg)
	assertions.True(ok)

	t.Cleanup(func() {
		broker.Close()
	})
}

func TestUnsubscribe(t *testing.T) {
	assertions := assert.New(t)
	answer := 42

	broker := New[int]()
	assertions.NotNil(broker)

	client := broker.Subscribe()
	assertions.NotNil(client)

	broker.Unsubscribe(client)
	assertions.NotPanics(func() {
		broker.Unsubscribe(client)
	})

	go broker.Publish(answer)

	msg, ok := <-client.Channel()
	assertions.Equal(0, msg)
	assertions.False(ok)

	t.Cleanup(func() {
		broker.Close()
	})
}

func TestClose(t *testing.T) {
	assertions := assert.New(t)
	answer := 42

	broker := New[int]()
	assertions.NotNil(broker)

	client := broker.Subscribe()
	assertions.NotNil(client)

	broker.Close()
	assertions.NotPanics(broker.Close)

	go broker.Publish(answer)

	msg, ok := <-client.Channel()
	assertions.Equal(0, msg)
	assertions.False(ok)
}

func TestPublishTimeout(t *testing.T) {
	assertions := assert.New(t)
	answer := 42

	broker := New[int]()
	assertions.NotNil(broker)

	client := broker.Subscribe()
	assertions.NotNil(client)

	go broker.Publish(answer)

	time.Sleep(broker.timeout + 100*time.Millisecond)

	select {
	case <-client.Channel():
		assertions.Fail("Received message not expected")
	case <-time.After(time.Second):
	}

	t.Cleanup(func() {
		broker.Close()
	})
}

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		broker := New[int]()

		b.Cleanup(func() {
			broker.Close()
		})
	}
}

func BenchmarkPublish(b *testing.B) {
	broker := New[int]()
	for i := 0; i < b.N; i++ {
		go broker.Publish(i)
	}

	b.Cleanup(func() {
		broker.Close()
	})
}

func BenchmarkSubscribe(b *testing.B) {
	broker := New[int]()
	for i := 0; i < b.N; i++ {
		broker.Subscribe()
	}

	b.Cleanup(func() {
		broker.Close()
	})
}

func ExampleNewBuilder() {
	NewBuilder[string]().Timeout(100 * time.Millisecond).BufferSize(50).Build()
}
