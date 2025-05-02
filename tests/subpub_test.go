package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kudras3r/VKSubPub/internal/subpub"
)

func TestSubscribeAndPublish(t *testing.T) {
	sp := subpub.NewSubPub()

	var wg sync.WaitGroup
	wg.Add(1)

	_, err := sp.Subscribe("test", func(msg interface{}) {
		defer wg.Done()
		if msg != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got %v", msg)
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	err = sp.Publish("test", "Hello, World!")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	wg.Wait()
}

func TestFIFOOrder(t *testing.T) {
	sp := subpub.NewSubPub()

	var received []string
	var mu sync.Mutex

	_, err := sp.Subscribe("fifo", func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.(string))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	messages := []string{"msg1", "msg2", "msg3"}
	for _, msg := range messages {
		err = sp.Publish("fifo", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	for i, msg := range messages {
		if received[i] != msg {
			t.Errorf("Expected %v, got %v", msg, received[i])
		}
	}
}

func TestExtraChFIFOOrder(t *testing.T) {
	sp := subpub.NewSubPub()

	var received []string
	var mu sync.Mutex

	_, err := sp.Subscribe("extra", func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.(string))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	messages := []string{"msg1", "msg2", "msg3", "msg4", "msg5"}
	for _, msg := range messages {
		err = sp.Publish("extra", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	for i, msg := range messages {
		if received[i] != msg {
			t.Errorf("Expected %v, got %v", msg, received[i])
		}
	}
}

func TestSlowSubscriber(t *testing.T) {
	sp := subpub.NewSubPub()

	var wg sync.WaitGroup
	wg.Add(1)

	_, err := sp.Subscribe("slow", func(msg interface{}) {
		time.Sleep(500 * time.Millisecond) // Медленный обработчик
		wg.Done()
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	start := time.Now()

	err = sp.Publish("slow", "test message")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	wg.Wait()

	elapsed := time.Since(start)
	if elapsed > time.Second {
		t.Errorf("Slow subscriber blocked too long: %v", elapsed)
	}
}

func TestUnsubscribe(t *testing.T) {
	sp := subpub.NewSubPub()

	var called int
	var mu sync.Mutex

	sub, err := sp.Subscribe("unsubscribe", func(msg interface{}) {
		mu.Lock()
		called++
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	sub.Unsubscribe()

	err = sp.Publish("unsubscribe", "test message")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if called != 0 {
		t.Errorf("Handler was called after unsubscribe")
	}
}

func TestClose(t *testing.T) {
	return

	sp := subpub.NewSubPub()

	_, err := sp.Subscribe("close", func(msg interface{}) {

	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err = sp.Close(ctx)
	if err != nil {
		t.Fatalf("Failed to close: %v", err)
	}
}
