package client

import (
	"encoding/json"
	"github.com/sensu/sensu-go/util/logging"
)

// LogLevelPath is the api path for loglevel.
var LogLevelPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "loglevel")

// SetLogLevel .
func (client *RestClient) SetLogLevel(logLevel logging.LogLevelRequest) error {
	logRequest := []logging.LogLevelRequest{logLevel}
	b, err := json.Marshal(logRequest)
	if err != nil {
		return err
	}

	path := LogLevelPath()
	res, err := client.R().SetBody(b).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() != 200 {
		return UnmarshalError(res)
	}
	return nil
}

// SetLogLevelAllModules .
func (client *RestClient) SetLogLevelAllModules(logLevel string) error {

	path := LogLevelPath("modules", logLevel)
	res, err := client.R().Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() != 200 {
		return UnmarshalError(res)
	}
	return nil
}

// GetLogLevel .
func (client *RestClient) GetLogLevel(logLevel string) error {
	path := LogLevelPath("modules", logLevel)
	res, err := client.R().Get(path)
	if err != nil {
		return err
	}

	if res.StatusCode() != 200 {
		return UnmarshalError(res)
	}
	return nil
}
