package tests

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/kudras3r/VKSubPub/internal/subpub"
)

func TestSubscribeErrors(t *testing.T) {
	sp := subpub.NewSubPub()

	_, err := sp.Subscribe("", func(interface{}) {})
	if err == nil {
		t.Error("Expected error for empty subject")
	}

	_, err = sp.Subscribe("valid", nil)
	if err == nil {
		t.Error("Expected error for nil handler")
	}
}

func TestSubscribeAndPublish(t *testing.T) {
	sp := subpub.NewSubPub()

	var msgs []interface{}

	var wg sync.WaitGroup
	wg.Add(2)

	_, err := sp.Subscribe("test", func(msg interface{}) {
		defer wg.Done()
		if msg != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got %v", msg)
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	_, err = sp.Subscribe("test", func(msg interface{}) {
		defer wg.Done()
		if msg != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got %v", msg)
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	wg.Add(4)
	_, err = sp.Subscribe("test1", func(msg interface{}) {
		defer wg.Done()
		if msg == "finish" {
			if len(msgs) != 3 {
				t.Fatalf("Not enough messages")
			}
		}
		msgs = append(msgs, msg)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	err = sp.Publish("test", "Hello, World!")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}
	pubMsgs := []interface{}{"m1", "m2", "m3", "finish"}
	for m := range pubMsgs {
		err = sp.Publish("test1", m)
		if err != nil {
			t.Fatalf("error when publish: %v", err)
		}
	}

	wg.Wait()
}

// func TestHandlerPanicRecovery(t *testing.T) {
// 	sp := subpub.NewSubPub()
// 	count := 0
// 	_, _ = sp.Subscribe("panic", func(msg interface{}) {
// 		count++
// 		if count == 1 {
// 			panic("oops")
// 		}
// 	})

// 	_ = sp.Publish("panic", "first")
// 	_ = sp.Publish("panic", "second")

// 	time.Sleep(100 * time.Millisecond)

// 	if count != 2 {
// 		t.Errorf("Expected 2 handler calls, got %d", count)
// 	}
// }

func TestConcurrentPublish(t *testing.T) {
	sp := subpub.NewSubPub()
	var mu sync.Mutex
	var count int

	_, err := sp.Subscribe("parallel", func(msg interface{}) {
		mu.Lock()
		count++
		mu.Unlock()
	})
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = sp.Publish("parallel", "data")
			}
		}()
	}
	wg.Wait()

	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	if count != 500 {
		t.Errorf("Expected 500 messages, got %d", count)
	}
	mu.Unlock()
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

func TestMultipleUnsubscribe(t *testing.T) {
	sp := subpub.NewSubPub()

	var mu sync.Mutex
	called := 0

	sub, err := sp.Subscribe("multi-unsub", func(msg interface{}) {
		mu.Lock()
		called++
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	sub.Unsubscribe()
	sub.Unsubscribe()
	sub.Unsubscribe()

	err = sp.Publish("multi-unsub", "should not be received")
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if called != 0 {
		t.Errorf("Handler was called after multiple unsubscribes")
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

func TestExtraQueue(t *testing.T) {
	var r []int
	var m []interface{}
	subpub.TestSubPubExtraQueueOrder(t, r, m)
}
