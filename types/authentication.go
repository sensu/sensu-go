package types

import (
	"errors"
	"time"
)

// Tokens contains the structure for exchanging tokens with the API
type Tokens struct {
	Access    string `json:"access_token"`
	ExpiresAt int64  `json:"expires_at"`
	Refresh   string `json:"refresh_token"`
}

// Validate returns an error if the tokens contain invalid values.
func (t *Tokens) Validate() error {
	if t.Access == "" {
		return errors.New("Access token cannot be empty")
	}

	if t.ExpiresAt == 0 {
		return errors.New("Expiration must be set")
	}

	if t.Refresh == "" {
		return errors.New("Refresh token cannot be empty")
	}

	return nil
}

// FixtureTokens given an access and refresh tokens returns valid tokens for use
// in tests
func FixtureTokens(accessToken, refreshToken string) *Tokens {
	return &Tokens{
		Access:    accessToken,
		ExpiresAt: time.Now().Unix(),
		Refresh:   refreshToken,
	}
}
