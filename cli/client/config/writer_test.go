package config

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/sensu/sensu-go/cli/client/credentials"
	"github.com/stretchr/testify/suite"
)

type ConfigWriterSuite struct {
	suite.Suite
	configFile *os.File
	config     *MultiConfig
}

func (suite *ConfigWriterSuite) SetupTest() {
	suite.config = &MultiConfig{profile: newProfilesConfig()}
	suite.configFile, _ = ioutil.TempFile("", "writer-test")
	CredentialsFilePath = suite.configFile.Name()
}

func (suite *ConfigWriterSuite) AfterTest() {
	suite.configFile.Close()
	os.Remove(suite.configFile.Name())
}

func (suite *ConfigWriterSuite) TestWriteURL() {
	err := suite.config.WriteURL("http://localhost:3000")
	suite.NoError(err, "Writes URL w/o error")
}

func (suite *ConfigWriterSuite) TestWriteCredentials() {
	err := suite.config.WriteCredentials(&credentials.AccessToken{})
	suite.NoError(err, "Writes credentials w/o error")
}

func TestRunWriterSuite(t *testing.T) {
	suite.Run(t, new(ConfigWriterSuite))
}
