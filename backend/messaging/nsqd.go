package messaging

import (
	"errors"
	"fmt"
	"os"
	"sync"

	nsq "github.com/nsqio/go-nsq"
	"github.com/nsqio/nsq/nsqd"
)

const (
	// NSQDTCPAddress is the default TCP listen address for NSQD
	NSQDTCPAddress = "127.0.0.1:4150"
	// NSQDHTTPAddress is the default HTTP listen address for NSQD
	NSQDHTTPAddress = "127.0.0.1:4151"
	// NSQDHTTPSAddress is the default HTTPS listen address for NSQD
	NSQDHTTPSAddress = "127.0.0.1:4152"
	// StatePath is the default location that Sensu stores data for nsqd
	StatePath = "/var/lib/sensu/nsqd"
)

// Config specifies the messagebus configuration
type Config struct {
	TCPAddress   string
	HTTPAddress  string
	HTTPSAddress string
	StatePath    string
}

// NewConfig returns a sane configuration with defaults.
func NewConfig() *Config {
	c := &Config{
		StatePath:    StatePath,
		TCPAddress:   NSQDTCPAddress,
		HTTPAddress:  NSQDHTTPAddress,
		HTTPSAddress: NSQDHTTPSAddress,
	}

	return c
}

func ensureDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if mkdirErr := os.MkdirAll(path, 0700); mkdirErr != nil {
				return mkdirErr
			}
		} else {
			return err
		}
	}
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("path exists and is not directory - %s", path)
	}
	return nil
}

// NewNSQD returns an initialized NSQD.
func NewNSQD(config *Config) (*nsqd.NSQD, error) {
	if err := ensureDir(config.StatePath); err != nil {
		return nil, err
	}

	opts := nsqd.NewOptions()
	opts.TCPAddress = config.TCPAddress
	opts.HTTPAddress = config.HTTPAddress
	opts.HTTPSAddress = config.HTTPSAddress
	opts.Verbose = false
	opts.DataPath = config.StatePath

	return nsqd.New(opts), nil
}

// NsqBus is a Daemon and a MessageBus.
type NsqBus struct {
	NSQD *nsqd.NSQD

	producer  *nsq.Producer
	consumers []*consumerChannel
	errChan   chan error
}

type consumerChannel struct {
	consumer *nsq.Consumer
	channel  chan []byte
}

// Start starts the underlying NSQD, and returns an error upon failure.
func (bus *NsqBus) Start() error {
	bus.errChan = make(chan error, 1)
	bus.consumers = []*consumerChannel{}
	done := make(chan struct{})
	go func() {
		bus.NSQD.Main()
		close(done)
	}()
	<-done
	producer, err := nsq.NewProducer(bus.NSQD.RealTCPAddr().String(), nsq.NewConfig())
	if err != nil {
		return err
	}
	bus.producer = producer
	return nil
}

// Stop stops the underlying NSQD and closes any open consumers/producers.
func (bus *NsqBus) Stop() error {
	if bus.consumers != nil {
		wg := &sync.WaitGroup{}
		for _, conChan := range bus.consumers {
			wg.Add(1)
			go func(cch *consumerChannel) {
				cch.consumer.Stop()
				<-cch.consumer.StopChan
				close(cch.channel)
				wg.Done()
			}(conChan)
		}
	}
	return nil
}

// Status returns the current error from NSQD.
func (bus *NsqBus) Status() error {
	// I think we can just return bus.NSQD.GetError() here, but I'm not entirely
	// sure, so I'm just leaving it like this for now.
	if bus.NSQD.GetHealth() != "ok" {
		return bus.NSQD.GetError()
	}
	return nil
}

// Err returns a channel to listen for any terminal errors from the NSQD
// monitor.
func (bus *NsqBus) Err() <-chan error {
	return bus.errChan
}

// Subscribe sets up a managed consumer. Upon terminal failure of the
// underlying NSQD or shutdown, the returned channel will be closed.
func (bus *NsqBus) Subscribe(topic, channel string) (<-chan []byte, error) {
	consumer, err := nsq.NewConsumer(topic, channel, nsq.NewConfig())
	if err != nil {
		return nil, err
	}
	addr := bus.NSQD.RealTCPAddr()
	msgChannel := make(chan []byte, 100)
	consumer.AddHandler(nsq.HandlerFunc(func(nsqMessage *nsq.Message) error {
		select {
		case msgChannel <- nsqMessage.Body:
			return nil
		default:
			return errors.New("cannot handle message at this time")
		}
	}))

	err = consumer.ConnectToNSQD(addr.String())
	if err != nil {
		return nil, err
	}

	cch := &consumerChannel{
		consumer: consumer,
		channel:  msgChannel,
	}
	bus.consumers = append(bus.consumers, cch)
	return msgChannel, nil
}

// Publish sends a message over a given topic. If the message failes to send
// for any reason, an error will be returned.
func (bus *NsqBus) Publish(topic string, msg []byte) error {
	return bus.producer.Publish(topic, msg)
}
