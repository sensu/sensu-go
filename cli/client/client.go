package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/version"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

// ErrNotImplemented is returned by client methods that haven't been
// implemented in Sensu Core.
var ErrNotImplemented = errors.New("method not implemented")

// RestClient wraps resty.Client
type RestClient struct {
	resty  *resty.Client
	config config.Config

	configured   bool
	expiredToken bool
}

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "cli-client",
	})
}

// New builds a new client with defaults
func New(config config.Config) *RestClient {
	restyInst := resty.New()
	client := &RestClient{resty: restyInst, config: config}

	// set http client timeout
	restyInst.SetTimeout(config.Timeout())

	// Standardize redirect policy
	restyInst.SetRedirectPolicy(resty.FlexibleRedirectPolicy(10))

	// JSON
	restyInst.SetHeader("Accept", "application/json")
	restyInst.SetHeader("Content-Type", "application/json")

	// Check that Access-Token has not expired
	restyInst.OnBeforeRequest(func(c *resty.Client, r *resty.Request) error {
		c.SetHeader("User-Agent", "sensuctl/"+version.Semver())

		// Guard against requests that are not sending auth details
		if c.Token == "" || r.UserInfo != nil {
			return nil
		}

		// If the client access token is expired, it means this request is trying to
		// retrieve a new access token and therefore we do not need to do it again
		// otherwise we will have an infinite loop!
		if client.expiredToken {
			return nil
		}

		tokens := config.Tokens()
		expiry := time.Unix(tokens.ExpiresAt, 0)

		// No-op if token has not yet expired
		if hasExpired := expiry.Before(time.Now()); !hasExpired {
			return nil
		}

		if tokens.Refresh == "" {
			return errors.New("configured access token has expired")
		}

		// Mark the token as expired to prevent an infinite loop in this method
		client.expiredToken = true

		// TODO: Move this into it's own file / package
		// Request a new access token from the server
		tokens, err := client.RefreshAccessToken(tokens)
		if err != nil {
			return fmt.Errorf(
				"failed to request new refresh token; client returned '%s'",
				err,
			)
		}

		// Write new tokens to disk
		err = config.SaveTokens(tokens)
		if err != nil {
			return fmt.Errorf(
				"failed to update configuration with new refresh token (%s)",
				err,
			)
		}

		// We can now mark the token as valid
		client.expiredToken = false

		c.SetAuthToken(tokens.Access)

		return nil
	})

	restyInst.SetLogger(logger)

	return client
}

// R returns new resty.Request from configured client
func (client *RestClient) R() *resty.Request {
	client.configure()
	request := client.resty.R()

	return request
}

// SetTLSClientConfig assigns client TLS config
func (client *RestClient) SetTLSClientConfig(c *tls.Config) {
	client.resty.SetTLSClientConfig(c)
}

// Reset client so that it reconfigure on next request
func (client *RestClient) Reset() {
	client.configured = false
}

// ClearAuthToken clears the authorization token from the client config
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
	restyInst.SetHostURL(config.APIUrl())

	tokens := config.Tokens()
	if tokens != nil && tokens.Access != "" {
		restyInst.SetAuthToken(tokens.Access)
	}

	client.configured = true
}

// ApplyListOptions mutates the given request to make it carry the semantics of
// the given options.
func ApplyListOptions(request *resty.Request, options *ListOptions) {
	if options.FieldSelector != "" {
		request.SetQueryParam("fieldSelector", options.FieldSelector)
	}

	if options.LabelSelector != "" {
		request.SetQueryParam("labelSelector", options.LabelSelector)
	}

	if options.ChunkSize > 0 {
		request.SetQueryParam("limit", strconv.Itoa(options.ChunkSize))
	}

	if options.ContinueToken != "" {
		request.SetQueryParam("continue", options.ContinueToken)
	}
}
