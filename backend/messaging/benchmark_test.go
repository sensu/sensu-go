package messaging

import (
	"context"
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
			ch := channelSubscriber{make(chan interface{}, 1000)}
			go func(client string, ch channelSubscriber) {
				subsc, _ := bus.Subscribe(topicName, client, ch)
				for {
					select {
					case <-ch.Channel:
					case <-done:
						wg.Done()
						_ = subsc.Cancel()
						close(ch.Channel)
						return
					}
				}
			}(fmt.Sprintf("client-%d", i), ch)
		}
		return done

	}

	for _, tc := range tt {
		b.Run(fmt.Sprintf("%d-clients", tc), func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			bus, _ := NewWizardBus(WizardBusConfig{})
			_ = bus.Start(ctx)

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
