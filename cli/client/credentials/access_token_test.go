package credentials

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccessTokenUnmarshalJSON(t *testing.T) {
	assert := assert.New(t)

	token := &AccessToken{}
	err := token.UnmarshalJSON([]byte(`{
		"access_token": "abc",
		"refresh_token": "123",
		"expires_at": 123
	}`))

	assert.NoError(err, "Unmarshals without error")
	assert.Equal(token.Token, "abc")
	assert.Equal(token.RefreshToken, "123")
	assert.Equal(token.ExpiresAt, time.Unix(123, 0))

	err = token.UnmarshalJSON([]byte(`{
		"access_token": 123,
		"refresh_token": "123",
		"expires_at": 123
	}`))
	assert.Error(err, "throws error when key does not match expectations")

	err = token.UnmarshalJSON([]byte(`{
		"access_token": "123",
		"refresh_token": {},
		"expires_at": 123
	}`))
	assert.Error(err, "throws error when key does not match expectations")

	err = token.UnmarshalJSON([]byte(`{
		"access_token": "123",
		"refresh_token": "123",
		"expires_at": []
	}`))
	assert.Error(err, "throws error when key does not match expectations")

	err = token.UnmarshalJSON([]byte(`[123]`))
	assert.Error(err, "throws error when data does not match expectations")
}
