package url

import (
	"net"
	"net/url"
	"strings"
)

// AppendPortIfMissing takes a URL and port as strings and appends the provided
// port to the URL if it is missing
func AppendPortIfMissing(u string, p string) (string, error) {
	parsedURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}

	port := parsedURL.Port()
	if port == "" {
		port = p
	}

	host := parsedURL.Hostname()
	if IsIPv6(host) {
		host = "[" + host + "]"
	}

	parsedURL.Host = host + ":" + port
	return parsedURL.String(), nil
}

// IsIPv6 takes a string containing an IP address and will a boolean value
// based on whether or not the IP is IPv6
func IsIPv6(str string) bool {
	ip := net.ParseIP(str)
	return ip != nil && strings.Contains(str, ":")
}
