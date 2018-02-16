package url

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	testCases := []struct {
		name        string
		backendURL  string
		port        string
		expectedURL string
	}{
		{
			name:        "URL without port",
			backendURL:  "ws://127.0.0.1",
			port:        "8081",
			expectedURL: "ws://127.0.0.1:8081",
		},
		{
			name:        "URL with port",
			backendURL:  "ws://127.0.0.1:8081",
			port:        "8081",
			expectedURL: "ws://127.0.0.1:8081",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			newURL, err := AppendPortIfMissing(tc.backendURL, tc.port)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, tc.expectedURL, newURL)
		})
	}
}
