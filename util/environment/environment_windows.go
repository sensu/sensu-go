//go:build windows
// +build windows

package environment

import "strings"

// On Windows, environment variables are case-insensitive; to avoid conflicts we
// coerce all keys to UPPER CASE.
// https://docs.microsoft.com/en-us/dotnet/api/system.environment.getenvironmentvariable?view=netframework-4.7.2
func coerceKey(k string) string {
	return strings.ToUpper(k)
}
