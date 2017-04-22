package commands

import (
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/client"
)

// SensuCli is an instance of the Sensu command line client;
// encapsulates API client, logger & general configuration.
type SensuCli struct {
	config map[string]string
	client *client.RestClient
	logger *logrus.Entry
	stdOut *OutStream
	stdErr io.Writer
}
