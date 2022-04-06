# go-broker

[![Latest Release](https://img.shields.io/github/release/mpe85/go-broker/all.svg?label=Latest%20Release)](https://github.com/mpe85/go-broker/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/mpe85/go-broker)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/mpe85/go-broker?style=flat-square)](https://goreportcard.com/report/github.com/mpe85/go-broker)
[![Build](https://github.com/mpe85/go-broker/actions/workflows/test.yml/badge.svg)](https://github.com/mpe85/go-broker/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/mpe85/go-broker/branch/master/graph/badge.svg?token=rWTO2Fk2jc)](https://codecov.io/gh/mpe85/go-broker)
[![License](https://img.shields.io/github/license/mpe85/grampa.svg?label=License)](https://github.com/mpe85/go-broker/blob/master/LICENSE)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/mpe85/go-broker)](https://pkg.go.dev/mod/github.com/mpe85/go-broker)

A generic thread-safe broker for Go 1.18+.

## Installation

```sh
go get github.com/mpe85/go-broker
```

## Usage

Build a new broker with default configuration:
```go
theBroker := broker.New[string]()
```

Build a new broker with custom configuration:
```go
theBroker := broker.NewBuilder[string]()
	.Timeout(3 * time.Second)
	.BufferSize(100)
	.Build()
```

Subscribe to the broker:
```go
client := theBroker.Subscribe()
```

Unsubscribe from the broker:
```go
theBroker.Unsubscribe(client)
```

Publish a message to the broker:
```go
theBroker.Publish("Hello")
```

Receive a single message from the broker:
```go
message := <-client
```

Receive a single message from the broker, with check if client is closed):
```go
message, ok := <-client
```

Iterate over all messages from the broker, until client is closed):
```go
for message := range client {
	// process message
}
```

Shutdown the broker, and close all clients that are still subscribed:
```go
theBroker.Close()
```

## Example

```go
package main

import (
	"fmt"
	"time"

	"github.com/mpe85/go-broker"
)

func main() {
	theBroker := broker.NewBuilder[int]().
		Timeout(100 * time.Millisecond).
		BufferSize(50).
		Build()

	defer theBroker.Close()

	client1 := theBroker.Subscribe()
	client2 := theBroker.Subscribe()

	go theBroker.Publish(42)

	fmt.Println(<-client1)
	fmt.Println(<-client2)
}
```