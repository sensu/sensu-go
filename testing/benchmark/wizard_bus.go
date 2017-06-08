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
			// Start the bus.
			bus := &messaging.WizardBus{}
			bus.Start()

			// resultMap[ client ID ] map[ metric name ] value
			resultMap := map[string]map[string][]float64{}
			resultMutex := &sync.Mutex{}

			// setup our topics
			topics := make([]string, numTopics)
			ticker := time.NewTicker(time.Duration(tickTime) * time.Millisecond)

			for i := 0; i < numTopics; i++ {
				topicName := fmt.Sprintf("topic%d", i)
				topics[i] = topicName
			}

			// start up our subscribers
			clientWG := &sync.WaitGroup{}
			clientWG.Add(numClients)

			channels := map[string]chan interface{}{}
			for i := 0; i < numClients; i++ {
				// subscriber id
				uid, _ := uuid.NewRandom()
				clientID := uid.String()

				// subscriber channel
				ch := make(chan interface{}, 1000)
				for _, t := range topics {
					bus.Subscribe(t, clientID, ch)
				}
				channels[clientID] = ch

				go func(channel chan interface{}, id string) {
					metrics := map[string][]float64{}
					metrics["received"] = make([]float64, numMessages)
					metrics["received"][0] = 0

					metrics["delay"] = make([]float64, numMessages*numTopics)

					msgCounter := 0
					for msg := range channel {
						msgStr := msg.(string)
						received := time.Now().UnixNano()
						sent, _ := strconv.ParseInt(msgStr, 10, 64)
						metrics["delay"][msgCounter] = float64(received) - float64(sent)
						msgCounter++
					}
					metrics["received"][0] = float64(msgCounter)

					resultMutex.Lock()
					resultMap[id] = metrics
					resultMutex.Unlock()
					clientWG.Done()
				}(ch, clientID)
			}

			go func() {
				for i := 0; i < numMessages; i++ {
					tickTime := <-ticker.C
					msg := strconv.FormatInt(tickTime.UnixNano(), 10)
					for _, t := range topics {
						bus.Publish(t, msg)
					}
				}
				ticker.Stop()
				fmt.Println("stopping wizardbus")
				bus.Stop()
				for _, ch := range channels {
					close(ch)
				}
			}()

			clientWG.Wait()

			totalDelay := float64(0)
			totalMessages := 0
			totalReceived := 0
			for _, client := range resultMap {
				for _, d := range client["delay"] {
					totalDelay += d
					totalMessages++
				}
				totalReceived += int(client["received"][0])
			}
			avgDelay := totalDelay / float64(totalMessages)
			fmt.Println("number of clients = ", numClients)
			fmt.Println("number of topics = ", numTopics)
			fmt.Println("total number of messages = ", totalMessages)
			fmt.Println("total messages received = ", totalReceived)
			fmt.Println("messages lost = ", totalMessages-totalReceived)
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
