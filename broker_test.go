package broker

import (
	"go.uber.org/goleak"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

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

func ExampleNewBuilder() {
	NewBuilder[string]().Timeout(10 * time.Millisecond).BufferSize(100).Build()
}
