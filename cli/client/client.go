package client

import (
	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/client/config"
	resty "gopkg.in/resty.v0"
)

var logger *logrus.Entry

// RestClient wraps resty.Client
type RestClient struct {
	client *resty.Client
	config config.Config

	configured bool
}

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})
}

// New builds a new client with defaults
func New(config config.Config) *RestClient {
	c := &RestClient{client: resty.New(), config: config}
	c.client.SetHeader("Accept", "application/json")
	c.client.SetHeader("Content-Type", "application/json")

	// logging
	w := logger.Writer()
	defer w.Close()
	c.client.SetLogger(w)

	return c
}

func (c *RestClient) configure() {
	if c.configured {
		return
	}

	c.client.SetHostURL(c.config.GetString("api-url"))

	// Token authentication (Not implemented yet)
	//c.client.SetAuthToken(c.config.GetString("secret"))

	// Password authentication
	username := c.config.GetString("userid")
	password := c.config.GetString("secret")
	if username != "" && password != "" {
		c.client.SetBasicAuth(username, password)
	}

	c.configured = true
}

// R returns new resty.Request from configured client
func (c *RestClient) R() *resty.Request {
	c.configure()
	r := c.client.R()
	return r
}
