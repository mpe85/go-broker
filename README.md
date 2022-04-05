# go-broker

[![Latest Release](https://img.shields.io/github/release/mpe85/go-broker/all.svg?label=Latest%20Release)](https://github.com/mpe85/go-broker/releases/latest)
[![Go](https://img.shields.io/github/go-mod/go-version/mpe85/go-broker)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/mpe85/go-broker?style=flat-square)](https://goreportcard.com/report/github.com/mpe85/go-broker)
[![Build](https://github.com/mpe85/go-broker/actions/workflows/test.yml/badge.svg)](https://github.com/mpe85/go-broker/actions/workflows/test.yml)
[![Coverage](https://codecov.io/gh/mpe85/go-broker/branch/master/graph/badge.svg)](https://codecov.io/gh/mpe85/go-broker)
[![License](https://img.shields.io/github/license/mpe85/grampa.svg?label=License)](https://github.com/mpe85/go-broker/blob/master/LICENSE)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/mpe85/go-broker)](https://pkg.go.dev/mod/github.com/mpe85/go-broker)

A generic thread-safe broker for Go 1.18+.

## Installation

```sh
go get github.com/mpe85/go-broker
```

## Example

```go
package main

import (
	"fmt"
	"github.com/mpe85/go-broker"
	"time"
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
