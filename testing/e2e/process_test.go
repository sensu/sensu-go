package e2e

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/sensu/sensu-go/testing/testutil"
)

type backendProcess struct {
	AgentHost               string
	AgentPort               int
	APIHost                 string
	APIPort                 int
	DashboardHost           string
	DashboardPort           int
	StateDir                string
	EtcdPeerURL             string
	EtcdClientURL           string
	EtcdInitialCluster      string
	EtcdInitialClusterState string
	EtcdName                string
	EtcdInitialClusterToken string

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

// newBackendProcess initializes a backendProcess struct
func newBackendProcess() (*backendProcess, func()) {
	ports := make([]int, 5)
	err := testutil.RandomPorts(ports)
	if err != nil {
		log.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		log.Panic(err)
	}

	etcdClientURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	etcdPeerURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	apiPort := ports[2]
	agentPort := ports[3]
	dashboardPort := ports[4]
	initialCluster := fmt.Sprintf("default=%s", etcdPeerURL)

	bep := &backendProcess{
		AgentHost:               "127.0.0.1",
		AgentPort:               agentPort,
		APIHost:                 "127.0.0.1",
		APIPort:                 apiPort,
		DashboardHost:           "127.0.0.1",
		DashboardPort:           dashboardPort,
		StateDir:                tmpDir,
		EtcdClientURL:           etcdClientURL,
		EtcdPeerURL:             etcdPeerURL,
		EtcdInitialCluster:      initialCluster,
		EtcdInitialClusterState: "new",
		EtcdName:                "default",
	}

	return bep, func() { os.RemoveAll(tmpDir) }
}

// Start starts a backend process as configured. All exported variables in
// backendProcess must be configured.
func (b *backendProcess) Start() error {
	cmd := exec.Command(
		backendPath, "start",
		"-d", b.StateDir,
		"--agent-host", b.AgentHost,
		"--agent-port", strconv.FormatInt(int64(b.AgentPort), 10),
		"--api-host", b.APIHost,
		"--api-port", strconv.FormatInt(int64(b.APIPort), 10),
		"--dashboard-host", b.DashboardHost,
		"--dashboard-port", strconv.FormatInt(int64(b.DashboardPort), 10),
		"--store-client-url", b.EtcdClientURL,
		"--store-peer-url", b.EtcdPeerURL,
		"--store-initial-cluster", b.EtcdInitialCluster,
		"--store-initial-cluster-state", b.EtcdInitialClusterState,
		"--store-node-name", b.EtcdName,
		"--store-initial-advertise-peer-url", b.EtcdPeerURL,
		"--store-initial-cluster-token", b.EtcdInitialClusterToken,
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	b.Stdout = io.Reader(stdout)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	b.Stderr = io.Reader(stderr)

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Println(stdoutScanner.Text())
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}
	log.Printf("backend started with pid %d\n", cmd.Process.Pid)
	b.cmd = cmd
	return nil
}

func (b *backendProcess) Kill() error {
	if err := b.cmd.Process.Kill(); err != nil {
		return err
	}
	return b.cmd.Process.Release()
}

type agentProcess struct {
	BackendURLs []string
	AgentID     string
	APIPort     int

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

func (a *agentProcess) Start() error {
	port := make([]int, 1)
	err := testutil.RandomPorts(port)
	if err != nil {
		log.Fatal(err)
	}
	a.APIPort = port[0]

	cmd := exec.Command(
		agentPath, "start",
		"--backend-url", a.BackendURLs[0],
		"--backend-url", a.BackendURLs[1],
		"--id", a.AgentID,
		"--subscriptions", "test",
		"--environment", "default",
		"--organization", "default",
		"--api-port", strconv.Itoa(port[0]),
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	a.Stdout = stdout
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	a.Stderr = stderr

	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stdoutScanner.Scan() {
			fmt.Println(stdoutScanner.Text())
		}
	}()
	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()

	err = cmd.Start()
	if err != nil {
		return err
	}
	log.Printf("started agent with pid %d", cmd.Process.Pid)
	a.cmd = cmd
	return nil
}

func (a *agentProcess) Kill() error {
	if err := a.cmd.Process.Kill(); err != nil {
		return err
	}
	return a.cmd.Process.Release()
}

type sensuCtl struct {
	ConfigDir string
}

// newSensuCtl initializes a sensuctl
func newSensuCtl(apiURL, org, env, user, pass string) (*sensuCtl, func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensuctl")
	if err != nil {
		log.Panic(err)
	}

	ctl := &sensuCtl{
		ConfigDir: tmpDir,
	}

	// Authenticate sensuctl
	ctl.run("configure",
		"--url", apiURL,
		"--username", user,
		"--password", pass,
		"--format", "json",
		"--organization", org,
		"--environment", env,
	)

	return ctl, func() { os.RemoveAll(tmpDir) }
}

// run executes the sensuctl binary with the provided arguments
func (s *sensuCtl) run(args ...string) ([]byte, error) {
	// Make sure we point to our temporary config directory
	args = append([]string{"--config-dir", s.ConfigDir}, args...)

	out, err := exec.Command(sensuctlPath, args...).CombinedOutput()
	if err != nil {
		return out, err
	}

	return out, nil
}
