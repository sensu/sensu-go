package url

import "net/url"

// AppendPortIfMissing takes a URL and port as strings and appends the provided
// port to the URL if it is missing
func AppendPortIfMissing(u string, p string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	if parsedURL.Port() == "" {
		return parsedURL.String() + ":" + p, nil
	}
	return parsedURL.String(), nil
}
