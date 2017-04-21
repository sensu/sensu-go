package client

import resty "gopkg.in/resty.v0"

var defaultClient *RestClient

// RestClient wraps resty.Client
type RestClient struct {
	client *resty.Client
}

// New builds a new client with defaults
func New() *RestClient {
	c := &RestClient{client: resty.New()}
	c.client.SetHostURL("http://localhost:8080")
	c.client.SetLogger(&Logger{})
	c.client.SetHeader("Accept", "application/json")
	c.client.SetHeader("Content-Type", "application/json")
	// c.setAccessToken(...

	return c
}

// Request returns new resty.Request from default client
func Request() *resty.Request {
	return defaultClient.client.R()
}

func init() {
	defaultClient = New()
}
