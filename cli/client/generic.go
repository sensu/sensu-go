package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// PutResource ...
func (client *RestClient) PutResource(r types.Resource) error {
	path := r.URIPath()
	b, err := json.Marshal(r)
	if err != nil {
		return err
	}
	res, err := client.R().SetBody(b).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}
	if res.StatusCode() >= 400 {
		return fmt.Errorf("PUT %q: %s", path, res.String())
	}
	return nil
}
