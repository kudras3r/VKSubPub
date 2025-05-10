package subpub

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/kudras3r/VKSubPub/pkg/config"
	"github.com/kudras3r/VKSubPub/pkg/logger"
)

func TestSubPubExtraQueueOrder(t *testing.T) {
	var mu sync.Mutex

	msgs := make([]int, 0)
	received := make([]int, 0)

	handler := func(msg interface{}) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg.(int))
		time.Sleep(10 * time.Millisecond)
	}

	sp := NewSubPub(logger.New("debug"), &config.SPConf{
		SubQSize:       10, // ! CHECK
		MaxExSize:      1000,
		DefaultSubsCap: 5,
		DefaultExCap:   5,
	})
	sub, err := sp.Subscribe("test", handler)
	if err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	msgsLen := 10 + 30 // ! CHECK
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

	fmt.Println(received)
	fmt.Println(msgs)
	for i := range msgs {
		if received[i] != msgs[i] {
			t.Errorf("At index %d, expected %d, got %d", i, msgs[i], received[i])
		}
	}
}
