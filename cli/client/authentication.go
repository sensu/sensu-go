package client

import (
	"encoding/json"
	"errors"
	"fmt"

	creds "github.com/sensu/sensu-go/cli/client/credentials"
)

// CreateAccessToken returns a new access token given userid and password
func (client *RestClient) CreateAccessToken(userid string, password string) (*creds.AccessToken, error) {
	res, err := client.R().SetBasicAuth(userid, password).Get("/auth")
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
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
	bytes, err := json.Marshal(map[string]string{"refresh_token": token})
	if err != nil {
		return nil, err
	}

	res, err := client.R().SetBody(bytes).Post("/auth/token")
	if err != nil {
		return nil, err
	}

	var accessToken *creds.AccessToken
	if err = json.Unmarshal(res.Body(), accessToken); err != nil {
		return nil, fmt.Errorf("Unable to unmarshal response from server. %s", err)
	}

	return accessToken, err
}
