package hooks

import (
	"testing"

	clientMock "github.com/sensu/sensu-go/cli/client/testing"
	cmdTesting "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestConfigurationPresent(t *testing.T) {
	testAssert := assert.New(t)

	cmdNoCheck := &cobra.Command{
		Annotations: map[string]string{
			ConfigurationRequirement: ConfigurationNotRequired,
		},
	}

	cmdWithCheck := &cobra.Command{
		Annotations: map[string]string{},
	}

	validTokens := &corev2.Tokens{
		Access:    "accesstoken",
		ExpiresAt: 1617721168,
		Refresh:   "refreshtoken",
	}

	invalidTokens := &corev2.Tokens{
		Access:    "",
		ExpiresAt: 1617721168,
		Refresh:   "",
	}

	apiKey := "de48ef7e-db2a-4333-bedf-7549de9541f4"
	noApiKey := ""

	apiURL := "http://localhost:8080"

	tests := []struct {
		name          string
		cmd           *cobra.Command
		apiKey        string
		apiURL        string
		tokens        *corev2.Tokens
		errorExpected bool
	}{
		{
			name:          "NoCheck",
			cmd:           cmdNoCheck,
			apiKey:        noApiKey,
			apiURL:        "",
			tokens:        nil,
			errorExpected: false,
		}, {
			name:          "NoURL",
			cmd:           cmdWithCheck,
			apiKey:        apiKey,
			apiURL:        "",
			tokens:        validTokens,
			errorExpected: true,
		}, {
			name:          "ApiKeyNoTokens",
			cmd:           cmdWithCheck,
			apiKey:        apiKey,
			apiURL:        apiURL,
			tokens:        nil,
			errorExpected: false,
		}, {
			name:          "NoAPIKeyValidTokens",
			cmd:           cmdWithCheck,
			apiKey:        noApiKey,
			apiURL:        apiURL,
			tokens:        validTokens,
			errorExpected: false,
		}, {
			name:          "NoAPIKeyNoTokens",
			cmd:           cmdWithCheck,
			apiKey:        noApiKey,
			apiURL:        apiURL,
			tokens:        nil,
			errorExpected: true,
		}, {
			name:          "NoAPIKeyInvalidAccessToken",
			cmd:           cmdWithCheck,
			apiKey:        noApiKey,
			apiURL:        apiURL,
			tokens:        invalidTokens,
			errorExpected: true,
		}, {
			name:          "APIKeyInvalidAccessToken",
			cmd:           cmdWithCheck,
			apiKey:        apiKey,
			apiURL:        apiURL,
			tokens:        invalidTokens,
			errorExpected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockCli := cmdTesting.NewMockCLI()
			mockConfig := mockCli.Config.(*clientMock.MockConfig)
			mockConfig.On("APIKey").Return(test.apiKey)
			mockConfig.On("APIUrl").Return(test.apiURL)
			mockConfig.On("Tokens").Return(test.tokens)

			if test.errorExpected {
				testAssert.Error(ConfigurationPresent(test.cmd, mockCli))
			} else {
				testAssert.NoError(ConfigurationPresent(test.cmd, mockCli))
			}
		})
	}
}
