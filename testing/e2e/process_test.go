package e2e

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"syscall"
	"testing"
	"time"
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
		"--etcd-listen-client-urls", b.EtcdClientURL,
		"--etcd-listen-peer-urls", b.EtcdPeerURL,
		"--etcd-initial-cluster", b.EtcdInitialCluster,
		"--etcd-initial-cluster-state", b.EtcdInitialClusterState,
		"--etcd-name", b.EtcdName,
		"--etcd-initial-advertise-peer-urls", b.EtcdPeerURL,
		"--etcd-initial-cluster-token", b.EtcdInitialClusterToken,
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

type sensuCtl struct {
	ConfigDir string
	Namespace string
	wsURL     string
	httpURL   string
	stdin     io.Reader
}

func newCustomSensuctl(t *testing.T, wsURL, httpURL, namespace string) (*sensuCtl, func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensuctl")
	if err != nil {
		t.Fatal(err)
	}

	ctl := &sensuCtl{
		Namespace: namespace,
		ConfigDir: tmpDir,
		stdin:     os.Stdin,
		wsURL:     wsURL,
		httpURL:   httpURL,
	}

	// Authenticate sensuctl
	out, err := ctl.run("configure",
		"-n",
		"--url", httpURL,
		"--username", "admin",
		"--password", "P@ssw0rd!",
		"--format", "json",
		"--namespace", "default",
	)
	if err != nil {
		t.Fatal(err, string(out))
	}
	if namespace != "default" {
		out, err = ctl.run("namespace", "create", namespace)
		if err != nil {
			t.Fatal(err, string(out))
		}
	}

	// Switch to the configured namespace
	_, err = ctl.run("configure",
		"-n",
		"--url", httpURL,
		"--username", "admin",
		"--password", "P@ssw0rd!",
		"--format", "json",
		"--namespace", namespace,
	)
	if err != nil {
		t.Fatal(err)
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
