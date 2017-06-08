package client

import (
	"encoding/json"
	"errors"
	"fmt"

	creds "github.com/sensu/sensu-go/cli/client/credentials"
)

// CreateAccessToken returns a new access token given userid and password
func (client *RestClient) CreateAccessToken(url, userid, password string) (*creds.AccessToken, error) {
	// Make sure any existing auth token doesn't get injected instead
	client.ClearAuthToken()
	defer client.Reset()

	// Execute
	res, err := client.R().SetBasicAuth(userid, password).Get(url + "/auth")
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, errors.New("Bad username or password given")
	}

	var token creds.AccessToken
	if err = json.Unmarshal(res.Body(), &token); err != nil {
		return nil, errors.New("Unable to unmarshal response from server")
	}

	return &token, err
}

// RefreshAccessToken returns a new access token given valid refresh token
func (client *RestClient) RefreshAccessToken(token string) (*creds.AccessToken, error) {
	res, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{"refresh_token": token}).
		SetResult(&creds.AccessToken{}).
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

	accessToken, ok := res.Result().(*creds.AccessToken)
	if !ok {
		return nil, fmt.Errorf("Unable to unmarshal response from server")
	}

	return accessToken, err
}
