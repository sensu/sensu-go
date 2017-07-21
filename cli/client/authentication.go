package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// CreateAccessToken returns a new access token given userid and password
func (client *RestClient) CreateAccessToken(url, userid, password string) (*types.Tokens, error) {
	// Make sure any existing auth token doesn't get injected instead
	client.ClearAuthToken()
	defer client.Reset()

	// Execute
	res, err := client.R().SetBasicAuth(userid, password).Get(url + "/auth")
	if err != nil {
		return nil, err
	}

	if res.StatusCode() == 401 {
		return nil, errors.New(string(res.Body()))
	} else if res.StatusCode() >= 400 {
		// TODO: (JK) we may want to expose a bit more of the error here
		return nil, errors.New("Received an unexpected response from the API")
	}

	var tokens types.Tokens
	if err = json.Unmarshal(res.Body(), &tokens); err != nil {
		return nil, errors.New("Unable to unmarshal response from server")
	}

	return &tokens, err
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
		return fmt.Errorf("The server returned the error: %d %s",
			res.StatusCode(),
			res.String(),
		)
	}

	return nil
}

// RefreshAccessToken returns a new access token given valid refresh token
func (client *RestClient) RefreshAccessToken(token string) (*types.Tokens, error) {
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"refresh_token": token}).
		SetResult(&types.Tokens{}).
		Post("/auth/token")
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("The server returned the error: %d %s",
			res.StatusCode(),
			res.String(),
		)
	}

	tokens, ok := res.Result().(*types.Tokens)
	if !ok {
		return nil, fmt.Errorf("Unable to unmarshal response from server")
	}

	err = tokens.Validate()
	if err != nil {
		return nil, err
	}

	return tokens, err
}
