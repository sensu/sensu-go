package e2e

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/types"
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

	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func newDefaultBackend() (*backendProcess, func(), error) {
	return newBackendProcess(40000, 40001, 40002, 40003, 40004)
}

// newBackendProcess initializes a backendProcess struct
func newBackendProcess(etcdClientPort, etcdPeerPort, agentPort, apiPort, dashboardPort int) (*backendProcess, func(), error) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		return nil, nil, err
	}

	etcdClientURL := fmt.Sprintf("http://127.0.0.1:%d", etcdClientPort)
	etcdPeerURL := fmt.Sprintf("http://127.0.0.1:%d", etcdPeerPort)
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
		HTTPURL:                 fmt.Sprintf("http://127.0.0.1:%d", apiPort),
		WSURL:                   fmt.Sprintf("ws://127.0.0.1:%d", agentPort),
	}

	cleanup := func() {
		if err := bep.Terminate(); err != nil {
			log.Println(err)
		}
		_ = os.RemoveAll(tmpDir)
	}
	return bep, cleanup, nil
}

// Start starts a backend process as configured. All exported variables in
// backendProcess must be configured.
func (b *backendProcess) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel
	cmd := exec.CommandContext(
		ctx,
		backendPath, "start",
		"-d", b.StateDir,
		"--agent-host", b.AgentHost,
		"--agent-port", fmt.Sprintf("%d", b.AgentPort),
		"--api-host", b.APIHost,
		"--api-port", strconv.FormatInt(int64(b.APIPort), 10),
		"--dashboard-host", b.DashboardHost,
		"--dashboard-port", strconv.FormatInt(int64(b.DashboardPort), 10),
		"--listen-client-urls", b.EtcdClientURL,
		"--listen-peer-urls", b.EtcdPeerURL,
		"--initial-cluster", b.EtcdInitialCluster,
		"--initial-cluster-state", b.EtcdInitialClusterState,
		"--name", b.EtcdName,
		"--initial-advertise-peer-urls", b.EtcdPeerURL,
		"--initial-cluster-token", b.EtcdInitialClusterToken,
		"--log-level", "warn",
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

func (b *backendProcess) Terminate() error {
	return terminateProcess(b.cmd.Process)
}

type agentProcess struct {
	agentConfig
	backendURL string
	APIPort    int
	SocketPort int

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

type agentConfig struct {
	CustomAttributes  string
	ID                string
	Redact            []string
	KeepaliveTimeout  int
	KeepaliveInterval int
	Organization      string
	Environment       string
}

// newAgent abstracts the initialization of an agent process and returns a
// ready-to-use agent or exit with a fatal error if an error occurred while
// initializing it
func newAgent(config agentConfig, sensuctl *sensuCtl, t *testing.T) (*agentProcess, func()) {
	agent := &agentProcess{agentConfig: config, backendURL: sensuctl.wsURL}
	agent.APIPort = int(atomic.AddInt64(&agentPortCounter, 1))
	agent.SocketPort = int(atomic.AddInt64(&agentPortCounter, 1))
	if agent.Organization == "" {
		agent.Organization = sensuctl.Organization
	}
	if agent.Environment == "" {
		agent.Environment = sensuctl.Environment
	}

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
		if err := agent.Terminate(); err != nil {
			t.Fatal(err)
		}
	}
}

func (a *agentProcess) Start(t *testing.T) error {
	var interval string
	if a.agentConfig.KeepaliveInterval == 0 {
		interval = "1"
	} else {
		interval = strconv.Itoa(a.agentConfig.KeepaliveInterval)
	}

	var timeout string
	if a.agentConfig.KeepaliveTimeout == 0 {
		timeout = "10"
	} else {
		timeout = strconv.Itoa(a.agentConfig.KeepaliveTimeout)
	}

	args := []string{
		"start",
		"--id", a.ID,
		"--subscriptions", "test",
		"--environment", a.Environment,
		"--organization", a.Organization,
		"--api-port", strconv.Itoa(a.APIPort),
		"--socket-port", strconv.Itoa(a.SocketPort),
		"--keepalive-interval", interval,
		"--keepalive-timeout", timeout,
		"--backend-url", a.backendURL,
		"--statsd-disable",
	}

	// Support custom attributes
	if a.CustomAttributes != "" {
		args = append(args, "--custom-attributes")
		args = append(args, a.CustomAttributes)
	}

	// Support redact fields
	if len(a.Redact) != 0 {
		args = append(args, "--redact")
		args = append(args, strings.Join(a.Redact, ","))
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

func (a *agentProcess) Terminate() error {
	return terminateProcess(a.cmd.Process)
}

type sensuCtl struct {
	ConfigDir    string
	Organization string
	Environment  string
	wsURL        string
	httpURL      string
	stdin        io.Reader
}

func newCustomSensuctl(t *testing.T, wsURL, httpURL, org, env string) (*sensuCtl, func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensuctl")
	if err != nil {
		t.Fatal(err)
	}

	ctl := &sensuCtl{
		Organization: org,
		Environment:  env,
		ConfigDir:    tmpDir,
		stdin:        os.Stdin,
		wsURL:        wsURL,
		httpURL:      httpURL,
	}

	// Authenticate sensuctl
	out, err := ctl.run("configure",
		"-n",
		"--url", httpURL,
		"--username", "admin",
		"--password", "P@ssw0rd!",
		"--format", "json",
		"--organization", "default",
		"--environment", "default",
	)
	if err != nil {
		t.Fatal(err, string(out))
	}
	if org != "default" {
		out, err = ctl.run("organization", "create", org)
		if err != nil {
			t.Fatal(err, string(out))
		}
	}
	if env != "default" {
		b, err := json.Marshal(types.Wrapper{
			Value: &types.Environment{
				Name:         env,
				Organization: org,
				Description:  t.Name(),
			},
			Type: "Environment",
		})
		if err != nil {
			t.Fatal(err)
		}
		ctl.stdin = bytes.NewReader(b)
		// creates the environment. have to do it this way due to #1514
		out, err = ctl.run("create")
		if err != nil {
			t.Fatal(err, string(out))
		}
	}

	// Set default environment to newly created org and env
	_, err = ctl.run("configure",
		"-n",
		"--url", httpURL,
		"--username", "admin",
		"--password", "P@ssw0rd!",
		"--format", "json",
		"--organization", org,
		"--environment", env,
	)
	if err != nil {
		t.Fatal(err)
	}

	return ctl, func() { _ = os.RemoveAll(tmpDir) }
}

// newSensuCtl initializes a sensuctl
func newSensuCtl(t *testing.T) (*sensuCtl, func()) {
	org := uuid.New().String()
	env := uuid.New().String()

	return newCustomSensuctl(t, backend.WSURL, backend.HTTPURL, org, env)
}

// run executes the sensuctl binary with the provided arguments
func (s *sensuCtl) run(args ...string) ([]byte, error) {
	// Make sure we point to our temporary config directory
	args = append([]string{"--config-dir", s.ConfigDir}, args...)

	cmd := exec.Command(sensuctlPath, args...)
	cmd.Stdin = s.stdin
	out, err := cmd.CombinedOutput()

	return out, err
}

func (s *sensuCtl) SetStdin(r io.Reader) {
	s.stdin = r
}

// Terminate a Process by sending SIGTERM and waiting for it to exit cleanly.
// On Windows, there is no way to ask a program to exit cleanly. Instead, it
// must be killed with SIGKILL.
func terminateProcess(p *os.Process) error {
	signal := syscall.SIGTERM
	// Windows doesn't support SIGTERM, fall back to SIGKILL
	if runtime.GOOS == "windows" {
		signal = syscall.SIGKILL
	}

	if err := p.Signal(signal); err != nil {
		return err
	}
	// allow the process to exit cleanly
	_, err := p.Wait()
	return err
}

func waitForAgent(id string, sensuctl *sensuCtl) bool {
	fmt.Println("WAITING....", id, sensuctl.Organization, sensuctl.Environment)
	for i := 0; i < 5; i++ {
		_, err := sensuctl.run(
			"event", "info", id, "keepalive",
			"--organization", sensuctl.Organization,
			"--environment", sensuctl.Environment,
		)
		if err != nil {
			log.Println("keepalive not received, sleeping...")
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		log.Println("agent ready")
		return true
	}
	out, err := sensuctl.run("entity", "list")
	if err != nil {
		panic(err)
	}
	fmt.Println("ENTITY LIST", string(out))
	return false
}

func waitForBackend(url string) bool {
	for i := 0; i < 5; i++ {
		resp, getErr := http.Get(fmt.Sprintf("%s/health", url))
		if getErr != nil {
			log.Println("backend not ready, sleeping...")
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}
		_ = resp.Body.Close()

		if resp.StatusCode != 200 && resp.StatusCode != 401 {
			log.Printf("backend returned non-200/401 status code: %d\n", resp.StatusCode)
			time.Sleep(time.Duration(i+1) * time.Second)
			continue
		}

		log.Println("backend ready")
		return true
	}
	return false
}

func writeTempFile(t *testing.T, content []byte, filename string) (string, func()) {
	file, err := ioutil.TempFile("", filename)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.Write(content); err != nil {
		_ = os.Remove(file.Name())
		t.Fatal(err)
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(file.Name())
		t.Fatal(err)
	}

	return file.Name(), func() { _ = os.Remove(file.Name()) }
}
