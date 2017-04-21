package client

import resty "gopkg.in/resty.v0"

var defaultClient *resty.Client

// RestClient wraps resty.Client
type RestClient struct {
	client *resty.Client
}

// New builds a new client with defaults
func New() *RestClient {
	c := &RestClient{client: resty.New()}
	c.setLogger(Logger)
	c.setHeader("Accept", "application/json")
	c.setHeader("Content-Type", "application/json")
	// c.setAccessToken(...

	return c
}

// Request returns new resty.Request from default client
func Request() resty.Request {
	return defaultClient.R()
}

func init() {
	defaultClient = New()
}
