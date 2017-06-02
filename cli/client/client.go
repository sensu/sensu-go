package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/client/config"
	resty "gopkg.in/resty.v0"
)

var logger *logrus.Entry

// RestClient wraps resty.Client
type RestClient struct {
	resty  *resty.Client
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
	c := &RestClient{resty: resty.New(), config: config}
	c.resty.SetHeader("Accept", "application/json")
	c.resty.SetHeader("Content-Type", "application/json")

	// logging
	w := logger.Writer()
	defer w.Close()
	c.resty.SetLogger(w)

	return c
}

// R returns new resty.Request from configured client
func (client *RestClient) R() *resty.Request {
	client.configure()
	request := client.resty.R()

	return request
}

// Reset client so that it reconfigure on next request
func (client *RestClient) Reset() {
	client.configured = false
}

// ClearAuthToken clears the authoization token from the client config
func (client *RestClient) ClearAuthToken() {
	client.configure()
	client.resty.SetAuthToken("")
}

func (client *RestClient) configure() {
	if client.configured {
		return
	}

	restyInst := client.resty
	config := client.config

	// Set URL & access token
	restyInst.SetHostURL(config.GetString("api-url"))
	restyInst.SetAuthToken(config.GetString("secret"))

	// Check that Access-Token has not expired
	restyInst.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		// Guard against requests that are not sending auth details
		if c.Token == "" || r.UserInfo != nil {
			return nil
		}

		expiry := config.GetTime("expires-at")
		refreshToken := config.GetString("refresh-token")

		// No-op if token has not yet expired
		if hasExpired := expiry.Before(time.Now()); !hasExpired {
			return nil
		}

		if refreshToken == "" {
			return errors.New("configured access token has expired")
		}

		// TODO: Move this into it's own file / package
		// Request a new access token from the server
		newAccessToken, err := client.RefreshAccessToken(refreshToken)
		if err != nil {
			return fmt.Errorf(
				"failed to request new refresh token; client returned '%s'",
				err,
			)
		}

		// Write new tokens to disk
		err = config.WriteCredentials(newAccessToken)
		if err != nil {
			return fmt.Errorf(
				"failed to update configuration with new refresh token (%s)",
				err,
			)
		}

		c.SetAuthToken(newAccessToken.Token)

		return nil
	})

	client.configured = true
}
