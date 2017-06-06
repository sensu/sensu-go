package bytes

import (
	"crypto/rand"
)

// Random securely generates random string of n size
func Random(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}
