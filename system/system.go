// Package system provides information about the system of the current
// process. System information is used for Agent (and potentially
// Backend) Entity context.
package system

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"

	_ "unsafe"

	"github.com/sensu/sensu-go/types"
	"github.com/shirou/gopsutil/host"
	shirounet "github.com/shirou/gopsutil/net"
)

const defaultHostname = "unidentified-hostname"

//go:linkname goarm runtime.goarm
var goarm int32

// Info describes the local system, hostname, OS, platform, platform
// family, platform version, and network interfaces.
func Info() (types.System, error) {
	info, err := host.Info()

	if err != nil {
		return types.System{}, err
	}

	system := types.System{
		Arch:            runtime.GOARCH,
		ARMVersion:      goarm,
		Hostname:        info.Hostname,
		OS:              info.OS,
		Platform:        info.Platform,
		PlatformFamily:  info.PlatformFamily,
		PlatformVersion: info.PlatformVersion,
	}

	if system.Hostname == "" {
		system.Hostname = defaultHostname
	}

	network, err := NetworkInfo()

	if err == nil {
		system.Network = network
	}

	vmSystem, vmRole, err := host.Virtualization()
	if err == nil {
		system.VMSystem = vmSystem
		system.VMRole = vmRole
	}

	if runtime.GOOS == "linux" {
		libcType, err := getLibCType()
		if err == nil {
			system.LibCType = libcType
		}
	}

	system.CloudProvider = getCloudProvider()

	return system, nil
}

func getLibCType() (string, error) {
	output, err := exec.Command("ldd", "--version").CombinedOutput()
	if err != nil {
		return "", err
	}
	text := strings.ToLower(string(output))
	if strings.Contains(text, "gnu") {
		return "glibc", nil
	}
	if strings.Contains(text, "musl") {
		return "musl", nil
	}
	return "", nil
}

func getCloudProvider() string {
	switch runtime.GOOS {
	case "linux":
		// Issue a dmidecode command to see if we are on EC2 or GCP
		output, err := exec.Command("sudo", "-n", "dmidecode", "-s", "bios-version").CombinedOutput()
		if err != nil {
			return linuxCloudFallback()
		}
		text := strings.ToLower(string(output))
		if strings.Contains(text, "amazon") {
			return "EC2"
		}
		if strings.Contains(text, "google") {
			return "GCP"
		}
		// At this point, we need to issue a slightly different command to see
		// if we are on Azure
		output, err = exec.Command("sudo", "-n", "dmidecode", "-s", "system-manufacturer").CombinedOutput()
		if err != nil {
			return linuxCloudFallback()
		}
		text = strings.ToLower(string(output))
		if strings.Contains(text, "microsoft") {
			return "Azure"
		}
	case "windows":
		return ""
	}
	return ""
}

func linuxCloudFallback() string {
	// Older EC2 instances have this file available to unprivileged users
	// See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/identify_ec2_instances.html
	b, err := ioutil.ReadFile("/sys/hypervisor/uuid")
	if err == nil && bytes.HasPrefix(b, []byte("ec2")) {
		return "EC2"
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c1 := dialEC2(ctx)
	c2 := dialGCP(ctx)
	c3 := dialAzure(ctx)
	select {
	case found := <-c1:
		return found
	case found := <-c2:
		return found
	case found := <-c3:
		return found
	}

	return ""
}

func dialEC2(ctx context.Context) <-chan string {
	c := make(chan string, 1)
	go func() {
		req, err := http.NewRequestWithContext(
			ctx, "GET", "http://169.254.169.254/latest/dynamic/instance-identity", nil)
		if err != nil {
			// unlikely
			panic(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			c <- "EC2"
		}
	}()
	return c
}

func dialGCP(ctx context.Context) <-chan string {
	c := make(chan string, 1)
	go func() {
		_, err := net.DefaultResolver.LookupHost(ctx, "metadata.google.internal")
		if err == nil {
			c <- "GCP"
		}
	}()
	return c
}

func dialAzure(ctx context.Context) <-chan string {
	c := make(chan string, 1)
	go func() {

	}()
	return c
}

// NetworkInfo describes the local network interfaces, including their
// names (e.g. eth0), MACs (if available), and addresses.
func NetworkInfo() (types.Network, error) {
	interfaces, err := shirounet.Interfaces()

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
