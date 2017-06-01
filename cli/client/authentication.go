package client

import (
	"encoding/json"
	"errors"
	"time"
)

// AccessToken wraps user's authorization secret
type AccessToken struct {
	Token        string
	RefreshToken string
	ExpiresAt    time.Time
}

// UnmarshalJSON updates access token given payload
func (a *AccessToken) UnmarshalJSON(data []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(payload, unmarshalledPayload); err != nil {
		return nil, err
	}

	if a.Token, ok = payload["access_token"].(string); !ok {
		return errors.New("given 'access_token' appears invalid")
	}

	if a.RefreshToken, _ = payload["refresh_token"].(string); !ok {
		return errors.New("given 'refresh_token' appears invalid")
	}

	if expiresAt, ok = unmarshalledPayload["expires_at"].(int64); ok {
		a.ExpiresAt = time.Unix(expiresAt, 0)
	} else if !ok {
		return errors.New("given 'expires_at' value appears invalid")
	}
}

// CreateAccessToken returns a new access token given userid and password
func (client *RestClient) CreateAccessToken(userid string, password string) (AccessToken, error) {
	res, err := client.R().SetBasicAuth(userid, password).Get("/auth")
	if err != nil {
		return nil, err
	}

	var accessToken *AccessToken
	if err = json.Unmarshal(res.Body(), accessToken); err != nil {
		return fmt.Errorf("Unable to unmarshal response from server. %s", err)
	}

	return accessToken, err
}

// RefreshAccessToken returns a new access token given valid refresh token
func (client *RestClient) RefreshAccessToken(token string) (AccessToken, error) {
	bytes, err := json.Marshal(map[string]string{"refresh_token": token})
	if err != nil {
		return nil, err
	}

	res, err := client.R().SetBody(bytes).Post("/auth/token")
	if err != nil {
		return nil, err
	}

	var accessToken *AccessToken
	if err = json.Unmarshal(res.Body(), accessToken); err != nil {
		return fmt.Errorf("Unable to unmarshal response from server. %s", err)
	}

	return accessToken, err
}
