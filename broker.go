// Package broker provides a simple generic message broker implementation for Go 1.18+.
package broker

import (
	"sync"
	"time"
)

// Client defines a client that is registered to the broker and receives messages from it.
type Client[T any] struct {
	ch        chan T
	closeOnce sync.Once
}

// void represents an empty struct that consumes no memory.
type void struct{}

// Broker broadcasts events to registered clients
type Broker[T any] struct {
	clients        map[*Client[T]]void
	stop           chan void
	newClients     chan *Client[T]
	removedClients chan *Client[T]
	messages       chan T
	timeout        time.Duration
	closeOnce      sync.Once
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

// Channel returns a read-only channel that receives published messages.
func (client *Client[T]) Channel() <-chan T {
	return client.ch
}

// close is closing the client's channel. Repeated calls are no-ops.
func (client *Client[T]) close() {
	client.closeOnce.Do(func() {
		close(client.ch)
	})
}

// Publish publishes a new message to the broker
func (broker *Broker[T]) Publish(message T) {
	broker.messages <- message
}

// Subscribe registers a new client to the broker and returns it to the caller.
func (broker *Broker[T]) Subscribe() *Client[T] {
	client := &Client[T]{ch: make(chan T)}
	broker.newClients <- client
	return client
}

// Unsubscribe removes a client from the broker. Repeated calls are no-ops.
func (broker *Broker[T]) Unsubscribe(client *Client[T]) {
	broker.removedClients <- client
}

// Close stops the broker and removes all leftover clients from it. Repeated calls are no-ops.
func (broker *Broker[T]) Close() {
	broker.closeOnce.Do(func() {
		close(broker.stop)
	})
}

// run starts the broker loop.
func (broker *Broker[T]) run() {
	for {
		select {
		case <-broker.stop:
			// close all leftover clients and break the broker loop
			for client := range broker.clients {
				client.close()
			}
			return
		case client := <-broker.newClients:
			// add new client
			broker.clients[client] = void{}
		case client := <-broker.removedClients:
			// remove and close client
			delete(broker.clients, client)
			client.close()
		case msg := <-broker.messages:
			// broadcast published message to all clients
			for client := range broker.clients {
				// send message or discard message after timeout
				select {
				case client.ch <- msg:
				case <-time.After(broker.timeout):
				}
			}
		}
	}
}

// NewBuilder constructs a new builder.
func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{defaultTimeout, defaultBufferSize}
}

// New constructs a new broker with default configuration:
// - timeout = 1 * time.Second
// - bufferSize = 10
func New[T any]() *Broker[T] {
	return NewBuilder[T]().Build()
}

// Timeout configures the broker timeout.
func (builder *Builder[T]) Timeout(timeout time.Duration) *Builder[T] {
	builder.timeout = timeout
	return builder
}

// BufferSize configures the message buffer size.
func (builder *Builder[T]) BufferSize(bufferSize int) *Builder[T] {
	builder.bufferSize = bufferSize
	return builder
}

// Build builds a new broker using the configuration of the builder.
func (builder *Builder[T]) Build() *Broker[T] {
	broker := &Broker[T]{
		clients:        make(map[*Client[T]]void),
		stop:           make(chan void),
		newClients:     make(chan *Client[T]),
		removedClients: make(chan *Client[T]),
		messages:       make(chan T, builder.bufferSize),
		timeout:        builder.timeout,
	}
	go broker.run()
	return broker
}
