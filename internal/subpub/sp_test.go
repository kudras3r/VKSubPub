package subpub

import (
	"sync"
	"testing"
	"time"
)

func TestSubPubExtraQueueOrder(t *testing.T, received []int, msgs []interface{}) {
	var mu sync.Mutex

	handler := func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.(int))
		time.Sleep(10 * time.Millisecond)
	}

	sp := &subPub{subcrs: make(map[string][]*subscriber)}
	sub, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	msgsLen := conf.SubQSize + 30
	for i := 0; i < msgsLen; i++ {
		msgs = append(msgs, i)
	}
	for _, m := range msgs {
		if err := sp.Publish("test", m); err != nil {
			t.Errorf("Publish failed: %v", err)
		}
	}

	time.Sleep(time.Second)

	sub.Unsubscribe()

	mu.Lock()
	defer mu.Unlock()

	if len(received) != len(msgs) {
		t.Fatalf("Expected %d messages, got %d", len(msgs), len(received))
	}

	for i := range msgs {
		if received[i] != msgs[i] {
			t.Errorf("At index %d, expected %d, got %d", i, msgs[i], received[i])
		}
	}
}
