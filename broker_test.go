package broker

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	msg := 4711
	b := NewBuilder[int]().Timeout(10 * time.Millisecond).BufferSize(5).Build()
	c := b.Subscribe()
	go b.Publish(msg)

	if <-c != msg {
		t.Fatal("Received wrong value")
	}

	b.Close()
}
