package credentials

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
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	var ok bool
	if a.Token, ok = payload["access_token"].(string); !ok {
		return errors.New("given value for 'access_token' appears invalid")
	}

	if a.RefreshToken, ok = payload["refresh_token"].(string); !ok {
		return errors.New("given value for 'refresh_token' appears invalid")
	}

	if expiresAt, ok := payload["expires_at"].(float64); ok {
		a.ExpiresAt = time.Unix(int64(expiresAt), 0)
	} else {
		return errors.New("given value for 'expires_at' appears invalid")
	}

	return nil
}
