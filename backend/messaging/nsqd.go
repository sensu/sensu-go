package messaging

import "github.com/nsqio/nsq/nsqd"

const (
	// NSQDTCPAddress is the default TCP listen address for NSQD
	NSQDTCPAddress = "127.0.0.1:4150"
	// NSQDHTTPAddress is the default HTTP listen address for NSQD
	NSQDHTTPAddress = "127.0.0.1:4151"
	// NSQDHTTPSAddress is the default HTTPS listen address for NSQD
	NSQDHTTPSAddress = "127.0.0.1:4152"
)

// NewNSQD returns an initialized NSQD.
func NewNSQD() *nsqd.NSQD {
	opts := nsqd.NewOptions()
	opts.TCPAddress = NSQDTCPAddress
	opts.HTTPAddress = NSQDHTTPAddress
	opts.HTTPSAddress = NSQDHTTPSAddress
	opts.Verbose = false

	nsqd := nsqd.New(opts)

	return nsqd
}
