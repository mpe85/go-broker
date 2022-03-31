package broker

import (
	"time"
)

// Client defines a client that is registered to the broker
type Client[T any] chan T

// void represents an empty struct that consumes no memory
type void struct{}

// Broker broadcasts events to registered clients
type Broker[T any] struct {
	// stop is cool
	stop chan void
	// published message channel
	messages chan T
	// new clients
	newClients chan Client[T]
	// removed clients
	removedClients chan Client[T]
	// client registry
	clients map[Client[T]]void
	// send timeout
	timeout time.Duration
}

// defaultTimeout specifies the default timeout when the Broker tries to send a message to a Client
const defaultTimeout = time.Second

// defaultBufferSize specifies the default size of the message buffer
const defaultBufferSize uint = 10

// Builder builds a new Broker
type Builder[T any] struct {
	timeout    time.Duration
	bufferSize uint
}

// Timeout configures the broker timeout
func (builder *Builder[T]) Timeout(timeout time.Duration) *Builder[T] {
	builder.timeout = timeout
	return builder
}

// BufferSize configures the message buffer size
func (builder *Builder[T]) BufferSize(bufferSize uint) *Builder[T] {
	builder.bufferSize = bufferSize
	return builder
}

// Build builds a new Broker using the configuration of the Builder
func (builder *Builder[T]) Build() *Broker[T] {
	broker := &Broker[T]{
		stop:           make(chan void),
		messages:       make(chan T, builder.bufferSize),
		newClients:     make(chan Client[T]),
		removedClients: make(chan Client[T]),
		clients:        make(map[Client[T]]void),
		timeout:        builder.timeout,
	}
	go broker.run()
	return broker
}

// NewBuilder constructs a new Builder
func NewBuilder[T any]() *Builder[T] {
	return &Builder[T]{defaultTimeout, defaultBufferSize}
}

// New constructs a new broker with default configuration
func New[T any]() *Broker[T] {
	return NewBuilder[T]().Build()
}

// Publish publishes a new message to the broker
func (broker *Broker[T]) Publish(message T) {
	broker.messages <- message
}

// Subscribe registers a new client to the broker and returns it to the caller
func (broker *Broker[T]) Subscribe() Client[T] {
	client := make(Client[T])
	broker.newClients <- client
	return client
}

// Unsubscribe removes a client from the broker
func (broker *Broker[T]) Unsubscribe(client Client[T]) {
	broker.removedClients <- client
}

// Close stops the broker and removes all leftover clients from it
func (broker *Broker[T]) Close() {
	close(broker.stop)
}

// run starts the broker loop
func (broker *Broker[T]) run() {
	for {
		select {
		case <-broker.stop:
			for client := range broker.clients {
				close(client)
			}
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
				// send message or discard message after timeout
				select {
				case client <- msg:
				case <-time.After(broker.timeout):
				}
			}
		}
	}
}
