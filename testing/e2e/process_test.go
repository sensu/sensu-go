package e2e

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
)

type backendProcess struct {
	APIPort            int
	AgentPort          int
	DashboardPort      int
	StateDir           string
	EtcdPeerURL        string
	EtcdClientURL      string
	EtcdInitialCluster string

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

// Start starts a backend process as configured. All exported variables in
// backendProcess must be configured.
func (b *backendProcess) Start() error {
	// path := strings.Split(os.Getenv("PATH"), filepath.ListSeparator)
	// append([]string{bin_dir}, path...)
	exe := filepath.Join(binDir, "sensu-backend")
	cmd := exec.Command(exe, "start", "-d", b.StateDir, "--api-port", strconv.FormatInt(int64(b.APIPort), 10), "--agent-port", strconv.FormatInt(int64(b.AgentPort), 10), "--store-client-url", b.EtcdClientURL, "--store-peer-url", b.EtcdPeerURL, "--store-initial-cluster", b.EtcdInitialCluster, "--dashboard-port", strconv.FormatInt(int64(b.DashboardPort), 10))
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
	BackendURL string
	AgentID    string

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

func (a *agentProcess) Start() error {
	exe := filepath.Join(binDir, "sensu-agent")
	cmd := exec.Command(exe, "start", "--backend-url", a.BackendURL, "--id", a.AgentID, "--subscriptions", "test")
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
