// Package broker provides a simple generic message broker implementation for Go 1.18+.
package broker

import (
	"errors"
	"time"
)

// Client defines a client that is registered to the broker and receives messages from it.
type Client[T any] struct {
	ch chan T
}

// void represents an empty struct that consumes no memory.
type void struct{}

// Broker broadcasts messages to registered clients
type Broker[T any] struct {
	clients        map[chan T]void
	stop           chan void
	newClients     chan chan T
	removedClients chan chan T
	messages       chan T
	timeout        time.Duration
}

// Builder encapsulates the construction of a new broker.
type Builder[T any] struct {
	timeout    time.Duration
	bufferSize int
}

// defaultTimeout specifies the default timeout when the broker tries to send a message to a client.
const defaultTimeout = time.Second

// defaultBufferSize specifies the default size of the message buffer.
const defaultBufferSize int = 10

// ErrClosed is the error returned when the broker is already closed.
var ErrClosed = errors.New("broker already closed")

// Messages returns a read-only channel used to receive incoming messages.
func (client Client[T]) Messages() <-chan T {
	return client.ch
}

// Publish publishes a new message to the broker.
// Returns ErrClosed if the broker is already closed.
func (broker *Broker[T]) Publish(message T) error {
	return safeSend(broker.messages, message)
}

// Subscribe registers a new client to the broker and returns it to the caller.
// Returns ErrClosed if the broker is already closed.
func (broker *Broker[T]) Subscribe() (Client[T], error) {
	ch := make(chan T)
	if err := safeSend(broker.newClients, ch); err != nil {
		return Client[T]{}, err
	}
	return Client[T]{ch}, nil
}

// Unsubscribe removes a client from the broker.
// Returns ErrClosed if the broker is already closed.
func (broker *Broker[T]) Unsubscribe(client Client[T]) error {
	return safeSend(broker.removedClients, client.ch)
}

// Close stops the broker and removes all leftover clients from it.
// Returns ErrClosed if the broker is already closed.
func (broker *Broker[T]) Close() error {
	return safeSend(broker.stop, void{})
}

// run starts the broker loop.
func (broker *Broker[T]) run() {
	for {
		select {
		case <-broker.stop:
			// close all channels
			close(broker.stop)
			close(broker.newClients)
			close(broker.removedClients)
			close(broker.messages)
			// close all leftover clients
			for client := range broker.clients {
				close(client)
			}
			// break the broker loop
			return
		case client := <-broker.newClients:
			// add new client
			broker.clients[client] = void{}
		case client := <-broker.removedClients:
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
// - timeout = 1 * time.Second
// - bufferSize = 10
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
		clients:        make(map[chan T]void),
		stop:           make(chan void),
		newClients:     make(chan chan T),
		removedClients: make(chan chan T),
		messages:       make(chan T, builder.bufferSize),
		timeout:        builder.timeout,
	}
	go broker.run()
	return broker
}

// safeSend sends to a channel without panicking. Returns ErrClosed if the channel is already closed.
func safeSend[T any](ch chan T, msg T) (err error) {
	defer func() {
		if recover() != nil {
			err = ErrClosed
		}
	}()
	ch <- msg
	return nil
}
