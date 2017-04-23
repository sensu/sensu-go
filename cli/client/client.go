package client

import (
	resty "gopkg.in/resty.v0"
)

// RestClient wraps resty.Client
type RestClient struct {
	client *resty.Client
	config *Config

	configured bool
}

// New builds a new client with defaults
func New(config *Config) *RestClient {
	c := &RestClient{client: resty.New(), config: config}
	c.client.SetLogger(&Logger{})
	c.client.SetHeader("Accept", "application/json")
	c.client.SetHeader("Content-Type", "application/json")

	return c
}

func (c *RestClient) configure() {
	if c.configured {
		return
	}

	c.client.SetHostURL(c.config.GetString("url"))
	c.client.SetAuthToken(c.config.GetString("secret"))
	c.configured = true
}

// Request returns new resty.Request from default client
func (c *RestClient) R() *resty.Request {
	c.configure()
	r := c.client.R()
	return r
}
