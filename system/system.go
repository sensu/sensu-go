// Package system provides information about the system of the current
// process. System information is used for Agent (and potentially
// Backend) Entity context.
package system

import (
	"runtime"

	"github.com/sensu/sensu-go/types"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/net"
)

// Info describes the local system, hostname, OS, platform, platform
// family, platform version, and network interfaces.
func Info() (types.System, error) {
	info, err := host.Info()

	if err != nil {
		return types.System{}, err
	}

	system := types.System{
		Arch:            runtime.GOARCH,
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
	}

	network, err := NetworkInfo()

	if err == nil {
		system.Network = network
	}

	return system, nil
}

// NetworkInfo describes the local network interfaces, including their
// names (e.g. eth0), MACs (if available), and addresses.
func NetworkInfo() (types.Network, error) {
	interfaces, err := net.Interfaces()

	network := types.Network{}

	if err != nil {
		return network, err
	}

	for _, i := range interfaces {
		nInterface := types.NetworkInterface{
			Name: i.Name,
			MAC:  i.HardwareAddr,
		}

		for _, address := range i.Addrs {
			nInterface.Addresses = append(nInterface.Addresses, address.Addr)
		}

		network.Interfaces = append(network.Interfaces, nInterface)
	}

	return network, nil
}
