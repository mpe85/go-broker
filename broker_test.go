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

	assertions.Nil(broker.Close())
}

func TestNewBuilderDefault(t *testing.T) {
	assertions := assert.New(t)

	broker := NewBuilder[int]().Build()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	assertions.Nil(broker.Close())
}

func TestNewBuilderTimeout(t *testing.T) {
	assertions := assert.New(t)

	timeout := time.Millisecond
	broker := NewBuilder[int]().Timeout(timeout).Build()
	assertions.NotNil(broker)
	assertions.Equal(timeout, broker.timeout)
	assertions.Equal(defaultBufferSize, cap(broker.messages))

	assertions.Nil(broker.Close())
}

func TestNewBuilderBufferSize(t *testing.T) {
	assertions := assert.New(t)

	bufferSize := 100
	broker := NewBuilder[int]().BufferSize(bufferSize).Build()
	assertions.NotNil(broker)
	assertions.Equal(defaultTimeout, broker.timeout)
	assertions.Equal(bufferSize, cap(broker.messages))

	assertions.Nil(broker.Close())
}

func TestSubscribe(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := New[int]()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client.Messages())
	assertions.Nil(err)

	go func() {
		assertions.Nil(broker.Publish(answer))
	}()

	msg, ok := <-client.Messages()
	assertions.Equal(answer, msg)
	assertions.True(ok)

	assertions.Nil(broker.Close())

	client, err = broker.Subscribe()
	assertions.Nil(client.Messages())
	assertions.ErrorIs(err, ErrClosed)
}

func TestUnsubscribe(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := New[int]()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client.Messages())
	assertions.Nil(err)

	assertions.Nil(broker.Unsubscribe(client))

	go func() {
		assertions.Nil(broker.Publish(answer))
	}()

	msg, ok := <-client.Messages()
	assertions.Equal(0, msg)
	assertions.False(ok)

	client, err = broker.Subscribe()
	assertions.NotNil(client.Messages())
	assertions.Nil(err)

	assertions.Nil(broker.Close())

	assertions.ErrorIs(broker.Unsubscribe(client), ErrClosed)
}

func TestClose(t *testing.T) {
	assertions := assert.New(t)

	answer := 42
	broker := New[int]()
	assertions.NotNil(broker)

	client, err := broker.Subscribe()
	assertions.NotNil(client.Messages())
	assertions.Nil(err)

	assertions.Nil(broker.Close())
	assertions.ErrorIs(broker.Close(), ErrClosed)

	assertions.ErrorIs(broker.Publish(answer), ErrClosed)

	msg, ok := <-client.Messages()
	assertions.Equal(0, msg)
	assertions.False(ok)

	client, err = broker.Subscribe()
	assertions.Nil(client.Messages())
	assertions.ErrorIs(err, ErrClosed)
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
	case <-client.Messages():
		assertions.Fail("Received message not expected")
	case <-time.After(time.Second):
	}

	assertions.Nil(broker.Close())
}

func BenchmarkNew(b *testing.B) {
	assertions := assert.New(b)
	for i := 0; i < b.N; i++ {
		broker := New[int]()
		assertions.NotNil(broker)
		b.Cleanup(func() {
			_ = broker.Close()
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
		_ = broker.Close()
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
		_ = broker.Close()
	})
}
