package client

// FetchLicense fetches the installed license. This is a temporary workaround
// until https://github.com/sensu/sensu-go/issues/1870 is implemented
func (client *RestClient) FetchLicense() (interface{}, error) {
	return nil, ErrNotImplemented
}

// UpdateLicense updates the installed license with the given one. This is a
// temporary workaround until https://github.com/sensu/sensu-go/issues/1870 is
// implemented
func (client *RestClient) UpdateLicense(license interface{}) error {
	return ErrNotImplemented
}
