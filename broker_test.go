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

	t.Cleanup(broker.Close)
}

func TestNewBuilderDefault(t *testing.T) {
	assertions := assert.New(t)

	broker := NewBuilder[int]().Build()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	t.Cleanup(broker.Close)
}

func TestNewBuilderTimeout(t *testing.T) {
	assertions := assert.New(t)

	timeout := time.Millisecond
	broker := NewBuilder[int]().Timeout(timeout).Build()
	assertions.NotNil(broker)
	assertions.Equal(timeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	t.Cleanup(broker.Close)
}

func TestNewBuilderBufferSize(t *testing.T) {
	assertions := assert.New(t)

	bufferSize := 100
	broker := NewBuilder[int]().BufferSize(bufferSize).Build()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(bufferSize, cap(broker.messages))

	t.Cleanup(broker.Close)
}

func TestSubscribe(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := NewBuilder[int]().Timeout(100 * time.Millisecond).Build()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client)
	assertions.Nil(err)

	go func() {
		assertions.Nil(broker.Publish(answer))
	}()

	msg, ok := <-client
	assertions.Equal(answer, msg)
	assertions.True(ok)

	broker.Close()

	client, err = broker.Subscribe()
	assertions.Nil(client)
	assertions.ErrorIs(err, ErrTimeout)
}

func TestUnsubscribe(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := NewBuilder[int]().Timeout(100 * time.Millisecond).Build()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client)
	assertions.Nil(err)

	assertions.Nil(broker.Unsubscribe(client))

	go func() {
		assertions.Nil(broker.Publish(answer))
	}()

	msg, ok := <-client
	assertions.Equal(0, msg)
	assertions.False(ok)

	client, err = broker.Subscribe()
	assertions.NotNil(client)
	assertions.Nil(err)

	broker.Close()

	time.Sleep(time.Second)
	assertions.ErrorIs(broker.Unsubscribe(client), ErrTimeout)
}

func TestClose(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := NewBuilder[int]().Timeout(100 * time.Millisecond).BufferSize(0).Build()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client)
	assertions.Nil(err)

	broker.Close()
	assertions.Panics(broker.Close)

	time.Sleep(time.Second)

	assertions.ErrorIs(broker.Publish(answer), ErrTimeout)

	msg, ok := <-client
	assertions.Equal(0, msg)
	assertions.False(ok)

	client, err = broker.Subscribe()
	assertions.Nil(client)
	assertions.ErrorIs(err, ErrTimeout)
}

func TestPublishTimeout(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := NewBuilder[int]().Timeout(100 * time.Millisecond).Build()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client)
	assertions.Nil(err)

	go func() {
		assertions.Nil(broker.Publish(answer))
	}()

	time.Sleep(200 * time.Millisecond)

	select {
	case <-client:
		assertions.Fail("Received message not expected")
	case <-time.After(time.Second):
	}

	broker.Close()
}

func BenchmarkNew(b *testing.B) {
	assertions := assert.New(b)
	for i := 0; i < b.N; i++ {
		broker := New[int]()
		assertions.NotNil(broker)
		b.Cleanup(func() {
			broker.Close()
		})
	}
}

func BenchmarkPublish(b *testing.B) {
	assertions := assert.New(b)
	broker := New[int]()
	assertions.NotNil(broker)
	for i := 0; i < b.N; i++ {
		assertions.Nil(broker.Publish(i))
	}
	b.Cleanup(func() {
		broker.Close()
	})
}

func BenchmarkSubscribe(b *testing.B) {
	assertions := assert.New(b)
	broker := New[int]()
	for i := 0; i < b.N; i++ {
		client, err := broker.Subscribe()
		assertions.NotNil(client)
		assertions.Nil(err)
	}
	b.Cleanup(func() {
		broker.Close()
	})
}
