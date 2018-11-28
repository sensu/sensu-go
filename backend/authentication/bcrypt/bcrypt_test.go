package bcrypt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPassword(t *testing.T) {
	hash := "$2a$10$iyYyGmveS9dcYp5DHMbOm.LShX806vB0ClzoPyt1TIgkZ9KQ62cOO"
	password := "P@ssw0rd!"

	assert.False(t, CheckPassword(hash, "foo"))
	assert.True(t, CheckPassword(hash, password))
}

func TestHashPassword(t *testing.T) {
	password := "P@ssw0rd!"

	hash, err := HashPassword(password)
	assert.NotEqual(t, password, hash)
	assert.NoError(t, err)
}
