// Package system provides information about the system of the current
// process. System information is use for Agent (and potentially
// Backend) entity context.
package system

import (
        "github.com/shirou/gopsutil/host"
        "github.com/sensu/sensu-go/types"
)

// Info describes the system of the current process.
func Info() (types.System, error) {
        i, err := host.Info()

	if err != nil {
		return types.System{}, err
	}

        s := types.System{
                Hostname: i.Hostname,
                OS: i.OS,
                Platform: i.Platform,
                PlatformFamily: i.PlatformFamily,
                PlatformVersion: i.PlatformVersion,
        }

	return s, nil
}
