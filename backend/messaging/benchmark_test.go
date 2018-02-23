package messaging

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkWizardBusPublish(b *testing.B) {
	topicName := "topic"

	tt := []int{1, 10, 100, 1000, 10000}

	startClients := func(wg *sync.WaitGroup, bus *WizardBus, numClients int) (done chan struct{}) {
		done = make(chan struct{})
		for i := 0; i < numClients; i++ {
			ch := make(chan interface{}, 1000)
			go func(ch chan interface{}, i int) {
				_ = bus.Subscribe(topicName, string(i), ch)
				for {
					select {
					case <-ch:
					case <-done:
						wg.Done()
						close(ch)
						return
					}
				}
			}(ch, i)
		}
		return done

	}

	for _, tc := range tt {
		b.Run(fmt.Sprintf("%d-clients", tc), func(b *testing.B) {
			bus := &WizardBus{}
			_ = bus.Start()

			wg := &sync.WaitGroup{}
			wg.Add(tc)

			done := startClients(wg, bus, tc)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := bus.Publish(topicName, &i); err != nil {
					b.FailNow()
				}
			}
			close(done)
			wg.Wait()
		})
	}
}
