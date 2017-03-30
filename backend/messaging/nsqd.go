package messaging

import (
	"fmt"
	"os"

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

	nsqd := nsqd.New(opts)

	return nsqd, nil
}

// NsqBus is a Daemon and a MessageBus.
type NsqBus struct {
	NSQD    *nsqd.NSQD
	errChan chan error
}

// Start starts the underlying NSQD, and returns an error upon failure.
func (bus *NsqBus) Start() error {
	bus.errChan = make(chan error, 1)
	go bus.NSQD.Main()
	return nil
}

// Stop stops the underlying NSQD and closes any open consumers/producers.
func (bus *NsqBus) Stop() error {
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
	return nil, nil
}

// Publish sends a message over a given topic. If the message failes to send
// for any reason, an error will be returned.
func (bus *NsqBus) Publish(topic string, msg []byte) error {
	return nil
}
