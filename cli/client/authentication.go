package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
	jwt "github.com/golang-jwt/jwt/v4"
	corev2 "github.com/sensu/core/v2"
)

// CreateAccessToken returns a new access token given userid and password
func (client *RestClient) CreateAccessToken(url, userid, password string) (*corev2.Tokens, error) {
	// Make sure any existing auth token doesn't get injected instead
	client.ClearAuthToken()
	defer client.Reset()

	// Execute
	res, err := client.R().SetBasicAuth(userid, password).Get(url + "/auth")
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, errors.New(string(res.Body()))
	}

	tokens := &corev2.Tokens{}
	if err = json.Unmarshal(res.Body(), tokens); err != nil {
		return nil, fmt.Errorf("could not unmarshal response from server: %s", err)
	}

	return tokens, err
}

// TestCreds checks if the provided User credentials are valid
func (client *RestClient) TestCreds(userid, password string) error {
	client.ClearAuthToken()

	res, err := client.R().SetBasicAuth(userid, password).Get("/auth/test")
	if err != nil {
		return err
	}

	if res.StatusCode() == 401 {
		return errors.New(string(res.Body()))
	} else if res.StatusCode() >= 400 {
		//lint:ignore ST1005 this error is written to stdout/stderr
		return errors.New("Received an unexpected response from the API")
	}

	return nil
}

// Logout performs a logout of the configured user
func (client *RestClient) Logout(token string) error {
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"refresh_token": token}).
		Post("/auth/logout")
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		//lint:ignore ST1005 this error is written to stdout/stderr
		return fmt.Errorf("The server returned the error: %d %s",
			res.StatusCode(),
			res.String(),
		)
	}

	return nil
}

// RefreshAccessToken returns a new access token given valid refresh token
func (client *RestClient) RefreshAccessToken(tokens *corev2.Tokens) (*corev2.Tokens, error) {
	var err error
	var res *resty.Response
	url := "/auth/token"

	// Parse the access claims so we can determine the issuer URL
	parser := new(jwt.Parser)
	claims := &corev2.Claims{}
	_, _, _ = parser.ParseUnverified(tokens.Access, claims)

	request := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"refresh_token": tokens.Refresh}).
		SetResult(&corev2.Tokens{})

	switch issuer := claims.Issuer; {
	case issuer != "":
		// Attempt to renew the access token with the issuer
		res, err = request.Post(claims.Issuer + url)

		// If we received a successful response, break of this switch and directly
		// decode the response body
		if err == nil && (res != nil && res.StatusCode() < 400) {
			break
		}

		// If the issuer could not renew the access token (could be because it's no
		// longer available etc.), try to renew the access token with the configured
		// API URL instead
		fallthrough
	default:
		res, err = request.Post(url)
	}

	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		//lint:ignore ST1005 this error is written to stdout/stderr
		return nil, fmt.Errorf("The server returned the error: %d %s",
			res.StatusCode(),
			res.String(),
		)
	}

	tokens, ok := res.Result().(*corev2.Tokens)
	if !ok {
		//lint:ignore ST1005 this error is written to stdout/stderr
		return nil, fmt.Errorf("Unable to unmarshal response from server")
	}

	err = tokens.Validate()
	if err != nil {
		return nil, err
	}

	return tokens, err
}
