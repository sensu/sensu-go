package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newWizardBusCommand())
}

var (
	numClients  int
	numTopics   int
	numMessages int
	tickTime    int
)

func newWizardBusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bus",
		Short: "benchmark the wizardbus",
		RunE: func(cmd *cobra.Command, args []string) error {
			bus := &messaging.WizardBus{}
			bus.Start()

			resultMap := map[string]map[string][]float64{}
			resultMutex := &sync.Mutex{}

			topics := make([]string, numTopics)

			done := make(chan struct{})
			ticker := time.NewTicker(time.Duration(tickTime) * time.Millisecond)
			go func() {
				for {
					select {
					case <-done:
						ticker.Stop()
						return
					case tm := <-ticker.C:
						wg := &sync.WaitGroup{}
						wg.Add(len(topics))
						for _, t := range topics {
							go func(topic string) {
								msg := strconv.FormatInt(tm.UnixNano(), 10)
								bus.Publish(topic, []byte(msg))
								wg.Done()
							}(t)
						}
						wg.Wait()
					}
				}
			}()

			for i := 0; i < numTopics; i++ {
				topicName := fmt.Sprintf("topic%d", i)
				topics[i] = topicName
			}

			clientWG := &sync.WaitGroup{}
			clientWG.Add(numClients)
			for i := 0; i < numClients; i++ {
				uid, _ := uuid.NewRandom()
				clientID := uid.String()
				ch := make(chan []byte, 100)
				for _, t := range topics {
					bus.Subscribe(t, clientID, ch)
				}
				go func() {
					delay := make([]float64, numMessages)
					for msgCounter := 0; msgCounter < numMessages; msgCounter++ {
						msg := <-ch
						received := time.Now().UnixNano()
						sent, _ := strconv.ParseInt(string(msg), 10, 64)
						delay[msgCounter] = float64(received) - float64(sent)
					}

					for _, t := range topics {
						bus.Unsubscribe(t, clientID)
					}

					metrics := map[string][]float64{}
					metrics["delay"] = delay

					resultMutex.Lock()
					resultMap[clientID] = metrics
					resultMutex.Unlock()
					clientWG.Done()
				}()
			}

			clientWG.Wait()
			close(done)

			totalDelay := float64(0)
			totalMessages := 0
			for _, client := range resultMap {
				for _, d := range client["delay"] {
					totalDelay += d
					totalMessages++
				}
			}
			avgDelay := totalDelay / float64(totalMessages)
			fmt.Println("number of clients = ", numClients)
			fmt.Println("number of topics = ", numTopics)
			fmt.Println("total number of messages = ", totalMessages)
			fmt.Println("message frequency per topic(ms) = ", tickTime)
			fmt.Println("average delay(ms) = ", avgDelay/float64(time.Millisecond))

			return nil
		},
	}

	cmd.Flags().IntVarP(&numClients, "clients", "c", 10, "number of connected clients")
	cmd.Flags().IntVarP(&numTopics, "topics", "t", 10, "number of topics")
	cmd.Flags().IntVarP(&numMessages, "messages", "m", 10, "number of topics")
	cmd.Flags().IntVarP(&tickTime, "frequency", "f", 100, "milliseconds between publishing messages")

	return cmd
}
