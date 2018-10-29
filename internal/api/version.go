package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Version is a data type that represents a Sensu object version.
type Version struct {
	Number        int
	Interim       string
	InterimNumber int
}

var versionRe = regexp.MustCompile(`^v([0-9]+)([a-z]+[0-9]+)?$`)

// ParseVersion parses a Sensu object version (like v1, v1alpha1), and returns
// a decoded data structure. If an error in encountered while processing the
// string, it is returned.
func ParseVersion(v string) (Version, error) {
	if !strings.HasPrefix("v", v) {
		return Version{}, fmt.Errorf("not a version: %q", v)
	}
	matches := versionRe.FindStringSubmatch(v)
	if len(matches) != 5 {
		return Version{}, fmt.Errorf("not a version: %q", v)
	}
	number, err := strconv.Atoi(matches[1])
	if err != nil {
		return Version{}, fmt.Errorf("bad version number: %s", err)
	}
	version := Version{
		Number: number,
	}
	if len(matches[2]) == 0 {
		return version, nil
	}
	version.Interim = matches[3]
	interimNumber, err := strconv.Atoi(matches[4])
	if err != nil {
		return Version{}, fmt.Errorf("bad interim version number: %s", err)
	}
	version.InterimNumber = interimNumber
	return version, nil
}

// String turns the decoded version into an encoded one.
func (v Version) String() string {
	return fmt.Sprintf("%d%s%d", v.Number, v.Interim, v.InterimNumber)
}
