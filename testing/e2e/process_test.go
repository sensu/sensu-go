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
	"testing"

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

	HTTPURL string
	WSURL   string

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

// newBackend abstracts the initialization of a backend process and returns a
// ready-to-use backend or exit with a fatal error if an error occurred while
// initializing it
func newBackend(t *testing.T) (*backendProcess, func()) {
	backend, cleanup := newBackendProcess(t)

	if err := backend.Start(); err != nil {
		cleanup()
		t.Fatal(err)
	}

	// Set the HTTP & WS URLs
	backend.HTTPURL = fmt.Sprintf("http://127.0.0.1:%d", backend.APIPort)
	backend.WSURL = fmt.Sprintf("ws://127.0.0.1:%d/", backend.AgentPort)

	// Make sure the backend is ready
	isOnline := waitForBackend(backend.HTTPURL)
	if !isOnline {
		cleanup()
		t.Fatal("the backend never became ready in a timely fashion")
	}

	return backend, func() {
		cleanup()
		if err := backend.Kill(); err != nil {
			t.Fatal(err)
		}
	}
}

// newBackendProcess initializes a backendProcess struct
func newBackendProcess(t *testing.T) (*backendProcess, func()) {
	ports := make([]int, 5)
	err := testutil.RandomPorts(ports)
	if err != nil {
		t.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		t.Fatal(err)
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

	return bep, func() { _ = os.RemoveAll(tmpDir) }
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
	agentConfig

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

type agentConfig struct {
	ID          string
	BackendURLs []string
	APIPort     int
	SocketPort  int
}

// newAgent abstracts the initialization of an agent process and returns a
// ready-to-use agent or exit with a fatal error if an error occurred while
// initializing it
func newAgent(config agentConfig, sensuctl *sensuCtl, t *testing.T) (*agentProcess, func()) {
	agent := &agentProcess{agentConfig: config}

	// Start the agent
	if err := agent.Start(t); err != nil {
		t.Fatal(err)
	}

	// Wait for the agent to send its first keepalive so we are sure it's
	// connected to the backend
	if ready := waitForAgent(agent.ID, sensuctl); !ready {
		t.Fatal("the backend never received a keepalive from the agent")
	}

	return agent, func() {
		if err := agent.Kill(); err != nil {
			t.Fatal(err)
		}
	}
}

func (a *agentProcess) Start(t *testing.T) error {
	port := make([]int, 2)
	err := testutil.RandomPorts(port)
	if err != nil {
		t.Fatal(err)
	}
	a.APIPort = port[0]
	a.SocketPort = port[1]

	args := []string{
		"start",
		"--id", a.ID,
		"--subscriptions", "test",
		"--environment", "default",
		"--organization", "default",
		"--api-port", strconv.Itoa(port[0]),
		"--socket-port", strconv.Itoa(port[1]),
		"--keepalive-interval", "1",
	}

	// Support a single or multiple backend URLs
	// backendURLs := []string{}
	for _, url := range a.BackendURLs {
		args = append(args, "--backend-url")
		args = append(args, url)
	}

	cmd := exec.Command(agentPath, args...)
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
	stdin     io.Reader
}

// newSensuCtl initializes a sensuctl
func newSensuCtl(apiURL, org, env, user, pass string) (*sensuCtl, func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensuctl")
	if err != nil {
		log.Panic(err)
	}

	ctl := &sensuCtl{
		ConfigDir: tmpDir,
		stdin:     os.Stdin,
	}

	// Authenticate sensuctl
	_, err = ctl.run("configure",
		"-n",
		"--url", apiURL,
		"--username", user,
		"--password", pass,
		"--format", "json",
		"--organization", org,
		"--environment", env,
	)
	if err != nil {
		log.Panic(err)
	}

	return ctl, func() { _ = os.RemoveAll(tmpDir) }
}

// run executes the sensuctl binary with the provided arguments
func (s *sensuCtl) run(args ...string) ([]byte, error) {
	// Make sure we point to our temporary config directory
	args = append([]string{"--config-dir", s.ConfigDir}, args...)

	cmd := exec.Command(sensuctlPath, args...)
	cmd.Stdin = s.stdin
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, err
	}

	return out, nil
}

func (s *sensuCtl) SetStdin(r io.Reader) {
	s.stdin = r
}
