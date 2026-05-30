//go:build arm
// +build arm

package system

import (
	"runtime/debug"
	"strconv"
)

func getARMVersion() int32 {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return 0
	}
	for _, setting := range info.Settings {
		if setting.Key == "GOARM" {
			v, err := strconv.Atoi(setting.Value)
			if err != nil {
				return 0
			}
			return int32(v)
		}
	}
	return 0
}
