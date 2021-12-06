package poll

import (
	"fmt"
	"sync"
	"testing"
)

func TestPoller_Wakeup(t *testing.T) {
	poll, err := NewPoller()
	if err != nil {
		t.Fatal(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go block(poll, t, wg)

	poll.Wakeup()

	wg.Wait()
}

func block(poll *Poller, t *testing.T, wg *sync.WaitGroup) {
	events, err := poll.Wait(100000000000)
	if err != nil {
		t.Fatal(err)
	}

	if len(events) == 0 {
		t.Fail()
	}

	fmt.Println("complete")
	wg.Done()
}
