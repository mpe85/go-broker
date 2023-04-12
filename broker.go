// Package broker provides a simple generic thread-safe message broker implementation for Go 1.18+.
package broker

import (
	"errors"
	"time"
)

// Client defines a client that is registered to the broker and receives messages from it.
type Client[T any] chan T

// void represents an empty struct that consumes no memory.
type void struct{}

// Broker broadcasts messages to registered clients
type Broker[T any] struct {
	clients              map[Client[T]]void
	stop                 chan void
	subscribingClients   chan Client[T]
	unsubscribingClients chan Client[T]
	messages             chan T
	timeout              time.Duration
}

// Builder encapsulates the construction of a new broker.
type Builder[T any] struct {
	timeout    time.Duration
	bufferSize int
}

// defaultTimeout specifies the default timeout when the broker tries to send a message to a client,
// a message is published to the broker, or a client subscribes or unsubscribes.
const defaultTimeout = time.Second

// defaultBufferSize specifies the default size of the message buffer.
const defaultBufferSize int = 10

// ErrTimeout is the error returned when a broker operation timed out.
var ErrTimeout = errors.New("timeout")

// Publish publishes a message to the broker.
// Returns ErrTimeout on timeout.
func (broker *Broker[T]) Publish(message T) error {
	select {
	case broker.messages <- message:
		return nil
	case <-time.After(broker.timeout):
		return ErrTimeout
	}
}

// Subscribe registers a new client to the broker and returns it to the caller.
// Returns ErrTimeout on timeout.
func (broker *Broker[T]) Subscribe() (Client[T], error) {
	client := make(Client[T])
	select {
	case broker.subscribingClients <- client:
		return client, nil
	case <-time.After(broker.timeout):
		return nil, ErrTimeout
	}
}

// Unsubscribe removes a client from the broker.
// Returns ErrTimeout on timeout.
func (broker *Broker[T]) Unsubscribe(client Client[T]) error {
	select {
	case broker.unsubscribingClients <- client:
		return nil
	case <-time.After(broker.timeout):
		return ErrTimeout
	}
}

// Close stops the broker and removes all leftover clients from it.
// Panics when the broker is already stopped.
func (broker *Broker[T]) Close() {
	close(broker.stop)
}

// run starts the broker loop.
func (broker *Broker[T]) run() {
	for {
		select {
		case <-broker.stop:
			// close all leftover clients and break the broker loop
			for client := range broker.clients {
				close(client)
			}
			return
		case client := <-broker.subscribingClients:
			// add new client
			broker.clients[client] = void{}
		case client := <-broker.unsubscribingClients:
			// remove and close client
			delete(broker.clients, client)
			close(client)
		case msg := <-broker.messages:
			// broadcast published message to all clients
			for client := range broker.clients {
				// send message to client (or discard message after timeout)
				select {
				case client <- msg:
				case <-time.After(broker.timeout):
				}
			}
		}
	}
}

// NewBuilder constructs a new builder.
func NewBuilder[T any]() Builder[T] {
	return Builder[T]{defaultTimeout, defaultBufferSize}
}

// New constructs a new broker with default configuration:
//   - timeout = 1 * time.Second
//   - bufferSize = 10
func New[T any]() *Broker[T] {
	return NewBuilder[T]().Build()
}

// Timeout configures the broker timeout.
func (builder Builder[T]) Timeout(timeout time.Duration) Builder[T] {
	builder.timeout = timeout
	return builder
}

// BufferSize configures the message buffer size.
func (builder Builder[T]) BufferSize(bufferSize int) Builder[T] {
	builder.bufferSize = bufferSize
	return builder
}

// Build builds a new broker using the configuration of the builder.
func (builder Builder[T]) Build() *Broker[T] {
	broker := &Broker[T]{
		clients:              make(map[Client[T]]void),
		stop:                 make(chan void),
		subscribingClients:   make(chan Client[T]),
		unsubscribingClients: make(chan Client[T]),
		messages:             make(chan T, builder.bufferSize),
		timeout:              builder.timeout,
	}
	go broker.run()
	return broker
}
