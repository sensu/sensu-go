package e2e

import (
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
)

type backendProcess struct {
	APIPort            int
	AgentPort          int
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
	exe := filepath.Join(binDir, "sensu-backend")
	cmd := exec.Command(exe, "start", "-d", b.StateDir, "--api-port", strconv.FormatInt(int64(b.APIPort), 10), "--agent-port", strconv.FormatInt(int64(b.AgentPort), 10), "--store-client-url", b.EtcdClientURL, "--store-peer-url", b.EtcdPeerURL, "--store-initial-cluster", b.EtcdInitialCluster)
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

	err = cmd.Start()
	if err != nil {
		return err
	}
	log.Printf("backend started with pid %d\n", cmd.Process.Pid)
	if err != nil {
		return err
	}
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

	Stdout io.Reader
	Stderr io.Reader

	cmd *exec.Cmd
}

func (a *agentProcess) Start() error {
	exe := filepath.Join(binDir, "sensu-agent")
	cmd := exec.Command(exe, "start", "-b", a.BackendURL)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	a.Stdout = out
	errp, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	a.Stderr = errp

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
