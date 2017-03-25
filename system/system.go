// Package system provides information about the system of the current
// process. System information is use for Agent (and potentially
// Backend) entity context.
package system

import (
        "github.com/shirou/gopsutil/host"
        "github.com/shirou/gopsutil/net"
        "github.com/sensu/sensu-go/types"
)

// Info describes the local system.
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

        n, err := NetworkInfo()

        if err == nil {
		s.Network = n
        }

        return s, nil
}


// NetworkInfo describes the local network interfaces.
func NetworkInfo() (types.Network, error) {
        i, err := net.Interfaces()

        n := types.Network{}

        if err != nil {
                return n, err
        }

        for _, ni := range i {
		item := types.NetworkInterface{
			Name: ni.Name,
		}

		n.Interfaces = append(n.Interfaces, item)
        }

        return n, nil
}
