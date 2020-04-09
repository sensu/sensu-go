// Package system provides information about the system of the current
// process. System information is used for Agent (and potentially
// Backend) Entity context.
package system

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"

	_ "unsafe"

	"github.com/sensu/sensu-go/types"
	"github.com/shirou/gopsutil/host"
	shirounet "github.com/shirou/gopsutil/net"
)

const defaultHostname = "unidentified-hostname"

//go:linkname goarm runtime.goarm
var goarm int32
var gomips string

type azureMetadata struct {
	Error          string   `json:"error"`
	NewestVersions []string `json:"newest-versions"`
}

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
		FloatType:       gomips,
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

	return system, nil
}

func getLibCType() (string, error) {
	output, err := exec.Command("ldd", "--version").CombinedOutput()
	// The command above will return an exit code of 1 on alpine, but still output
	// the relevant information. Therefore, as a workaround, we can inspect stderr
	// and ignore the error if it contains pertinent data about the C library
	if err != nil && !strings.Contains(strings.ToLower(string(output)), "libc") {
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

// GetCloudProvider inspects local files, looks up hostnames, and makes HTTP
// requests in order to determine if the local system is running within a cloud
// provider.
func GetCloudProvider(ctx context.Context) string {
	switch runtime.GOOS {
	case "linux":
		// Issue a dmidecode command to see if we are on EC2 or GCP
		logger.Debug("sudo -n dmidecode -s bios-version")
		output, err := exec.Command("sudo", "-n", "dmidecode", "-s", "bios-version").CombinedOutput()
		if err != nil {
			logger.WithError(err).Debug("couldn't run dmidecode")
			return cloudMetadataFallback(ctx)
		}
		text := strings.ToLower(string(output))
		if strings.Contains(text, "amazon") {
			logger.Debug("Running on EC2")
			return "EC2"
		}
		if strings.Contains(text, "google") {
			logger.Debug("Running on GCP")
			return "GCP"
		}
		// At this point, we need to issue a slightly different command to see
		// if we are on Azure
		logger.Debug("sudo -n dmidecode -s system-manufacturer")
		output, err = exec.Command("sudo", "-n", "dmidecode", "-s", "system-manufacturer").CombinedOutput()
		if err != nil {
			logger.WithError(err).Debug("couldn't run dmidecode")
			return cloudMetadataFallback(ctx)
		}
		text = strings.ToLower(string(output))
		if strings.Contains(text, "microsoft") {
			logger.Debug("Running on Azure")
			return "Azure"
		}
	case "windows":
		return cloudMetadataFallback(ctx)
	}
	return ""
}

func cloudMetadataFallback(outerCtx context.Context) string {
	// Older EC2 instances have this file available to unprivileged users
	// See https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/identify_ec2_instances.html
	logger.Debug("Reading /sys/hypervisor/uuid")
	b, err := ioutil.ReadFile("/sys/hypervisor/uuid")
	if err == nil && bytes.HasPrefix(b, []byte("ec2")) {
		logger.Debug("Running on EC2")
		return "EC2"
	} else {
		logger.WithError(err).Debug("couldn't read /sys/hypervisor/uuid")
	}
	ctx, cancel := context.WithCancel(outerCtx)
	defer cancel()
	wg := new(sync.WaitGroup)
	wg.Add(3)
	c1 := dialEC2(ctx, wg)
	c2 := dialGCP(ctx, wg)
	c3 := dialAzure(ctx, wg)
	go func() {
		defer cancel()
		wg.Wait()
	}()
	select {
	case found := <-c1:
		return found
	case found := <-c2:
		return found
	case found := <-c3:
		return found
	case <-ctx.Done():
		return ""
	}
}

func dialEC2(ctx context.Context, wg *sync.WaitGroup) <-chan string {
	c := make(chan string, 1)
	go func() {
		defer wg.Done()
		logger.Debug("GET http://169.254.169.254/latest/dynamic/instance-identity")
		req, err := http.NewRequestWithContext(
			ctx, "GET", "http://169.254.169.254/latest/dynamic/instance-identity", nil)
		if err != nil {
			// unlikely
			panic(err)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.WithError(err).Debug("request failed")
			return
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			logger.Debug("Running on EC2")
			c <- "EC2"
		}
	}()
	return c
}

func dialGCP(ctx context.Context, wg *sync.WaitGroup) <-chan string {
	c := make(chan string, 1)
	go func() {
		defer wg.Done()
		logger.Debug("Resolving metadata.google.internal")
		_, err := net.DefaultResolver.LookupHost(ctx, "metadata.google.internal")
		if err == nil {
			logger.Debug("Running on GCP")
			c <- "GCP"
		} else {
			logger.WithError(err).Debug("couldn't resolve metadata.google.internal")
		}
	}()
	return c
}

func dialAzure(ctx context.Context, wg *sync.WaitGroup) <-chan string {
	c := make(chan string, 1)
	go func() {
		defer wg.Done()
		logger.Debug("GET http://169.254.169.254/metadata/instance")
		req, err := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/metadata/instance", nil)
		if err != nil {
			// unlikely
			panic(err)
		}
		req.Header.Set("Metadata", "true")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.WithError(err).Debug("request failed")
			return
		}
		defer resp.Body.Close()
		var meta azureMetadata
		if err := json.NewDecoder(resp.Body).Decode(&meta); err == nil {
			if len(meta.NewestVersions) > 0 {
				logger.Debug("Running on Azure")
				c <- "Azure"
			}
		}
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
